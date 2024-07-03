package csv_test

import (
	"fmt"
	"log"
	"strings"

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
