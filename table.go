package table

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

//option config parameters
var (
	//separate rows
	RowSeparator string = "\n"

	//separate columns, empty string means all the space characters
	ColumnSeparator string = ""

	//what means a empty table field
	Placeholder string = "_"

	//what to be filled in blank table field when row's too short
	BlankFilling string = ""

	//what to be filled in blank header field when row's too long
	BlankFillingForHeader string = ""

	//discard more columns or not when row's too long
	ColOverflow bool = true

	//use utf8 character to print board
	UseBoard bool = true

	//what to replace \n \b \t ...
	SpaceAlt byte = ' '

	//what to separator overflow columns
	OverFlowSeparator string = " "

	//what to fill into field in order to centralize
	CenterFilling byte = ' '

	//whether ignore empty header when all header fields are placeholder
	IgnoreEmptyHeader bool = true
)

//reset all the configs to default, if change the config, go defer it makes good
func Reset() {
	RowSeparator = "\n"
	ColumnSeparator = ""
	Placeholder = "_"
	BlankFilling = ""
	BlankFillingForHeader = ""
	ColOverflow = true
	UseBoard = true
	SpaceAlt = ' '
	OverFlowSeparator = " "
	CenterFilling = ' '
	IgnoreEmptyHeader = true
}

/*
Convert Interface for user

Description: Any type which implements this interface
	has the ability of converting to string. Especially
	to struct type, which can convert its field to
	string independently. For example:

	type Rect struct {
		Length int `table:"a, meter"`
		Width int `table:"b, meter"`
	}

	func (this Rect) Convert(field interface{}, typeStr string) (str string) {
		switch typeStr {
		case "meter":
			if v, ok := field.(int); ok {
				str = fmt.Sprintf("%dm", v)
			}
		}
		return str
	}

	The common style of table tag defined in the struct is:
	`table : [name] [,type] [,nolist]`
	1. 'name' renames the field in the table, if 'name' is '-', it means the field is totoally ignored
	2. 'type' allows user define the convertion behavior of the field
	3. 'nolist' defines whether the field appears when format object list or map

Parameters:
	field: Represents any field's value in struct
	typeStr: Type defined in struct's table tag, represents
		different type convertion.
*/
type Convertable interface {
	Convert(field interface{}, typeStr string) string
}

//raw string type, do not tokenize string's content
type RawString string

//the format API
func Format(obj interface{}) string {
	return format(encode(obj))
}

//quick print
func Print(obj interface{}) {
	fmt.Print(Format(obj))
}

//encode object, ignore panics
func encode(obj interface{}) (str string) {
	//ignore all the panic
	defer func() {
		if r := recover(); r != nil {
			str = createEmptyHeader(1) + createRow(fmt.Sprint(r))
		}
	}()

	v := reflect.ValueOf(obj)

	return encodeAny(v)
}

//encode any type
func encodeAny(v reflect.Value) (str string) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		str = encodeAny(v.Elem())
	case reflect.String:
		str = encodeString(v)
	case reflect.Array, reflect.Slice:
		str = encodeList(v)
	case reflect.Struct:
		str = encodeStruct(v)
	case reflect.Map:
		str = encodeMap(v)
	case reflect.Func:
		str = encodeFunc(v)
	default:
		_, str = encodePlain(v)
	}

	return str
}

//raw string
func encodeRawString(v reflect.Value) (str string) {
	var buf bytes.Buffer
	obj := v.Interface()

	if o, ok := obj.(RawString); ok {
		buf.WriteString(createEmptyHeader(1))
		buf.WriteString(createRow(string(o)))
	}

	return buf.String()
}

//string type, classic format type
func encodeString(v reflect.Value) (str string) {
	var buf bytes.Buffer
	if v.Kind() != reflect.String {
		return buf.String()
	}

	obj := v.Interface()

	//raw string
	if _, ok := obj.(RawString); ok {
		return encodeRawString(v)
	}

	//normal string
	if o, ok := obj.(string); ok {
		buf.WriteString(createRow(o))
	}

	return buf.String()
}

//function type, get the function name
func encodePlainFunc(v reflect.Value) (str string) {
	var buf bytes.Buffer

	if v.Kind() != reflect.Func {
		return buf.String()
	}

	buf.WriteString(createRow(runtime.FuncForPC(v.Pointer()).Name()))

	return buf.String()
}

//function type, get the function name
func encodeFunc(v reflect.Value) (str string) {
	var buf bytes.Buffer

	if v.Kind() != reflect.Func {
		return buf.String()
	}

	buf.WriteString(createEmptyHeader(1))
	buf.WriteString(encodePlainFunc(v))

	return buf.String()
}

//base types
func encodePlain(v reflect.Value) (key, str string) {
	key = Placeholder
	switch v.Kind() {
	case reflect.Invalid:

	case reflect.Ptr, reflect.Interface:
		key, str = encodePlain(v.Elem())
	case reflect.Struct:
		key, str = encodePlainStruct(v)
	case reflect.Func:
		str = encodePlainFunc(v)
	default:
		str = fmt.Sprint(v.Interface())
	}

	return key, str
}

//map type
func encodeMap(v reflect.Value) (str string) {
	var buf bytes.Buffer

	if v.Kind() != reflect.Map {
		return buf.String()
	}

	keys := v.MapKeys()
	for i, key := range keys {
		value := v.MapIndex(key)

		k1, v1 := encodePlain(key)
		k2, v2 := encodePlain(value)

		if i == 0 {
			buf.WriteString(createRow(k1, k2))
		}
		buf.WriteString(createRow(v1, v2))
	}
	return buf.String()
}

//array, slice type
func encodeList(v reflect.Value) (str string) {
	var buf bytes.Buffer

	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return buf.String()
	}

	//format list
	for i := 0; i < v.Len(); i++ {
		key, val := encodePlain(v.Index(i))

		if i == 0 {
			buf.WriteString(createRow(Placeholder, key))
		}
		buf.WriteString(createRow(strconv.Itoa(i+1), val))
	}

	return buf.String()
}

//return key string and value string
func encodePlainStruct(v reflect.Value) (string, string) {
	_, _, keys, vals := processStruct(v)

	if len(keys) == 0 {
		keys = []string{Placeholder}
		vals = []string{fmt.Sprint(v.Interface())}
	}

	return createRow(keys...), createRow(vals...)
}

//struct type
func encodeStruct(v reflect.Value) (str string) {
	var buf bytes.Buffer

	keys, vals, _, _ := processStruct(v)
	if len(keys) == 0 {
		return fmt.Sprint(v.Interface())
	}

	buf.WriteString(createEmptyHeader(2))

	for i := 0; i < len(keys); i++ {
		buf.WriteString(createRow(keys[i], vals[i]))
	}

	return buf.String()
}

//process struct, return objfmt fields and listfmt fields
func processStruct(v reflect.Value) (detKeys, detVals, absKeys, absVals []string) {
	detKeys = []string{}
	detVals = []string{}
	absKeys = []string{}
	absVals = []string{}

	obj := v.Interface()

	if v.Kind() != reflect.Struct {
		return detKeys, detVals, absKeys, absVals
	}

	//struct fields
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		//get field name and value
		name := field.Name
		value := v.FieldByName(field.Name)
		val := value.Interface()

		tag := field.Tag.Get("table")
		nameTag, typeTag, listTag := parseTag(tag)

		//name tag
		if nameTag == "-" {
			continue
		} else if nameTag != "" {
			name = nameTag
		}

		//type tag
		if o, ok := obj.(Convertable); ok && typeTag != "" {
			val = o.Convert(val, typeTag)
		}

		//list tag
		valStr := fmt.Sprintf("%v", val)
		detKeys = append(detKeys, name)
		detVals = append(detVals, valStr)
		if listTag != "nolist" {
			absKeys = append(absKeys, name)
			absVals = append(absVals, valStr)
		}
	}
	return detKeys, detVals, absKeys, absVals
}

//parse tag, process tag: `table:"-|<newName>[,<newType>][,<nolist>]"`
func parseTag(tag string) (nameTag, typeTag, listTag string) {
	//tokenize
	values := strings.Split(tag, ",")
	num := len(values)
	if num > 0 {
		nameTag = values[0]
	}
	if num > 1 {
		typeTag = values[1]
	}
	if num > 2 {
		listTag = values[2]
	}

	return nameTag, typeTag, listTag
}

//merge placehold woth col sep
func createEmptyHeader(colNum int) string {
	fields := make([]string, colNum)
	for i, _ := range fields {
		fields[i] = Placeholder
	}
	return createRow(fields...)
}

//merge fields with col sep
func createRow(fields ...string) string {
	sep := " "
	if ColumnSeparator != "" {
		sep = ColumnSeparator
	}

	var buf bytes.Buffer
	for i, field := range fields {
		field = strings.TrimSuffix(field, RowSeparator)
		if i != 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(field)
	}
	buf.WriteString(RowSeparator)

	return buf.String()
}

//table format
func format(data string) string {
	//convert string to table
	tb := preProcess(data)

	//print table
	if UseBoard {
		return boardFormat(tb)
	} else {
		return simpleFormat(tb)
	}
}

//utf8 table characters
const (
	hrLine = "─"
	vtLine = "│"

	topLeft   = "┌"
	topCenter = "┬"
	topRight  = "┐"

	middleLeft   = "├"
	middleCenter = "┼"
	middleRight  = "┤"

	bottomLeft   = "└"
	bottomCenter = "┴"
	bottomRight  = "┘"
)

//format with board
func boardFormat(tb [][]string) string {
	if len(tb) == 0 {
		tb = [][]string{{string(CenterFilling) + BlankFillingForHeader + string(CenterFilling)}}
	}
	//table attributes
	rowNum := len(tb)*2 + 1
	colNum := len(tb[0])*2 + 1
	colWidth := make([]int, colNum)
	for i, _ := range tb[0] {
		colWidth[i] = width(tb[0][i])
	}

	//init fill as --- ...
	fill := make([]string, colNum/2)
	for i, _ := range fill {
		fill[i] = strings.Repeat(hrLine, colWidth[i])
	}

	//init top ┌───┬───┐
	topLine := initLine(topLeft, topCenter, topRight, fill)

	//init middle ├───┼───┤
	middleLine := initLine(middleLeft, middleCenter, middleRight, fill)

	//init bottom └───┴───┘
	bottomLine := initLine(bottomLeft, bottomCenter, bottomRight, fill)

	//create board table
	table := make([][]string, rowNum)
	for i, _ := range table {
		switch {
		case i == 0:
			table[i] = topLine
		case i == rowNum-1:
			table[i] = bottomLine
		case i%2 == 0:
			table[i] = middleLine
		default:
			table[i] = initLine(vtLine, vtLine, vtLine, tb[i/2])
		}
	}

	//output table
	var buf bytes.Buffer
	for _, line := range table {
		for _, val := range line {
			buf.WriteString(val)
		}
		buf.WriteString("\n")
	}

	return buf.String()

}

//format without board
func simpleFormat(tb [][]string) string {
	if len(tb) == 0 {
		tb = [][]string{{string(CenterFilling) + BlankFillingForHeader + string(CenterFilling)}}
	}
	//out put table
	var buf bytes.Buffer
	for _, line := range tb {
		for _, val := range line {
			buf.WriteString(val)
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

//split str and filt empty line
func getLines(str string) []string {
	var lines []string
	if RowSeparator == "" {
		lines = strings.Fields(str)
	} else {
		lines = strings.Split(str, RowSeparator)
	}

	//filt empty string
	ret := []string{}
	for _, f := range lines {
		if len(f) > 0 {
			ret = append(ret, f)
		}
	}
	return ret
}

//split line and filt empty elements
func getFields(line string) []string {
	var fields []string
	if ColumnSeparator == "" {
		fields = strings.Fields(line)
	} else {
		fields = strings.Split(line, ColumnSeparator)
	}

	//filt empty string
	ret := []string{}
	for _, f := range fields {
		if len(f) > 0 {
			ret = append(ret, f)
		}
	}
	return ret
}

//change all the space character (\t \n _ \b) to space
func handleSpace(str string) string {
	arr := make([]rune, utf8.RuneCountInString(str))
	index := 0
	for _, c := range str {
		if unicode.IsSpace(c) && c != ' ' {
			c = rune(SpaceAlt)
		}
		arr[index] = c
		index++
	}
	return string(arr)
}

//how long is string in screen, Chinese chararter is 2 length
func width(str string) int {
	sum := 0
	for _, c := range str {
		if utf8.RuneLen(c) > 1 {
			sum += 2
		} else {
			sum++
		}
	}
	return sum
}

//convert string to 2-D slice
func preProcess(data string) [][]string {
	//get non-blank lines
	lines := []string{}
	//for _, line := range strings.Split(data, RowSeparator) {
	for _, line := range getLines(data) {
		if len(getFields(line)) != 0 {
			lines = append(lines, line)
		}
	}

	rowNum := len(lines)

	//handle empty table
	if rowNum == 0 {
		//use place holder to represent a empty table
		return [][]string{{string(CenterFilling) + BlankFillingForHeader + string(CenterFilling)}}
	}

	//get columns
	colNum := len(getFields(lines[0]))
	//max width of each column
	colWidth := make([]int, colNum)

	//process empty header
	if IgnoreEmptyHeader {
		header := getFields(lines[0])
		ignore := true
		for _, val := range header {
			if val != Placeholder {
				ignore = false
				break
			}
		}
		if ignore {
			lines = lines[1:]
			rowNum--
		}
	}

	tb := make([][]string, rowNum)
	for row, line := range lines {
		tb[row] = make([]string, colNum)

		//fillings
		filling := BlankFilling
		if row == 0 {
			filling = BlankFillingForHeader
		}

		//init row as blank filling
		for index, _ := range tb[row] {
			tb[row][index] = filling
		}

		//get fields
		fields := getFields(line)
		for col, val := range fields {
			//handle placeholder
			if val == Placeholder {
				val = filling
			}

			//handle column overflow
			if col >= colNum {
				if ColOverflow {
					col = colNum - 1
					val = tb[row][col] + OverFlowSeparator + val
				} else {
					//discard more cols
					break
				}
			}
			tb[row][col] = handleSpace(val)
		}
	}

	//calcu max width, extend colwidth + 2 to store blank
	for col := 0; col < colNum; col++ {
		for row := 0; row < rowNum; row++ {
			val := tb[row][col]
			size := width(val)
			if size > colWidth[col] {
				colWidth[col] = size
			}
		}
		colWidth[col] += 2
	}

	//middle value with blank
	cfill := string(CenterFilling)
	for row, line := range tb {
		for col, val := range line {
			size := width(val)
			left := (colWidth[col] - size) / 2
			right := colWidth[col] - size - left
			tb[row][col] = strings.Repeat(cfill, left) + val + strings.Repeat(cfill, right)
		}
	}

	return tb

}

//form table line
func initLine(left, center, right string, fill []string) []string {
	colNum := len(fill)*2 + 1
	line := make([]string, colNum)
	for i, _ := range line {
		tmp := ""
		switch {
		case i == 0:
			tmp = left
		case i == colNum-1:
			tmp = right
		case i%2 == 0:
			tmp = center
		default:
			tmp = fill[i/2]
		}
		line[i] = tmp
	}
	return line
}
