# rapidjson

Fast go json parser wrapping [C++ rapidjson](https://github.com/miloyip/rapidjson)

# Quick start

In package main:

    import "github.com/bottlenose-inc/rapidjson"
    ...
    json, err := rapidjson.NewParsedStringJson(`{"project":"rapidjson","stars":10,"use":"everywhere"}`)
    ...
    fmt.Println(json.String())

# Key concepts

rapidjson has two types: Container and Doc. Container is a generalized value type, and can take on any specific type (int, array, object, etc.). Doc is a specialized Container with additional parsing and memory allocation functionality. In general, create one Doc, get its Container, and work with that Container. Doc should be freed manually, but Containers associated with a Doc will be freed when Doc is freed. Key values pairs in rapidjson objects are referred to as members.

# Parsing

    func (json *Doc) Parse(input []byte) error
    func (json *Doc) ParseString(input string) error
    func NewParsedJson(input []byte) (*Doc, error)
    func NewParsedStringJson(input string) (*Doc, error)
    func (json *Doc) HasParseError() bool

A call to HasParseError() is included at the end of each parsing func, and error is returned with details.

# Getters

For outputting:

    func (json *Doc) String() string
    func (json *Doc) Pretty() string
    func (json *Doc) Bytes() []byte
    func (ct *Container) String() string
    func (ct *Container) Pretty() string
    func (ct *Container) Bytes() []byte

Getting a Doc's Container:

    func (json *Doc) GetContainer() *Container
    func (json *Doc) GetContainerNewObj() *Container

Working with Containers:

    func (ct *Container) HasMember(key string) bool
    func (ct *Container) GetMemberCount() (int, error)
    func (ct *Container) GetMemberName(index int) string
    func (ct *Container) GetMemberNames() ([]string, error)
    func (ct *Container) GetMemberMap() (map[string]*Container, error)
    func (ct *Container) GetMember(key string) (*Container, error)
    func (ct *Container) GetPathContainer(path string) (*Container, error)
    func (ct *Container) GetPathNewContainer(path string) (*Container, error)
    func (ct *Container) IsEqual(other *Container) bool

Typed getters:

    func (ct *Container) GetType() int
    func (ct *Container) GetInt() (int, error)
    func (ct *Container) GetInt64() (int64, error)
    func (ct *Container) GetFloat() (float64, error)
    func (ct *Container) GetBool() (bool, error)
    func (ct *Container) GetString() (string, error)
    func (ct *Container) GetArraySize() (int, error)
    func (ct *Container) GetArrayValue(index int) *Container
    func (ct *Container) GetIntArray() ([]int, error)
    func (ct *Container) GetStringArray() ([]string, error)
    func (ct *Container) GetArray() ([]*Container, int, error)

    func (ct *Container) GetValue() (interface{}, error)

Making new Container, use root Doc to use memory allocator. Freeing Doc will free associated Containers:

    func (json *Doc) NewContainer() *Container
    func (json *Doc) NewContainerObj() *Container
    func (json *Doc) NewContainerArray() *Container

# Setters

SetValue() can be used for basic types int (all sizes), float64, bool, string, nil. Setters will overwrite previous type.

    func (ct *Container) SetValue(v interface{}) error
    func (ct *Container) SetContainer(item *Container)
    func (ct *Container) SetContainerCopy(item *Container)
    func (ct *Container) InitObj()
    func (ct *Container) AddValue(key string, v interface{}) error
    func (ct *Container) AddMember(key string, item *Container) error
    func (ct *Container) AddMemberCopy(key string, item *Container) error
    func (ct *Container) AddMemberArray(key string, items []*Container) error
    func (ct *Container) AddMemberAtPath(path string, item *Container) error
    func (ct *Container) AddValueAtPath(path string, v interface{}) error
    func (ct *Container) SetMember(key string, item *Container) error
    func (ct *Container) SetMemberCopy(key string, item *Container) error
    func (ct *Container) SetMemberValue(key string, v interface{}) error
    func (ct *Container) InitArray()
    func (ct *Container) ArrayAppendContainer(item *Container) error
    func (ct *Container) ArrayAppendCopy(item *Container) error
    func (ct *Container) ArrayAppend(v interface{}) error

# Errorless

This set of functions duplicate functionality in some previous functions, but do not return errors so that they can be chained.

    func (ct *Container) GetMemberCountOrNil() int
    func (ct *Container) GetMemberNamesOrNil() []string
    func (ct *Container) GetMemberMapOrNil() map[string]*Container
    func (ct *Container) GetMemberOrNil(key string) *Container
    func (ct *Container) GetPathContainerOrNil(path string) *Container
    func (ct *Container) GetIntArrayOrNil() []int
    func (ct *Container) GetStringOrNil() []string
    func (ct *Container) GetArrayOrNil() []*Container

Usage example:

    example, err := container.GetMemberOrNil("path1").GetMemberOrNil("path2").GetString()

# Removes

    func (ct *Container) RemoveMember(key string) error
    func (ct *Container) ArrayClear() error
    func (ct *Container) ArrayRemove(index int) error
    func (ct *Container) RemoveMemberAtPath(path string) error
    func (ct *Container) StripNulls(leaveEmptyArray bool) *Container

# Value types:

	TypeNull   = 0
	TypeFalse  = 1
	TypeTrue   = 2
	TypeObject = 3
	TypeArray  = 4
	TypeString = 5
	TypeNumber = 6

# Error types

	ErrNotArray     - Not an array
	ErrNotObject    - Not an object
	ErrPathNotFound - Path not found
	ErrNotInt       - Not an int
	ErrNotFloat     - Not a float
	ErrNotBool      - Not a bool
	ErrNotString    - Not a string
	ErrBadType      - Bad type
	ErrMemberExists - Member already exists
	ErrOutOfBounds  - Array index out of bounds

# Benchmarks

The following benchmark test was performed by reading 100,000 lines of JSON from a file (~1.5kb per line), and parsing each with Go's `encoding/json` as well as `rapidjson`. One binary was built using each library (all other code was identical) and both were run 5 times on a Mid 2015 Macbook Pro (2.8 GHz Intel Core i7).

RJ took **a full second less** to complete this operation, running in about **59%** of the time it took Go's library to do the same task.

Using `rapidjson`: (average total **2.0018s**)

    $ time ./rj
    100000 documents parsed
    ./rj  2.10s user 0.29s system 116% cpu 2.047 total
    $ time ./rj
    100000 documents parsed
    ./rj  2.05s user 0.27s system 118% cpu 1.958 total
    $ time ./rj
    100000 documents parsed
    ./rj  2.11s user 0.27s system 118% cpu 2.010 total
    $ time ./rj
    100000 documents parsed
    ./rj  2.07s user 0.27s system 117% cpu 1.985 total
    $ time ./rj
    100000 documents parsed
    ./rj  2.09s user 0.27s system 117% cpu 2.009 total

Using `encoding/json`: (average total **3.3674s**)

    $ time ./json
    100000 documents parsed
    ./json  3.40s user 0.12s system 105% cpu 3.335 total
    $ time ./json
    100000 documents parsed
    ./json  3.46s user 0.12s system 105% cpu 3.374 total
    $ time ./json
    100000 documents parsed
    ./json  3.44s user 0.11s system 105% cpu 3.349 total
    $ time ./json
    100000 documents parsed
    ./json  3.50s user 0.12s system 105% cpu 3.412 total
    $ time ./json
    100000 documents parsed
    ./json  3.45s user 0.11s system 105% cpu 3.367 total
