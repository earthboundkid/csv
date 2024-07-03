package csv

import (
	"encoding/csv"
	"io"
	"iter"
)

// NULL is used to override the default separator of ',' and use 0x00 as the field separator.
const NULL = -1

// Options is a wrapper around encoding/csv.Reader that allows reference to
// columns in a CSV source by field names. Its usage is like bufio.Scanner.
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

// Rows loads the next row into the Options. It returns false upon
// encountering an error in the underlying io.Reader.
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
			idx: make(map[string]int, len(fieldnames)),
		}
		for n, field := range fieldnames {
			r.idx[field] = n
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
	idx map[string]int
	row []string
}

// Field returns the value in the currently loaded row of the column
// corresponding to fieldname.
func (r *Row) Field(fieldname string) string {
	if idx, ok := r.idx[fieldname]; ok {
		return r.row[idx]
	}
	return ""
}

// Fields returns a map from fieldnames to values for the current row.
func (r *Row) Fields() map[string]string {
	m := make(map[string]string, len(r.idx))
	for key, idx := range r.idx {
		m[key] = r.row[idx]
	}
	return m
}
