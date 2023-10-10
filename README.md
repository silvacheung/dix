# dix
dix is golang simple dependency inject (di) kit lib

### 1.Dependencies can be injected through providers
#### .Implement dix.Provider interface
``` go
type XFieldProvider struct{}

func (XFieldProvider) Symbol() string {
	return "xfp"
}

func (XFieldProvider) Provide(ctx context.Context, tag *Tag) (any, error) {
	return "x_field", nil
}
```
#### .Using tags in the struct using values from the provider's Symbol method
``` go
// X is a target di struct
type X struct {
    Field1 string `dix:"from:xfp"`
    Field2 string `dix:"from:xfp"`
}
```
#### .Bind registration provider and call DI
``` go
func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding Provider using namespace `def`
    dix.Binding[dix.Provider](XFieldProvider{})
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field1, x.Field2)
}
```

### 2.Support specifying a namespace, if not specified, use the default namespace
``` go
// X is a target di struct
type X struct {
    Field1 string `dix:"from:xfp;namespace:ns1"`
    Field2 string `dix:"from:xfp;namespace:ns2"`
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding Provider using namespace `ns1, ns2`
    dix.Binding[dix.Provider](XFieldProvider{}, "ns1", "ns2")
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field1, x.Field2)
}
```

### 3.Support binding interface implementation using namespace, The interface under the specified namespace will be injected into the implementation of the binding, The implementation of the interface will automatically inject (look at Article 5)
``` go
// StringerImpl implement fmt.Stringer
type StringerImpl struct{
	
}

func (StringerImpl) String() string {
    return "string_impl"
}

// X is a target di struct
type X struct {
    Field1 fmt.Stringer `dix:"from:?;namespace:ns1"`
    Field2 fmt.Stringer `dix:"from:?;namespace:ns2"`
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding fmt.Stringer implement using namespace `ns1, ns2`
    dix.Binding[fmt.Stringer](StringerImpl{}, "ns1", "ns2")
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field1, x.Field2)
}
```

### 4.Supports binding of specified type values using namespace, The type under the specified namespace will be injected into the values of the binding
``` go
// X is a target di struct
type X struct {
    Field1 string `dix:"from:?;namespace:ns1"`
    Field2 int `dix:"from:?;namespace:ns1"`
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding type values using namespace `ns1`
    dix.Binding[string]("stringValue", "ns1")
    dix.Binding[int](100, "ns1")
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field1, x.Field2)
}
```

### 5.Supports automatic injection of struct and pointer struct
``` go
// StringerImpl implement fmt.Stringer
type StringerImpl struct{
	Field string `dix:"from:?"` // Will be automatic di
}

func (StringerImpl) String() string {
    return "string_impl"
}

// X is a target di struct
type X struct {
    Field string `dix:"from:?"`
    FieldS fmt.Stringer `dix:"from:?"` // Will be use `StringerImpl{}` and automatic di StringerImpl's fields
    FieldY Y `dix:"from:?"` // This field Y struct will be automatic di
}

// Y is a target di struct
type Y struct {
     Field string `dix:"from:?"` // This field will be automatic injection values
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding type values using namespace `def`
    dix.Binding[string]("stringValue")
    dix.Binding[fmt.Stringer](StringerImpl{})
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field, x.FieldS.Field, x.FieldY.Field)
}
```

### 6.Support cycled dependency injection check
``` go
// X is a target di struct
type X struct {
    Field string `dix:"from:?"`
    FieldY Y `dix:"from:?"` // This field Y struct will be automatic di
}

// Y is a target di struct
type Y struct {
     Field string `dix:"from:?"` // This field will be automatic injection values
     FieldX X `dix:"from:?"` // Cycled dependency field
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding type values using namespace `def`
    dix.Binding[string]("stringValue")
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // !!! Will be get a cycled dependency di error
    }
}
```

### 7.Support provider's custom tags
``` go
type XFieldProvider struct{}

func (XFieldProvider) Symbol() string {
	return "xfp"
}

func (XFieldProvider) Provide(ctx context.Context, tag *Tag) (any, error) {
	// You can obtain custom tags in the structure when calling DI
	tagValue, ok := tag.GetCustomize("kind")
	if ok && tagValue == "kind01" {
	    return "kind01", nil
	}
	return "x_field", nil
}

// X is a target di struct
type X struct {
    Field1 string `dix:"from:xfp;kind:kind01"`
    Field2 string `dix:"from:xfp;kind:kind02"`
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Binding Provider using namespace `def`
    dix.Binding[dix.Provider](XFieldProvider{})
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
    
    // Print x
    fmt.Println(x.Field1, x.Field2)
}
```

### 8.When there is no binding value of the specified type, the type value can be instantiated, and instantiation of map, slice, chan, array, and other types of zero values can be supported
``` go
// X is a target di struct
type X struct {
    Field0 string `dix:"from:?"` // Zero value ''
    Field1 int `dix:"from:?"` // Zero value 0
    Field2 []string `dix:"from:?;slice_len:3;slice_cap=3"` // Use slice_len and slice_cap like `make([]string, 3, 3)` 
    Field3 [5]string `dix:"from:?"` // Like `new([5]string)`
    Field4 map[string]string `dix:"from:?;map_size:3"` // Use map_size like `make(map[string]string, 3)`
    Field5 chan string `dix:"from:?;chan_buf:3"` // Use chan_buf like `make(chan string, 3)`
}

func main () {
    // Logging dix log
    dix.Logging(true)
    
    // Call DI method make a di object
    x, err := dix.DI[X](context.Background())
    if err != nil {
        // ...
    }
}
```

### 9.Usage suggestions
#### .Using single instance injection mode
#### .Do not inject unnecessary fields
#### .Using Tree Hierarchy for Injection
