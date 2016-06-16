# rapidjson

Go package using [C++ rapidjson](https://github.com/miloyip/rapidjson) for better JSON handling 

# Quick start

In package main:

    import "github.com/bottlenose-inc/rapidjson"
    ...
    json, err := rapidjson.NewParsedStringJson(`{"project":"rapidjson","stars":10,"use":"everywhere"}`)
    ...
    fmt.Println(json.String())

# Key concepts

rapidjson has two types: Container and Doc. Container is a generalized value type, and can take on any specific type (int, array, object, etc.). Doc is a specialized Container with additional parsing and memory allocation functionality. In general, create one Doc, get its Container, and work with that Container. Doc is a pointer to C land, and should be freed manually. Key values pairs in rapidjson objects are referred to as members.

# Parsing

    func (json *Doc) Parse(input []byte) error
    func (json *Doc) ParseString(input string) error
    func NewParsedJson(input []byte) (*Doc, error)
    func NewParsedStringJson(input string) (*Doc, error)
    func (json *Doc) HasParseError() bool

A call to HasParseError() is included at the end of each parsing func, and error is returned.

# Getters

For outputting:

    func (json *Doc) String() string
    func (json *Doc) Bytes() []byte
    func (ct *Container) String() string
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
    func (ct *Container) PathExists(path string) bool
    func (ct *Container) GetPathNewContainer(path string) (*Container, error)

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

Making new Container, use root Doc to use memory allocator. Freeing Doc will free associated Containers:

    func (json *Doc) NewContainer() *Container
    func (json *Doc) NewContainerObj() *Container
    func (json *Doc) NewContainerArray() *Container

# Setters

SetValue() can be used for basic types int (all sizes), float64, bool, string, nil. Setters will overwrite previous type.

    func (ct *Container) SetValue(v interface{}) error
    func (ct *Container) SetContainer(item *Container)
    func (ct *Container) InitObj()
    func (ct *Container) AddValue(key string, v interface{}) error
    func (ct *Container) AddMember(key string, item *Container) error
    func (ct *Container) AddMemberArray(key string, items []*Container) error
    func (ct *Container) AddMemberAtPath(path string, item *Container) error
    func (ct *Container) AddValueAtPath(path string, v interface{}) error
    func (ct *Container) SetMember(key string, item *Container) error
    func (ct *Container) SetMemberValue(key string, v interface{}) error 
    func (ct *Container) InitArray()
    func (ct *Container) ArrayAppendContainer(item *Container) error
    func (ct *Container) ArrayAppendCopy(item *Container) error
    func (ct *Container) ArrayAppend(v interface{}) error

# Removes

    func (ct *Container) RemoveMember(key string) error
    func (ct *Container) ArrayRemove(index int) error

# Value types:

	TypeNull   = 0
	TypeFalse  = 1
	TypeTrue   = 2
	TypeObject = 3
	TypeArray  = 4
	TypeString = 5
	TypeNumber = 6

# Error types

	ErrJsonParse    - JSON parsing error
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
