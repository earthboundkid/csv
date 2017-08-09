package csv

import (
	"encoding/csv"
	"io"
)

type FieldReader struct {
	// Comma is the field delimiter.
	// It is set to comma (',') by NewReader.
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
	// left nil, it will be  set to the first row read.
	FieldNames []string

	r   csv.Reader
	idx map[string]int
	row []string
	err error
}

func NewFieldReader(r io.Reader) *FieldReader {
	cr := csv.NewReader(r)
	return &FieldReader{
		Comma: ',',
		r:     *cr,
	}
}

func (f *FieldReader) Scan() bool {
	if f.err != nil {
		return false
	}

	if f.idx == nil {
		f.r.Comma = f.Comma
		f.r.Comment = f.Comment
		f.r.LazyQuotes = f.LazyQuotes
		f.r.TrimLeadingSpace = f.TrimLeadingSpace
	}

	if f.FieldNames == nil {
		f.FieldNames, f.err = f.r.Read()
		if f.err != nil {
			return false
		}
	}

	if f.idx == nil {
		f.idx = make(map[string]int, len(f.row))
		for n, field := range f.FieldNames {
			f.idx[field] = n
		}
	}

	f.row, f.err = f.r.Read()
	if f.err != nil {
		return false
	}
	return true
}

func (f *FieldReader) Field(fieldname string) string {
	if idx, ok := f.idx[fieldname]; ok {
		return f.row[idx]
	}
	return ""
}

func (f *FieldReader) Fields() map[string]string {
	m := make(map[string]string, len(f.idx))
	for key, idx := range f.idx {
		m[key] = f.row[idx]
	}
	return m
}

func (f *FieldReader) ReadAll() ([]map[string]string, error) {
	rows := make([]map[string]string, 0)
	for f.Scan() {
		rows = append(rows, f.Fields())
	}
	return rows, f.Err()
}

func (f *FieldReader) Err() error {
	if f.err != io.EOF {
		return f.err
	}
	return nil
}
