package rapidjson

// #cgo CFLAGS: -I. -fpic
// #include <stdlib.h>
// #include <stdint.h>
// #include "rjwrapper.h"
import "C"
import "unsafe"

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrJsonParse    = errors.New("JSON parsing error")
	ErrNotArray     = errors.New("Not an array")
	ErrNotObject    = errors.New("Not an object")
	ErrPathNotFound = errors.New("Path not found")
	ErrNotInt       = errors.New("Not an int")
	ErrNotFloat     = errors.New("Not a float")
	ErrNotBool      = errors.New("Not a bool")
	ErrNotString    = errors.New("Not a string")
	ErrBadType      = errors.New("Bad type")
	ErrMemberExists = errors.New("Member already exists")
	ErrOutOfBounds  = errors.New("Array index out of bounds")

	parseErrors = []string{
		"No error",
		"The document is empty",
		"The document root must not follow by other values",
		"Invalid value",
		"Missing a name for object member",
		"Missing a colon after a name of object member",
		"Missing a comma or '}' after an object member",
		"Missing a comma or ']' after an array element",
		"Incorrect hex digit after \\u escape in string",
		"The surrogate pair in string is invalid",
		"Invalid escape character in string",
		"Missing a closing quotation mark in string",
		"Invalid encoding in string",
		"Number too big to be stored in double",
		"Miss fraction part in number",
		"Miss exponent in number",
		"Parsing was terminated",
		"Unspecific syntax error",
	}
)

const (
	TypeNull   int = 0
	TypeFalse  int = 1
	TypeTrue   int = 2
	TypeObject int = 3
	TypeArray  int = 4
	TypeString int = 5
	TypeNumber int = 6
)

type RJCommon interface {
	Free()
}

type Doc struct {
	json      C.JsonDoc
	allocated []RJCommon
}

type Container struct {
	doc *Doc
	ct  C.JsonVal
}

// bool helpers
func CBoolTest(result C.int) bool {
	if result == 0 {
		return false
	} else {
		return true
	}
}
func BoolToC(b bool) C.int {
	if b {
		return 1
	} else {
		return 0
	}
}

// initialization
func NewDoc() *Doc {
	var json Doc
	json.json = C.JsonInit()
	return &json
}
func (json *Doc) Free() {
	if json == nil {
		return
	}
	for _, ct := range json.allocated {
		ct.Free()
	}
	C.JsonFree(unsafe.Pointer(json.json))
}
func (json *Doc) NewContainer() *Container {
	var ct Container
	ct.doc = json
	ct.ct = C.ValInit()
	json.allocated = append([]RJCommon{&ct}, json.allocated...)
	return &ct
}
func (json *Doc) NewContainerObj() *Container {
	ct := json.NewContainer()
	ct.InitObj()
	return ct
}
func (json *Doc) NewContainerArray() *Container {
	ct := json.NewContainer()
	ct.InitArray()
	return ct
}
func (ct *Container) Free() {
	if ct == nil {
		return
	}
	if ct != nil {
		C.ValFree(unsafe.Pointer(ct.ct))
	}
}
func (json *Doc) GetContainer() *Container {
	var ct Container
	ct.ct = C.JsonVal(unsafe.Pointer(json.json))
	ct.doc = json
	return &ct
}
func (json *Doc) GetContainerNewObj() *Container {
	var ct Container
	ct.ct = C.JsonVal(unsafe.Pointer(json.json))
	ct.doc = json
	ct.InitObj()
	return &ct
}
func (ct *Container) GetCopy() *Container {
	if ct == nil {
		return nil
	}
	copyDoc := NewDoc()
	ctCopy := copyDoc.GetContainer()
	ctCopy.SetContainerCopy(ct)
	return ctCopy
}

func (json *Doc) GetAllocated() int {
	return len(json.allocated)
}

// parse
func (json *Doc) Parse(input []byte) error {
	return json.ParseString(string(input))
}
func (json *Doc) ParseString(input string) error {
	cStr := C.CString(input)
	defer C.free(unsafe.Pointer(cStr))
	C.JsonParse(unsafe.Pointer(json.json), cStr)

	if json.HasParseError() {
		return ErrJsonParse
	} else {
		return nil
	}
}
func (json *Doc) GetParseError() string {
	return parseErrors[int(C.GetParseErrorCode(unsafe.Pointer(json.json)))]
}
func NewParsedJson(input []byte) (*Doc, error) {
	doc := NewDoc()
	err := doc.Parse(input)
	return doc, err
}
func NewParsedStringJson(input string) (*Doc, error) {
	doc := NewDoc()
	err := doc.ParseString(input)
	return doc, err
}
func (json *Doc) HasParseError() bool {
	return CBoolTest(C.HasParseError(unsafe.Pointer(json.json)))
}

// get string/bytes output
func (json *Doc) String() string {
	cStr := C.GetString(unsafe.Pointer(json.json))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (json *Doc) Pretty() string {
	cStr := C.GetPrettyString(unsafe.Pointer(json.json))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (json *Doc) Bytes() []byte {
	return []byte(json.String())
}

// various getters
func (ct *Container) HasMember(key string) bool {
	if ct == nil {
		return false
	} else if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		return CBoolTest(C.HasMember(unsafe.Pointer(ct.ct), cStr))
	} else {
		return false
	}
}
func (ct *Container) GetMemberCount() (int, error) {
	if ct == nil {
		return 0, ErrNotObject
	} else if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return int(C.GetMemberCount(unsafe.Pointer(ct.ct))), nil
	} else {
		return 0, ErrNotObject
	}
}
func (ct *Container) GetMemberName(index int) string {
	if ct == nil {
		return ""
	}
	cStr := C.GetMemberName(unsafe.Pointer(ct.ct), C.int(index))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) GetMemberNames() ([]string, error) {
	if ct == nil {
		return make([]string, 0), ErrPathNotFound
	}
	count, err := ct.GetMemberCount()
	result := make([]string, count)
	if err != nil {
		return result, err
	}
	for i := 0; i < count; i++ {
		result[i] = ct.GetMemberName(i)
	}
	return result, nil
}
func (ct *Container) GetMemberMap() (map[string]*Container, error) {
	if ct == nil {
		return make(map[string]*Container, 0), ErrPathNotFound
	}
	count, err := ct.GetMemberCount()
	result := make(map[string]*Container, count)
	if err != nil {
		return result, err
	}
	members, _ := ct.GetMemberNames()
	for _, m := range members {
		result[m], _ = ct.GetMember(m)
	}
	return result, nil
}

func (ct *Container) GetMember(key string) (*Container, error) {
	if ct == nil {
		return nil, ErrPathNotFound
	}
	cStr := C.CString(key)
	defer C.free(unsafe.Pointer(cStr))
	if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		if ct.HasMember(key) {
			var m Container
			m.doc = ct.doc
			m.ct = C.GetMember(unsafe.Pointer(ct.ct), cStr)
			return &m, nil
		} else {
			return nil, ErrPathNotFound
		}
	} else {
		return nil, ErrNotObject
	}
}
func (ct *Container) String() string {
	if ct == nil {
		return ""
	}
	cStr := C.ValGetString(unsafe.Pointer(ct.ct))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) Pretty() string {
	if ct == nil {
		return ""
	}
	cStr := C.ValGetPrettyString(unsafe.Pointer(ct.ct))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) Bytes() []byte {
	if ct == nil {
		return []byte("")
	}
	return []byte(ct.String())
}

func (ct *Container) GetPathContainer(path string) (*Container, error) {
	if ct == nil {
		return nil, ErrNotObject
	}
	keys := strings.Split(path, ".")
	next := ct
	var err error
	for _, key := range keys {
		next, err = next.GetMember(key)
		if err != nil {
			return next, err
		}
	}

	return next, nil
}
func (ct *Container) PathExists(path string) bool {
	// don't use this, just call GetPathContainer instead!
	if ct == nil {
		return false
	}
	_, err := ct.GetPathContainer(path)
	if err != nil {
		return false
	} else {
		return true
	}
}
func (ct *Container) GetPathNewContainer(path string) (*Container, error) {
	if ct == nil {
		return nil, ErrNotObject
	}
	keys := strings.Split(path, ".")
	next := ct
	prev := ct
	var err error
	var isNewPath bool
	isNewPath = false
	var addKeys []string
	for _, key := range keys {
		if isNewPath {
			addKeys = append(addKeys, key)
		} else {
			next, err = next.GetMember(key)
			if err == nil {
				prev = next
			} else if err == ErrPathNotFound {
				addKeys = append(addKeys, key)
				isNewPath = true
			} else if err == ErrNotObject {
				return next, err
			}
		}
	}

	for _, key := range addKeys {
		add := ct.doc.NewContainerObj()
		err = prev.AddMember(key, add)
		if err != nil {
			return prev, err
		}
		prev, err = prev.GetMember(key)

	}

	return prev, nil
}
func (ct *Container) IsEqual(other *Container) bool {
	if ct == nil || other == nil {
		return ct == other
	} else {
		res := C.IsValEqual(unsafe.Pointer(ct.ct), unsafe.Pointer(other.ct))
		return CBoolTest(res)
	}
}

// typed getters
func (ct *Container) GetType() int {
	if ct == nil {
		return TypeNull
	} else {
		return int(C.GetType(unsafe.Pointer(ct.ct)))
	}
}
func (ct *Container) GetInt() (int, error) {
	if ct == nil {
		var result int
		return result, ErrPathNotFound
	} else if CBoolTest(C.IsInt(unsafe.Pointer(ct.ct))) {
		result := int(C.ValGetInt(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result int
		return result, ErrNotInt
	}
}
func (ct *Container) GetInt64() (int64, error) {
	if ct == nil {
		var result int64
		return result, ErrPathNotFound
	} else if CBoolTest(C.IsInt64(unsafe.Pointer(ct.ct))) {
		result := int64(C.ValGetInt64(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result int64
		return result, ErrNotInt
	}
}
func (ct *Container) GetFloat() (float64, error) {
	if ct == nil {
		var result float64
		return result, ErrPathNotFound
	} else if CBoolTest(C.IsDouble(unsafe.Pointer(ct.ct))) {
		result := float64(C.ValGetDouble(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result float64
		return result, ErrNotFloat
	}
}
func (ct *Container) GetBool() (bool, error) {
	if ct == nil {
		var result bool
		return result, ErrPathNotFound
	} else if CBoolTest(C.IsBool(unsafe.Pointer(ct.ct))) {
		result := CBoolTest(C.ValGetBool(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result bool
		return result, ErrNotBool
	}
}
func (ct *Container) GetString() (string, error) {
	if ct == nil {
		var result string
		return result, ErrPathNotFound
	} else if CBoolTest(C.IsString(unsafe.Pointer(ct.ct))) {
		cStr := C.ValGetBasicString(unsafe.Pointer(ct.ct))
		defer C.free(unsafe.Pointer(cStr))
		str := C.GoString(cStr)
		return str, nil
	} else {
		var result string
		return result, ErrNotString
	}
}
func (ct *Container) GetValue() (interface{}, error) {
	switch ct.GetType() {
	case TypeString:
		return ct.GetString()
	case TypeTrue, TypeFalse:
		return ct.GetBool()
	case TypeNumber:
		if r, err := ct.GetInt64(); err == ErrNotInt {
			return ct.GetFloat()
		} else {
			return r, err
		}
	case TypeArray, TypeObject:
		return nil, ErrBadType
	default:
		return nil, nil
	}
}
func (ct *Container) GetArraySize() (int, error) {
	if ct == nil {
		return 0, ErrPathNotFound
	} else if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		size := int(C.ValArraySize(unsafe.Pointer(ct.ct)))
		return size, nil
	} else {
		return 0, ErrNotArray
	}
}
func (ct *Container) GetArrayValue(index int) *Container {
	if ct == nil {
		return nil
	}
	var a Container
	a.doc = ct.doc
	a.ct = C.GetArrayValueAt(unsafe.Pointer(ct.ct), C.int(index))
	return &a
}

func (ct *Container) GetIntArray() ([]int64, error) {
	if ct == nil {
		return make([]int64, 0), ErrPathNotFound
	}
	count, err := ct.GetArraySize()
	result := make([]int64, count)
	if err != nil {
		return result, err
	}

	for i := 0; i < count; i++ {
		item := ct.GetArrayValue(i)
		result[i], err = item.GetInt64()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
func (ct *Container) GetStringArray() ([]string, error) {
	if ct == nil {
		return make([]string, 0), ErrPathNotFound
	}
	count, err := ct.GetArraySize()
	result := make([]string, count)
	if err != nil {
		return result, err
	}

	for i := 0; i < count; i++ {
		item := ct.GetArrayValue(i)
		result[i], err = item.GetString()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
func (ct *Container) GetArray() ([]*Container, int, error) {
	count, err := ct.GetArraySize()
	result := make([]*Container, count)
	if err != nil {
		return result, TypeNull, err
	}

	for i := 0; i < count; i++ {
		result[i] = ct.GetArrayValue(i)
	}
	if count == 0 {
		return result, TypeNull, nil
	}
	return result, result[0].GetType(), nil
}

// setters
func (ct *Container) SetValue(v interface{}) error {
	if ct == nil {
		return ErrPathNotFound
	}
	if v == nil {
		C.SetNull(unsafe.Pointer(ct.ct))
		return nil
	}

	switch v.(type) {
	case int64:
		C.SetInt64(unsafe.Pointer(ct.ct), C.int64_t(v.(int64)))
		return nil
	case int32:
		C.SetInt64(unsafe.Pointer(ct.ct), C.int64_t(v.(int32)))
		return nil
	case int:
		C.SetInt64(unsafe.Pointer(ct.ct), C.int64_t(v.(int)))
		return nil
	case int16:
		C.SetInt(unsafe.Pointer(ct.ct), C.int(v.(int16)))
		return nil
	case int8:
		C.SetInt(unsafe.Pointer(ct.ct), C.int(v.(int8)))
		return nil
	case float64:
		C.SetDouble(unsafe.Pointer(ct.ct), C.double(v.(float64)))
		return nil
	case bool:
		C.SetBool(unsafe.Pointer(ct.ct), BoolToC(v.(bool)))
		return nil
	case string:
		cStr := C.CString(v.(string))
		defer C.free(unsafe.Pointer(cStr))
		C.SetString(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), cStr)
		return nil
	default:
		return ErrBadType
	}
}
func (ct *Container) SetContainer(item *Container) {
	if ct == nil {
		return
	}
	C.SetValue(unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
}
func (ct *Container) SetContainerCopy(item *Container) {
	if ct == nil {
		return
	}
	C.CopyFrom(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
}
func (ct *Container) InitObj() {
	if ct == nil {
		return
	}
	ct.ct = C.InitObj(unsafe.Pointer(ct.ct))
}
func (ct *Container) AddValue(key string, v interface{}) error {
	if ct == nil {
		return ErrPathNotFound
	}
	item := ct.doc.NewContainer()
	err := item.SetValue(v)
	if err != nil {
		return err
	}

	return ct.AddMember(key, item)
}
func (ct *Container) AddMember(key string, item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		if CBoolTest(C.HasMember(unsafe.Pointer(ct.ct), cStr)) {
			return ErrMemberExists
		} else {
			C.AddStrMember(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), cStr, unsafe.Pointer(item.ct))
			return nil
		}
	}
}
func (ct *Container) AddMemberCopy(key string, item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return ErrNotObject
	} else {
		ct.SetMemberValue(key, nil)
		target, err := ct.GetMember(key)
		if err != nil {
			return err
		}
		target.SetContainerCopy(item)
		return nil
	}
}
func (ct *Container) AddMemberArray(key string, items []*Container) error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		if CBoolTest(C.HasMember(unsafe.Pointer(ct.ct), cStr)) {
			return ErrMemberExists
		} else {
			array := ct.doc.NewContainerArray()
			for _, item := range items {
				array.ArrayAppendContainer(item)
			}
			C.AddStrMember(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), cStr, unsafe.Pointer(array.ct))
			return nil
		}

	}
}
func (ct *Container) SetMember(key string, item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	}
	target, err := ct.GetMember(key)
	if err == nil {
		target.SetContainer(item)
	} else if err == ErrNotObject {
		return err
	} else if err == ErrPathNotFound {
		ct.AddMember(key, item)
	}
	return nil
}
func (ct *Container) SetMemberValue(key string, v interface{}) error {
	if ct == nil {
		return ErrPathNotFound
	}
	item := ct.doc.NewContainer()
	err := item.SetValue(v)
	if err != nil {
		return err
	}

	return ct.SetMember(key, item)
}
func (ct *Container) SetMemberCopy(key string, item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	}
	target, err := ct.GetMember(key)
	if err == nil {
		target.SetContainerCopy(item)
	} else if err == ErrNotObject {
		return err
	} else if err == ErrPathNotFound {
		ct.AddMemberCopy(key, item)
	}
	return nil
}
func (ct *Container) AddMemberAtPath(path string, item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	}
	dest, err := ct.GetPathNewContainer(path)
	if err != nil {
		return err
	}
	dest.SetContainer(item)
	return nil
}
func (ct *Container) AddValueAtPath(path string, v interface{}) error {
	if ct == nil {
		return ErrPathNotFound
	}
	dest, err := ct.GetPathNewContainer(path)
	if err != nil {
		return err
	}
	item := ct.doc.NewContainer()
	item.SetValue(v)
	dest.SetContainer(item)
	return nil
}

func (ct *Container) InitArray() {
	if ct == nil {
		return
	}
	C.InitArray(unsafe.Pointer(ct.ct))
}
func (ct *Container) ArrayAppendContainer(item *Container) error {
	if ct == nil {
		return ErrPathNotFound
	} else if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		C.ArrayAppend(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
		return nil
	} else {
		return ErrNotArray
	}
}
func (ct *Container) ArrayAppendCopy(item *Container) error {
	if ct == nil || item == nil {
		return ErrPathNotFound
	} else if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		newCt := ct.doc.NewContainer()
		newCt.SetContainerCopy(item)
		C.ArrayAppend(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), unsafe.Pointer(newCt.ct))
		return nil
	} else {
		return ErrNotArray
	}
}
func (ct *Container) ArrayAppend(v interface{}) error {
	if ct == nil {
		return ErrPathNotFound
	}
	item := ct.doc.NewContainer()
	item.SetValue(v)
	return ct.ArrayAppendContainer(item)
}
func (ct *Container) SwapContainer(item *Container) {
	C.Swap(unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
}

// deleters
func (ct *Container) RemoveMember(key string) error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		C.RemoveMember(unsafe.Pointer(ct.ct), cStr)
	}
	return nil
}
func (ct *Container) ArrayClear() error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		return ErrNotArray
	} else {
		C.ArrayClear(unsafe.Pointer(ct.ct))
	}
	return nil
}
func (ct *Container) ArrayRemove(index int) error {
	if ct == nil {
		return ErrPathNotFound
	} else if !CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		return ErrNotArray
	} else if int(C.ValArraySize(unsafe.Pointer(ct.ct))) <= index {
		return ErrOutOfBounds
	} else {
		C.ArrayRemove(unsafe.Pointer(ct.ct), C.int(index))
	}

	return nil
}
func (ct *Container) RemoveMemberAtPath(path string) error {
	if ct == nil {
		return ErrPathNotFound
	}
	parts := strings.Split(path, ".")
	if len(parts) >= 1 {
		switch ct.GetType() {
		case TypeObject:
			if len(parts) > 1 {
				next := ct.GetMemberOrNil(parts[0])
				if err := next.RemoveMemberAtPath(strings.Join(parts[1:], ".")); err != nil {
					return err
				}
			} else {
				if err := ct.RemoveMember(parts[0]); err != nil {
					return err
				}
			}
		case TypeArray:
			array := ct.GetArrayOrNil()
			for _, c := range array {
				if err := c.RemoveMemberAtPath(path); err != nil {
					return err
				}
			}
		default:
			return ErrBadType
		}
	} else {
		return ErrBadType
	}
	return nil
}
func (ct *Container) StripNulls(leaveEmptyArray bool) *Container {
	switch ct.GetType() {
	case TypeArray:
		arr, _, _ := ct.GetArray()
		var removes []int
		for i, a := range arr {
			if filtered := a.StripNulls(leaveEmptyArray); filtered == nil {
				removes = append(removes, i)
			}
		}
		if len(removes) < len(arr) || leaveEmptyArray {
			sort.Sort(sort.Reverse(sort.IntSlice(removes)))
			for _, i := range removes {
				C.ArrayRemove(unsafe.Pointer(ct.ct), C.int(i))
			}
			return ct
		} else {
			return nil
		}
	case TypeObject:
		members := ct.GetMemberMapOrNil()
		if len(members) == 0 {
			return nil
		} else {
			start := len(members)
			var removes []string
			removed := 0
			for k, v := range members {
				if filtered := v.StripNulls(leaveEmptyArray); filtered == nil {
					removes = append(removes, k)
				}
			}
			for _, member := range removes {
				ct.RemoveMember(member)
				removed = removed + 1
			}
			if removed == start {
				return nil
			} else {
				return ct
			}
		}
	case TypeString, TypeNumber, TypeTrue, TypeFalse:
		return ct
	default:
		return nil
	}
}

// new style - no errors (returns nil instead), can be chained
func (ct *Container) GetMemberCountOrNil() int {
	if ct == nil {
		return 0
	} else if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return int(C.GetMemberCount(unsafe.Pointer(ct.ct)))
	} else {
		return 0
	}
}

func (ct *Container) GetMemberNamesOrNil() []string {
	count := ct.GetMemberCountOrNil()
	result := make([]string, count)
	if count == 0 {
		return result
	}
	for i := 0; i < count; i++ {
		result[i] = ct.GetMemberName(i)
	}
	return result
}

func (ct *Container) GetMemberMapOrNil() map[string]*Container {
	count := ct.GetMemberCountOrNil()
	result := make(map[string]*Container, count)
	if count == 0 {
		return result
	}
	members, _ := ct.GetMemberNames()
	for _, m := range members {
		result[m] = ct.GetMemberOrNil(m)
	}
	return result
}

func (ct *Container) GetMemberOrNil(key string) *Container {
	if ct == nil {
		return nil
	}
	cStr := C.CString(key)
	defer C.free(unsafe.Pointer(cStr))
	if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		if ct.HasMember(key) {
			var m Container
			m.doc = ct.doc
			m.ct = C.GetMember(unsafe.Pointer(ct.ct), cStr)
			return &m
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (ct *Container) GetPathContainerOrNil(path string) *Container {
	if ct == nil {
		return nil
	}
	keys := strings.Split(path, ".")
	next := ct
	var err error
	for _, key := range keys {
		next, err = next.GetMember(key)
		if err != nil {
			return nil
		}
	}

	return next
}

func (ct *Container) GetIntArrayOrNil() []int {
	if ct == nil {
		return make([]int, 0)
	}
	count, err := ct.GetArraySize()
	result := make([]int, count)
	if err != nil {
		return result
	}

	for i := 0; i < count; i++ {
		item := ct.GetArrayValue(i)
		result[i], err = item.GetInt()
		if err != nil {
			return make([]int, 0)
		}
	}
	return result
}

func (ct *Container) GetStringOrNil() []string {
	if ct == nil {
		return make([]string, 0)
	}
	count, err := ct.GetArraySize()
	result := make([]string, count)
	if err != nil {
		return result
	}

	for i := 0; i < count; i++ {
		item := ct.GetArrayValue(i)
		result[i], err = item.GetString()
		if err != nil {
			return make([]string, 0)
		}
	}
	return result
}

func (ct *Container) GetArrayOrNil() []*Container {
	count, err := ct.GetArraySize()
	result := make([]*Container, count)
	if err != nil || count == 0 {
		return result
	}

	for i := 0; i < count; i++ {
		result[i] = ct.GetArrayValue(i)
	}
	return result
}
