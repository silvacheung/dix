package dix

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"unsafe"
)

type TBProvider struct{}

func (TBProvider) Symbol() string {
	return "TBP"
}

func (TBProvider) Provide(ctx context.Context, tag *Tag) (any, error) {
	return TB{
		Int8:          100,
		Int16:         100,
		Int32:         100,
		Int64:         100,
		Uint:          100,
		Uint8:         100,
		Uint16:        100,
		Uint32:        100,
		Uint64:        100,
		Float64:       100,
		Complex128:    100,
		Array:         &[5]int{0, 1, 2, 3, 4},
		Func:          nil,
		Interface:     nil,
		Pointer:       nil,
		UnsafePointer: nil,
	}, nil
}

// TA is testing struct
// use `from` tag set inject provider, tag value is provider' Symbol string, `from:?` means use reflect create inject
// use `namespace` tag set the di working space
// use `chan_buf` tag on chan type field set the channel buffer
// use `map_size` tag set the map size
// use `slice_len`/`slice_cap` tag set the slice len and cap
// use other tag can use in Provider's Provide method args
// !!! not circular reference, otherwise, there will be a dead cycle
type TA struct {
	Bool           bool                        `dix:"from:?"`
	Int10          int                         `dix:"from:?"`
	Int100         int                         `dix:"from:?;namespace:int100"`
	Int8           int8                        `dix:"from:?"`
	Int16          int16                       `dix:"from:?"`
	Int32          int32                       `dix:"from:?"`
	Int64          int64                       `dix:"from:?"`
	Uint           uint                        `dix:"from:?"`
	Uint8          uint8                       `dix:"from:?"`
	Uint16         uint16                      `dix:"from:?"`
	Uint32         uint32                      `dix:"from:?"`
	Uint64         uint64                      `dix:"from:?"`
	Uintptr        uintptr                     `dix:"from:?"`
	Float32        float32                     `dix:"from:?"`
	Float64        float64                     `dix:"from:?"`
	Complex64      complex64                   `dix:"from:?"`
	Complex128     complex128                  `dix:"from:?"`
	Array          [5]int                      `dix:"from:?"`
	Func           func(context.Context) error `dix:"from:?"`
	Interface      interface{}                 `dix:"from:?"`
	Chan           chan any                    `dix:"from:?;chan_buf:10"`
	Map            map[any]any                 `dix:"from:?;map_size:10"`
	Slice          []string                    `dix:"from:?;slice_len:0;slice_cap:10"`
	Pointer        *string                     `dix:"from:?"`
	String         string                      `dix:"from:?"`
	Stringer1      fmt.Stringer                `dix:"from:?;namespace:stringer1"`
	Stringer2      fmt.Stringer                `dix:"from:?;namespace:stringer2"`
	Struct         struct{}                    `dix:"from:?"`
	UnsafePointer  unsafe.Pointer              `dix:"from:?"`
	TB             TB                          `dix:"from:TBP"`
	TBPointer      *TB                         `dix:"from:?"`
	TBSlice        []TB                        `dix:"from:?;slice_len:0;slice_cap:10"`
	TBPointerSlice []*TB                       `dix:"from:?;slice_len:0;slice_cap:10"`
	//TA           *TA                         `dix:"from:?"` // !!! not circular reference
}

// TB is testing struct
// use `from` tag set inject provider, tag value is provider' Symbol string, `from:?` means use reflect create inject
// use `namespace` tag set the di working space
// use `chan_buf` tag on chan type field set the channel buffer
// use `map_size` tag set the map size
// use `slice_len`/`slice_cap` tag set the slice len and cap
// use other tag can use in Provider's Provide method args
// !!! not circular reference, otherwise, there will be a dead cycle
type TB struct {
	Bool          *bool                        `dix:"from:?"`
	Int           *int                         `dix:"from:?"`
	Int8          int8                         `dix:"from:?"`
	Int16         int16                        `dix:"from:?"`
	Int32         int32                        `dix:"from:?"`
	Int64         int64                        `dix:"from:?"`
	Uint          uint                         `dix:"from:?"`
	Uint8         uint8                        `dix:"from:?"`
	Uint16        uint16                       `dix:"from:?"`
	Uint32        uint32                       `dix:"from:?"`
	Uint64        uint64                       `dix:"from:?"`
	Uintptr       *uintptr                     `dix:"from:?"`
	Float32       *float32                     `dix:"from:?"`
	Float64       float64                      `dix:"from:?"`
	Complex64     *complex64                   `dix:"from:?"`
	Complex128    complex128                   `dix:"from:?"`
	Array         *[5]int                      `dix:"from:?"`
	Chan          *chan any                    `dix:"from:?"`
	Func          *func(context.Context) error `dix:"from:?"`
	Interface     *interface{}                 `dix:"from:?"`
	Map           *map[any]any                 `dix:"from:?"`
	Pointer       **string                     `dix:"from:?"`
	Slice         *[]string                    `dix:"from:?"`
	String        *string                      `dix:"from:?"`
	Struct        *struct{}                    `dix:"from:?"`
	UnsafePointer *unsafe.Pointer              `dix:"from:?"`
	//TAPointer     *TA                        `dix:"from:?"` // !!! not circular reference
	//TBPointer     **TB                       `dix:"from:?"` // !!! not circular reference
}

// TStringer is impl fmt.Stringer test object
type TStringer struct {
	Str string
	Int int `dix:"from:?;namespace:int100"`
}

func (s TStringer) String() string {
	return s.Str + "(" + strconv.Itoa(s.Int) + ")"
}

func ExampleDI() {
	// set logging
	Logging(true)

	// binding the Provider, Still working as a provider in namespace
	Binding[Provider](TBProvider{})        // namespace is 'def'
	Binding[Provider](TBProvider{}, "tbp") // namespace is 'tbp'

	// binding the type in namespace, Will automatically di if type is 'zero value', otherwise set the value of binding
	Binding[int](10)                                                // namespace is 'def'
	Binding[int](100, "int100")                                     // namespace is 'int100'
	Binding[fmt.Stringer](TStringer{Str: "stringer1"}, "stringer1") // namespace is 'stringer1'
	Binding[fmt.Stringer](TStringer{Str: "stringer2"}, "stringer2") // namespace is 'stringer2'
	Binding[TA](TA{Int100: 1111111111})                             // namespace is 'def'
	Binding[*TA](&TA{Int100: 2222222222})                           // namespace is 'def'

	// use DI method inject
	ea, err := DI[TA](context.Background())
	if err != nil {
		// do something
		return
	}

	// use the target field value
	fmt.Println("EA.Bool-->", ea.Bool)
	fmt.Println("EA.Int-->", ea.Int10, ea.Int100)
	fmt.Println("EA.Uintptr-->", ea.Uintptr)
	fmt.Println("EA.Float32-->", ea.Float32)
	fmt.Println("EA.Complex64-->", ea.Complex64)
	fmt.Println("EA.Array-->", ea.Array)
	fmt.Println("EA.Interface-->", ea.Interface)
	fmt.Println("EA.Chan-->", ea.Chan)
	fmt.Println("EA.Map-->", ea.Map)
	fmt.Println("EA.Slice-->", ea.Slice)
	fmt.Println("EA.Pointer-->", ea.Pointer)
	fmt.Println("EA.String-->", ea.String)
	fmt.Println("EA.Stringer-->", ea.Stringer1, ea.Stringer2)
	fmt.Println("EA.Struct-->", ea.Struct)
	fmt.Println("EA.UnsafePointer-->", ea.UnsafePointer)
	fmt.Println("EA.TB-->", ea.TB)
	fmt.Println("EA.TBPointer-->", ea.TBPointer)
	fmt.Println("EA.TBSlice-->", ea.TBSlice)
	fmt.Println("EA.TBPointerSlice-->", ea.TBPointerSlice)
}

func TestDI(t *testing.T) {
	// set logging
	Logging(true)

	// binding the type
	Binding[int](10)
	Binding[int](100, "int100")
	Binding[Provider](TBProvider{})
	Binding[fmt.Stringer](TStringer{Str: "stringer1"}, "stringer1")
	Binding[fmt.Stringer](&TStringer{Str: "stringer2"}, "stringer2")
	Binding[TA](TA{Int100: 1111111111})
	Binding[*TA](&TA{Int100: 2222222222})

	// use DI method inject
	ea, err := DI[TA](context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// use the target field value
	t.Log("TA.Bool-->", ea.Bool)
	t.Log("EA.Int-->", ea.Int10, ea.Int100)
	t.Log("TA.Uintptr-->", ea.Uintptr)
	t.Log("TA.Float32-->", ea.Float32)
	t.Log("TA.Complex64-->", ea.Complex64)
	t.Log("TA.Array-->", ea.Array)
	t.Log("TA.Interface-->", ea.Interface)
	t.Log("TA.Chan-->", ea.Chan, cap(ea.Chan))
	t.Log("TA.Map-->", ea.Map, len(ea.Map))
	t.Log("TA.Slice-->", ea.Slice, len(ea.Slice), cap(ea.Slice))
	t.Log("TA.Pointer-->", ea.Pointer)
	t.Log("TA.String-->", ea.String)
	t.Log("TA.Stringer-->", ea.Stringer1, ea.Stringer2)
	t.Log("TA.Struct-->", ea.Struct)
	t.Log("TA.UnsafePointer-->", ea.UnsafePointer)
	t.Log("TA.TB-->", ea.TB)
	t.Log("TA.TBPointer-->", ea.TBPointer)
	t.Log("TA.TBSlice-->", ea.TBSlice, len(ea.TBSlice), cap(ea.TBSlice))
	t.Log("TA.TBPointerSlice-->", ea.TBPointerSlice, len(ea.TBPointerSlice), cap(ea.TBPointerSlice))
}

func BenchmarkDI(b *testing.B) {
	Binding[Provider](TBProvider{})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := DI[TA](context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
