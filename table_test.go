package table

import (
	"fmt"
	"testing"
	"time"
)

//object format definition
type Obj struct {
	Key     string `table:"Name"`
	Value   int64  `table:"Time,time"`
	Options string `table:"-"`
	Default []int  `table:",,nolist"`
}

//user-define type convertion
func (this Obj) Convert(field interface{}, typeStr string) (str string) {
	switch typeStr {
	case "time":
		if val, ok := field.(int64); ok {
			str = time.Unix(val/1e3, val%1e3*1e6).Format("2006-01-02 15:04:05")
		}
	}
	return str
}

//test cases
func TestFormat(t *testing.T) {
	//classic string
	str := ` ID _ Num Digit
	1 2 3你好
	4 _ 5 
	7 8 9 10 11`

	//map
	m := map[string]string{"key1xxxxxx": "value1", "key2": "value2"}

	//struct
	o := Obj{Key: "NOW", Value: time.Now().UnixNano() / 1e6, Default: []int{1, 2, 3}}

	//array
	list := [4]Obj{o, o, o, o}

	//raw string and string
	raw := RawString("I am a raw string")
	normal := "I am a normal string"

	//table format
	fmt.Print(Format(str))
	fmt.Print(Format(m))
	fmt.Print(Format(o))
	ColumnSeparator = "\v"
	fmt.Print(Format(list))
	Reset()
	fmt.Print(Format(raw))
	fmt.Print(Format(normal))
	fmt.Print(Format(nil))
	fmt.Print(Format(map[string]interface{}{"1": nil}))
	fmt.Print(Format(1))
	fmt.Print(Format(0.32))
	fmt.Print(Format(1 + 2i))
	fmt.Print(Format(Format))
}

//complicated cases
func TestListObj(t *testing.T) {
	ColumnSeparator = "*"
	RowSeparator = "//"
	defer Reset()

	//struct ptr
	o := &Obj{Key: "NOW", Value: time.Now().UnixNano() / 1e6, Default: []int{1, 2, 3}}

	//obj ptr array
	list := [4]*Obj{o, o, o, o}
	fmt.Println(Format(list))

	//int array
	fmt.Println(Format([]int{1, 2, 3, 4}))

	//int map
	fmt.Println(Format(map[int]int{1: 4, 2: 3, 3: 2, 4: 1}))

	//string object map
	fmt.Println(Format(map[string]*Obj{"12": o, "23": o}))

	//obj ptr map
	type My struct {
		Name string
		Age  int
	}
	m := &My{"jiushi", 1}
	n := &My{"hello", 22}
	fmt.Println(Format(map[*My]*Obj{m: o, n: o}))
}
