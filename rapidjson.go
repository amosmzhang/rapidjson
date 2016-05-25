package rapidjson

// #cgo CFLAGS: -I. -fpic
// #cgo LDFLAGS: -L. -lrjwrapper
// #include <stdlib.h>
// #include "rjwrapper.h"
import "C"
import "unsafe"

import (
	"errors"
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
func NewDoc() Doc {
	var json Doc
	json.json = C.JsonInit()
	return json
}
func (json *Doc) Free() {
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
	C.ValFree(unsafe.Pointer(ct.ct))
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
	ctStr := ct.String()
	copyDoc, _ := NewParsedStringJson(ctStr)

	ctCopy := copyDoc.GetContainer()
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
func NewParsedJson(input []byte) (Doc, error) {
	doc := NewDoc()
	err := doc.Parse(input)
	return doc, err
}
func NewParsedStringJson(input string) (Doc, error) {
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
func (json *Doc) Bytes() []byte {
	return []byte(json.String())
}

// various getters
func (ct *Container) HasMember(key string) bool {
	if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		return CBoolTest(C.HasMember(unsafe.Pointer(ct.ct), cStr))
	} else {
		return false
	}
}
func (ct *Container) GetMemberCount() (int, error) {
	if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return int(C.GetMemberCount(unsafe.Pointer(ct.ct))), nil
	} else {
		return 0, ErrNotObject
	}
}
func (ct *Container) GetMemberName(index int) string {
	cStr := C.GetMemberName(unsafe.Pointer(ct.ct), C.int(index))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) GetMemberNames() ([]string, error) {
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
	cStr := C.CString(key)
	defer C.free(unsafe.Pointer(cStr))
	if CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		var m Container
		if ct.HasMember(key) {
			m.doc = ct.doc
			m.ct = C.GetMember(unsafe.Pointer(ct.ct), cStr)
			return &m, nil
		} else {
			return &m, ErrPathNotFound
		}
	} else {
		var m Container
		return &m, ErrNotObject
	}
}
func (ct *Container) String() string {
	cStr := C.ValGetString(unsafe.Pointer(ct.ct))
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) Bytes() []byte {
	return []byte(ct.String())
}

func (ct *Container) GetPathContainer(path string) (*Container, error) {
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
	_, err := ct.GetPathContainer(path)
	if err != nil {
		return false
	} else {
		return true
	}
}
func (ct *Container) GetPathNewContainer(path string) (*Container, error) {
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

// typed getters
func (ct *Container) GetType() int {
	return int(C.GetType(unsafe.Pointer(ct.ct)))
}
func (ct *Container) GetInt() (int, error) {
	if CBoolTest(C.IsInt(unsafe.Pointer(ct.ct))) {
		result := int(C.ValGetInt(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result int
		return result, ErrNotInt
	}
}
func (ct *Container) GetFloat() (float64, error) {
	if CBoolTest(C.IsDouble(unsafe.Pointer(ct.ct))) {
		result := float64(C.ValGetDouble(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result float64
		return result, ErrNotFloat
	}
}
func (ct *Container) GetBool() (bool, error) {
	if CBoolTest(C.IsBool(unsafe.Pointer(ct.ct))) {
		result := CBoolTest(C.ValGetBool(unsafe.Pointer(ct.ct)))
		return result, nil
	} else {
		var result bool
		return result, ErrNotBool
	}
}
func (ct *Container) GetString() (string, error) {
	if CBoolTest(C.IsString(unsafe.Pointer(ct.ct))) {
		cStr := C.ValGetBasicString(unsafe.Pointer(ct.ct))
		defer C.free(unsafe.Pointer(cStr))
		str := C.GoString(cStr)
		return str, nil
	} else {
		var result string
		return result, ErrNotString
	}
}
func (ct *Container) GetArraySize() (int, error) {
	if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		size := int(C.ValArraySize(unsafe.Pointer(ct.ct)))
		return size, nil
	} else {
		return 0, ErrNotArray
	}
}
func (ct *Container) GetArrayValue(index int) *Container {
	var a Container
	a.doc = ct.doc
	a.ct = C.GetArrayValueAt(unsafe.Pointer(ct.ct), C.int(index))
	return &a
}

func (ct *Container) GetIntArray() ([]int, error) {
	count, err := ct.GetArraySize()
	result := make([]int, count)
	if err != nil {
		return result, err
	}

	for i := 0; i < count; i++ {
		item := ct.GetArrayValue(i)
		result[i], err = item.GetInt()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
func (ct *Container) GetStringArray() ([]string, error) {
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
	if v == nil {
		C.SetNull(unsafe.Pointer(ct.ct))
		return nil
	}

	switch v.(type) {
	case int:
		C.SetInt(unsafe.Pointer(ct.ct), C.int(v.(int)))
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
	C.SetValue(unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
}
func (ct *Container) InitObj() {
	ct.ct = C.InitObj(unsafe.Pointer(ct.ct))
}
func (ct *Container) AddValue(key string, v interface{}) error {
	item := ct.doc.NewContainer()
	err := item.SetValue(v)
	if err != nil {
		return err
	}

	return ct.AddMember(key, item)
}
func (ct *Container) AddMember(key string, item *Container) error {
	if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
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
func (ct *Container) AddMemberArray(key string, items []*Container) error {
	if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
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
	item := ct.doc.NewContainer()
	err := item.SetValue(v)
	if err != nil {
		return err
	}

	return ct.SetMember(key, item)
}
func (ct *Container) AddMemberAtPath(path string, item *Container) error {
	dest, err := ct.GetPathNewContainer(path)
	if err != nil {
		return err
	}
	dest.SetContainer(item)
	return nil
}
func (ct *Container) AddValueAtPath(path string, v interface{}) error {
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
	C.InitArray(unsafe.Pointer(ct.ct))
}
func (ct *Container) ArrayAppendContainer(item *Container) error {
	if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		C.ArrayAppend(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), unsafe.Pointer(item.ct))
		return nil
	} else {
		return ErrNotArray
	}
}
func (ct *Container) ArrayAppendCopy(item *Container) error {
	if CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		itemCopy := item.GetCopy()
		ct.doc.allocated = append(ct.doc.allocated, itemCopy.doc)
		C.ArrayAppend(unsafe.Pointer(ct.doc.json), unsafe.Pointer(ct.ct), unsafe.Pointer(itemCopy.ct))
		return nil
	} else {
		return ErrNotArray
	}
}
func (ct *Container) ArrayAppend(v interface{}) error {
	item := ct.doc.NewContainer()
	item.SetValue(v)
	return ct.ArrayAppendContainer(item)
}

// deleters
func (ct *Container) RemoveMember(key string) error {
	if !CBoolTest(C.IsObj(unsafe.Pointer(ct.ct))) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		C.RemoveMember(unsafe.Pointer(ct.ct), cStr)
	}
	return nil
}
func (ct *Container) ArrayRemove(index int) error {
	if !CBoolTest(C.IsArray(unsafe.Pointer(ct.ct))) {
		return ErrNotArray
	} else if int(C.ValArraySize(unsafe.Pointer(ct.ct))) <= index {
		return ErrOutOfBounds
	} else {
		C.ArrayRemove(unsafe.Pointer(ct.ct), C.int(index))
	}

	return nil
}
