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
	"sync"
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

	rjMutex = &sync.Mutex{}
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

type Doc struct {
	json      C.JsonDoc
	allocated []*Container
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
	rjMutex.Lock()
	var json Doc
	json.json = C.JsonInit()
	rjMutex.Unlock()
	return json
}
func (json *Doc) Free() {
	rjMutex.Lock()
	for _, ct := range json.allocated {
		ct.Free()
	}
	C.JsonFree(json.json)
	rjMutex.Unlock()
}
func (json *Doc) NewContainer() Container {
	rjMutex.Lock()
	var ct Container
	ct.doc = json
	ct.ct = C.ValInit()
	json.allocated = append(json.allocated, &ct)
	rjMutex.Unlock()
	return ct
}
func (json *Doc) NewContainerObj() Container {
	ct := json.NewContainer()
	ct.InitObj()
	return ct
}
func (json *Doc) NewContainerArray() Container {
	ct := json.NewContainer()
	ct.InitArray()
	return ct
}
func (ct *Container) Free() {
	C.ValFree(ct.ct)
}
func (json *Doc) GetContainer() Container {
	var ct Container
	ct.ct = C.GetValue(json.json)
	ct.doc = json
	return ct
}
func (json *Doc) GetContainerNewObj() Container {
	var ct Container
	ct.ct = C.GetValue(json.json)
	ct.doc = json
	ct.InitObj()
	return ct
}

func GetDocCount(side int) int {
	return int(C.GetDocCount(C.int(side)))
}

func GetCtCount(side int) int {
	return int(C.GetValCount(C.int(side)))
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
	C.JsonParse(json.json, cStr)

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
	return CBoolTest(C.HasParseError(json.json))
}

// get string/bytes output
func (json *Doc) String() string {
	cStr := C.GetString(json.json)
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (json *Doc) Bytes() []byte {
	return []byte(json.String())
}

// various getters
func (ct *Container) HasMember(key string) bool {
	if CBoolTest(C.IsObj(ct.ct)) {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		return CBoolTest(C.HasMember(ct.ct, cStr))
	} else {
		return false
	}
}
func (ct *Container) GetMemberCount() (int, error) {
	if CBoolTest(C.IsObj(ct.ct)) {
		return int(C.GetMemberCount(ct.ct)), nil
	} else {
		return 0, ErrNotObject
	}
}
func (ct *Container) GetMemberName(index int) string {
	cStr := C.GetMemberName(ct.ct, C.int(index))
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
func (ct *Container) GetMemberMap() (map[string]Container, error) {
	count, err := ct.GetMemberCount()
	result := make(map[string]Container, count)
	if err != nil {
		return result, err
	}
	members, _ := ct.GetMemberNames()
	for _, m := range members {
		result[m], _ = ct.GetMember(m)
	}
	return result, nil
}

func (ct *Container) GetMember(key string) (Container, error) {
	cStr := C.CString(key)
	defer C.free(unsafe.Pointer(cStr))
	if CBoolTest(C.IsObj(ct.ct)) {
		var m Container
		if ct.HasMember(key) {
			m.doc = ct.doc
			m.ct = C.GetMember(ct.ct, cStr)
			return m, nil
		} else {
			return m, ErrPathNotFound
		}
	} else {
		var m Container
		return m, ErrNotObject
	}
}
func (ct *Container) String() string {
	cStr := C.ValGetString(ct.ct)
	defer C.free(unsafe.Pointer(cStr))
	str := C.GoString(cStr)
	return str
}
func (ct *Container) Bytes() []byte {
	return []byte(ct.String())
}

func (ct Container) GetPathContainer(path string) (Container, error) {
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
func (ct Container) GetPathNewContainer(path string) (Container, error) {
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
	return int(C.GetType(ct.ct))
}
func (ct *Container) GetInt() (int, error) {
	if CBoolTest(C.IsInt(ct.ct)) {
		result := int(C.ValGetInt(ct.ct))
		return result, nil
	} else {
		var result int
		return result, ErrNotInt
	}
}
func (ct *Container) GetFloat() (float64, error) {
	if CBoolTest(C.IsDouble(ct.ct)) {
		result := float64(C.ValGetDouble(ct.ct))
		return result, nil
	} else {
		var result float64
		return result, ErrNotFloat
	}
}
func (ct *Container) GetBool() (bool, error) {
	if CBoolTest(C.IsBool(ct.ct)) {
		result := CBoolTest(C.ValGetBool(ct.ct))
		return result, nil
	} else {
		var result bool
		return result, ErrNotBool
	}
}
func (ct *Container) GetString() (string, error) {
	if CBoolTest(C.IsString(ct.ct)) {
		cStr := C.ValGetBasicString(ct.ct)
		str := C.GoString(cStr)
		return str, nil
	} else {
		var result string
		return result, ErrNotString
	}
}
func (ct *Container) GetArraySize() (int, error) {
	if CBoolTest(C.IsArray(ct.ct)) {
		size := int(C.ValArraySize(ct.ct))
		return size, nil
	} else {
		return 0, ErrNotArray
	}
}
func (ct *Container) GetArrayValue(index int) Container {
	var a Container
	a.doc = ct.doc
	a.ct = C.GetArrayValueAt(ct.ct, C.int(index))
	return a
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
func (ct *Container) GetArray() ([]Container, int, error) {
	count, err := ct.GetArraySize()
	result := make([]Container, count)
	if err != nil {
		return result, 0, err
	}

	for i := 0; i < count; i++ {
		result[i] = ct.GetArrayValue(i)
	}
	return result, result[0].GetType(), nil
}

// setters
func (ct *Container) SetValue(v interface{}) error {
	if v == nil {
		C.SetNull(ct.ct)
		return nil
	}

	switch v.(type) {
	case int:
		C.SetInt(ct.ct, C.int(v.(int)))
		return nil
	case float64:
		C.SetDouble(ct.ct, C.double(v.(float64)))
		return nil
	case bool:
		C.SetBool(ct.ct, BoolToC(v.(bool)))
		return nil
	case string:
		cStr := C.CString(v.(string))
		defer C.free(unsafe.Pointer(cStr))
		C.SetString(ct.doc.json, ct.ct, cStr)
		return nil
	default:
		return ErrBadType
	}
}
func (ct *Container) SetContainer(item Container) {
	C.SetValue(ct.ct, item.ct)
}
func (ct *Container) InitObj() {
	ct.ct = C.InitObj(ct.ct)
}
func (ct *Container) AddValue(key string, v interface{}) error {
	item := ct.doc.NewContainer()
	err := item.SetValue(v)
	if err != nil {
		return err
	}

	return ct.AddMember(key, item)
}
func (ct *Container) AddMember(key string, item Container) error {
	if !CBoolTest(C.IsObj(ct.ct)) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		if CBoolTest(C.HasMember(ct.ct, cStr)) {
			return ErrMemberExists
		} else {
			C.AddMember(ct.doc.json, ct.ct, cStr, item.ct)
			return nil
		}
	}
}
func (ct *Container) SetMember(key string, item Container) error {
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
func (ct *Container) AddMemberAtPath(path string, item Container) error {
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
	C.InitArray(ct.ct)
}
func (ct *Container) ArrayAppendContainer(item Container) error {
	if CBoolTest(C.IsArray(ct.ct)) {
		C.ArrayAppend(ct.doc.json, ct.ct, item.ct)
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
	if !CBoolTest(C.IsObj(ct.ct)) {
		return ErrNotObject
	} else {
		cStr := C.CString(key)
		defer C.free(unsafe.Pointer(cStr))
		C.RemoveMember(ct.ct, cStr)
	}
	return nil
}
func (ct *Container) ArrayRemove(index int) error {
	if !CBoolTest(C.IsArray(ct.ct)) {
		return ErrNotArray
	} else if int(C.ValArraySize(ct.ct)) <= index {
		return ErrOutOfBounds
	} else {
		C.ArrayRemove(ct.ct, C.int(index))
	}

	return nil
}
