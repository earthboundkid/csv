# csv [![GoDoc](https://godoc.org/github.com/earthboundkid/csv?status.svg)](https://pkg.go.dev/github.com/earthboundkid/csv/v2) [![Go Report Card](https://goreportcard.com/badge/github.com/earthboundkid/csv/v2)](https://goreportcard.com/report/github.com/earthboundkid/csv/v2) [![Coverage Status](https://coveralls.io/repos/github/earthboundkid/csv/badge.svg)](https://coveralls.io/github/earthboundkid/csv)

Go CSV reader like Python's DictReader.

```
go get github.com/earthboundkid/csv/v2
```

## Example

Source CSV

```
first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
```

User type

```go
type User struct {
    Username string `csv:"username"`
    First    string `csv:"first_name"`
    Last     string `csv:"last_name"`
}
```

Scanning file

```go
var user User
for err := range csv.Scan(csv.Options{Reader: src}, &user) {
    if err != nil {
        // Do something
    }
    fmt.Println(user.Username)
}
// Output:
// rob
// ken
// gri
```
