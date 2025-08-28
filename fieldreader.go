// Package csv is a wrapper around encoding/csv
// that makes it more convenient to work with named fields of CSV tables.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"reflect"
	"strconv"
)

// NULL is used to override the default separator of ',' and use 0x00 as the field separator.
const NULL = -1

// Options is a wrapper around encoding/csv.Reader
// that allows look up of columns in a CSV source by field name.
type Options struct {
	// Reader must be set
	Reader io.Reader

	// Comma is the field delimiter.
	// It is set to comma (',') by default.
	// To use 0x00 as the field separator, set it to -1
	Comma rune
	// Comment, if not 0, is the comment character. Lines beginning with the
	// Comment character without preceding whitespace are ignored.
	// With leading whitespace the Comment character becomes part of the
	// field, even if TrimLeadingSpace is true.
	Comment rune
	// If LazyQuotes is true, a quote may appear in an unquoted field and a
	// non-doubled quote may appear in a quoted field.
	LazyQuotes bool
	// If TrimLeadingSpace is true, leading white space in a field is ignored.
	// This is done even if the field delimiter, Comma, is white space.
	TrimLeadingSpace bool
	// FieldNames are the names for the fields on each row. If FieldNames is
	// left nil, it will be set to the first row read.
	FieldNames []string
}

// Rows returns a sequence yielding a Row for each row parsed from o.Reader.
// If o.Reader returns an error other than io.EOF, it will be yielded to the caller.
func (o *Options) Rows() iter.Seq2[*Row, error] {
	return func(yield func(*Row, error) bool) {
		cr := csv.NewReader(o.Reader)
		cr.ReuseRecord = true
		if o.Comma == NULL {
			cr.Comma = 0x00
		} else if o.Comma != 0 {
			cr.Comma = o.Comma
		}
		cr.Comment = o.Comment
		cr.LazyQuotes = o.LazyQuotes
		cr.TrimLeadingSpace = o.TrimLeadingSpace

		fieldnames := o.FieldNames
		if o.FieldNames == nil {
			row, err := cr.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(nil, err)
				return
			}
			fieldnames = row
		}

		r := Row{
			keyToColIdx: make(map[string]int, len(fieldnames)),
		}
		for n, field := range fieldnames {
			r.keyToColIdx[field] = n
		}

		var (
			row []string
			err error
		)
		for {
			row, err = cr.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(nil, err)
				return
			}
			r.row = row
			if !yield(&r, nil) {
				return
			}
		}
	}
}

// ReadAll consumes o.Reader and returns a slice of maps for each row.
func (o *Options) ReadAll() ([]map[string]string, error) {
	var rows []map[string]string
	for row, err := range o.Rows() {
		if err != nil {
			return nil, err
		}
		rows = append(rows, row.Fields())
	}
	return rows, nil
}

// Row represents one scanned row of a CSV file.
// It is only valid during the current iteration.
type Row struct {
	keyToColIdx map[string]int
	row         []string
}

// Field returns the value in the currently loaded row of the column
// corresponding to fieldname.
func (r *Row) Field(fieldname string) string {
	if colIdx, ok := r.keyToColIdx[fieldname]; ok {
		return r.row[colIdx]
	}
	return ""
}

// Fields returns a map from fieldnames to values for the current row.
func (r *Row) Fields() map[string]string {
	m := make(map[string]string, len(r.keyToColIdx))
	for key, colIdx := range r.keyToColIdx {
		m[key] = r.row[colIdx]
	}
	return m
}

// Scan returns an iterator reading from o.
// On each iteration it scans the row into v.
// See [Row.Scan].
func Scan[T any](o Options, v *T) iter.Seq[error] {
	return func(yield func(error) bool) {
		var (
			s        reflect.Value
			fieldIdx []structFieldToCSVColumn
		)
		for row, err := range o.Rows() {
			if fieldIdx == nil {
				s, fieldIdx = row.mapFields(v)
			}
			if err != nil {
				yield(err)
				return
			}
			if !yield(row.scan(s, fieldIdx)) {
				return
			}
		}
	}
}

// ScanAll returns a slice of all objects read from o or an error.
// See [Row.Scan].
func ScanAll[T any](o Options) ([]T, error) {
	var s []T
	var v T
	for err := range Scan(o, &v) {
		if err != nil {
			return nil, err
		}
		s = append(s, v)
	}
	return s, nil
}

// Scan reflects on the row and sets the appropriate fields of s.
// If v is not a pointer to a struct, Scan will panic.
// The struct fields to be scanned into must be exported, of type string,
// and have a csv field tag with the name of the field to copy.
func (r *Row) Scan(v any) error {
	return r.scan(r.mapFields(v))
}

type structFieldToCSVColumn struct {
	fieldIndex  []int
	columnIndex int
	reflect.Kind
}

func (r *Row) mapFields(v any) (reflect.Value, []structFieldToCSVColumn) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		panic("must scan into pointer to struct")
	}
	s := rv.Elem()
	if s.Kind() != reflect.Struct {
		panic("must scan into pointer to struct")
	}
	fields := reflect.VisibleFields(s.Type())
	fieldIdx := make([]structFieldToCSVColumn, 0, len(fields))
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		kind := field.Type.Kind()
		switch kind {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
		default:
			continue
		}
		key := field.Tag.Get("csv")
		if key == "" {
			continue
		}
		if keyIdx, ok := r.keyToColIdx[key]; ok {
			fieldIdx = append(fieldIdx, structFieldToCSVColumn{
				fieldIndex:  field.Index,
				columnIndex: keyIdx,
				Kind:        kind,
			})
		}
	}
	return s, fieldIdx
}

func (r *Row) scan(s reflect.Value, fieldColumns []structFieldToCSVColumn) error {
	for _, fieldColumn := range fieldColumns {
		switch fieldColumn.Kind {
		case reflect.String:
			s.FieldByIndex(fieldColumn.fieldIndex).SetString(r.row[fieldColumn.columnIndex])
		case reflect.Int:
			if n, err := strconv.ParseInt(r.row[fieldColumn.columnIndex], 0, 0); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetInt(n)
			}
		case reflect.Int8:
			if n, err := strconv.ParseInt(r.row[fieldColumn.columnIndex], 0, 8); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetInt(n)
			}
		case reflect.Int16:
			if n, err := strconv.ParseInt(r.row[fieldColumn.columnIndex], 0, 16); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetInt(n)
			}
		case reflect.Int32:
			if n, err := strconv.ParseInt(r.row[fieldColumn.columnIndex], 0, 32); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetInt(n)
			}
		case reflect.Int64:
			if n, err := strconv.ParseInt(r.row[fieldColumn.columnIndex], 0, 64); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetInt(n)
			}
		case reflect.Uint:
			if n, err := strconv.ParseUint(r.row[fieldColumn.columnIndex], 0, 0); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetUint(n)
			}
		case reflect.Uint8:
			if n, err := strconv.ParseUint(r.row[fieldColumn.columnIndex], 0, 8); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetUint(n)
			}
		case reflect.Uint16:
			if n, err := strconv.ParseUint(r.row[fieldColumn.columnIndex], 0, 16); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetUint(n)
			}
		case reflect.Uint32:
			if n, err := strconv.ParseUint(r.row[fieldColumn.columnIndex], 0, 32); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetUint(n)
			}
		case reflect.Uint64:
			if n, err := strconv.ParseUint(r.row[fieldColumn.columnIndex], 0, 64); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetUint(n)
			}
		case reflect.Float32:
			if n, err := strconv.ParseFloat(r.row[fieldColumn.columnIndex], 32); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetFloat(n)
			}
		case reflect.Float64:
			if n, err := strconv.ParseFloat(r.row[fieldColumn.columnIndex], 64); err != nil {
				return fieldErr(s, fieldColumn, err)
			} else {
				s.FieldByIndex(fieldColumn.fieldIndex).SetFloat(n)
			}
		}
	}
	return nil
}

func fieldErr(s reflect.Value, idx structFieldToCSVColumn, err error) error {
	name := s.Type().FieldByIndex(idx.fieldIndex).Name
	return fmt.Errorf("setting %s: %w", name, err)
}
