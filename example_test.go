package csv_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/earthboundkid/csv/v2"
)

func Example() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	csvopt := csv.Options{
		Reader: strings.NewReader(in),
	}

	for row, err := range csvopt.Rows() {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(row.Field("username"))
	}

	// Output:
	// rob
	// ken
	// gri
}

// This example shows how csv.FieldReader can be configured to handle other
// types of CSV files.
func ExampleOptions() {
	in := `"Rob";"Pike";rob
# lines beginning with a # character are ignored
Ken;Thompson;ken
"Robert";"Griesemer";"gri"
`
	csvopt := csv.Options{
		Reader:     strings.NewReader(in),
		Comma:      ';',
		Comment:    '#',
		FieldNames: []string{"first_name", "last_name", "username"},
	}

	rows, err := csvopt.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rows)

	// Output:
	// [map[first_name:Rob last_name:Pike username:rob] map[first_name:Ken last_name:Thompson username:ken] map[first_name:Robert last_name:Griesemer username:gri]]
}

func ExampleScan() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	csvopt := csv.Options{
		Reader: strings.NewReader(in),
	}

	var user struct {
		Username string `csv:"username"`
		First    string `csv:"first_name"`
		Last     string `csv:"last_name"`
	}
	for err := range csv.Scan(csvopt, &user) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(user)
	}

	// Output:
	// {rob Rob Pike}
	// {ken Ken Thompson}
	// {gri Robert Griesemer}
}

func BenchmarkRows(b *testing.B) {
	var buf strings.Builder
	buf.WriteString("first_name,last_name,username\n")
	for range 10_000 {
		buf.WriteString(`"Rob","Pike",rob` + "\n")
		buf.WriteString(`Ken,Thompson,ken` + "\n")
		buf.WriteString(`"Robert","Griesemer","gri"` + "\n")
	}
	in := buf.String()
	b.ResetTimer()

	for range b.N {
		csvopt := csv.Options{
			Reader: strings.NewReader(in),
		}
		for _, err := range csvopt.Rows() {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkScan(b *testing.B) {
	var buf strings.Builder
	buf.WriteString("first_name,last_name,username\n")
	for range 10_000 {
		buf.WriteString(`"Rob","Pike",rob` + "\n")
		buf.WriteString(`Ken,Thompson,ken` + "\n")
		buf.WriteString(`"Robert","Griesemer","gri"` + "\n")
	}
	in := buf.String()
	b.ResetTimer()

	for range b.N {
		csvopt := csv.Options{
			Reader: strings.NewReader(in),
		}
		var user struct {
			Username string `csv:"username"`
			First    string `csv:"first_name"`
			Last     string `csv:"last_name"`
		}
		for err := range csv.Scan(csvopt, &user) {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
