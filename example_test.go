package csv_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/carlmjohnson/csv"
)

func ExampleReader() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	r := csv.NewFieldReader(strings.NewReader(in))

	for r.Scan() {
		fmt.Println(r.Get("username"))
	}

	if err := r.Err(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// rob
	// ken
	// gri
}

// This example shows how csv.FieldReader can be configured to handle other
// types of CSV files.
func ExampleReader_options() {
	in := `first_name;last_name;username
"Rob";"Pike";rob
# lines beginning with a # character are ignored
Ken;Thompson;ken
"Robert";"Griesemer";"gri"
`
	r := csv.NewFieldReader(strings.NewReader(in))
	r.Comma = ';'
	r.Comment = '#'

	for r.Scan() {
		fmt.Println(r.Get("username"))
	}

	if err := r.Err(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// rob
	// ken
	// gri
}
