package dix

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
)

var provider = make(map[string]map[string]Provider, 64)
var binding = make(map[reflect.Type]map[string]ref, 64)
var cacheTF = make(map[reflect.Type][]reflect.StructField, 1000)
var logging = false

const ctxKeyCycled = "dix::ref::cycled"

// ref is the binding information struct
type ref struct {
	t reflect.Type
	v reflect.Value
}

// Provider is dependency provider
type Provider interface {
	Symbol() string
	Provide(context.Context, *Tag) (any, error)
}

// Logging is control print log
func Logging(ok bool) { logging = ok }

// Binding is binding Provider and other type, create inject type mapping with namespaces;
// If you want to set custom values and also want to automatically inject zero values, please use exportable fields
func Binding[X any](x X, namespaces ...string) {
	if len(namespaces) == 0 {
		namespaces = append(namespaces, DefNamespace)
	}

	var i interface{} = x
	switch ix := i.(type) {
	case Provider:
		n, ok := provider[ix.Symbol()]
		if !ok {
			n = make(map[string]Provider, 8)
		}

		for _, namespace := range namespaces {
			n[namespace] = ix
		}

		provider[ix.Symbol()] = n
		printProvider(ix, namespaces...)
	default:
		e := reflect.TypeOf(&x).Elem()
		t := reflect.TypeOf(x)
		v := reflect.ValueOf(x)

		// if struct must convert to addressable reflect.Value
		if t.Kind() == reflect.Struct {
			if e.Kind() == reflect.Interface {
				n := reflect.New(t).Elem()
				c := n.NumField()
				for i := 0; i < c; i++ {
					if f := n.Field(i); f.CanSet() {
						f.Set(v.Field(i))
					}
				}
				v = n
			} else {
				v = reflect.ValueOf(&x).Elem()
			}
		}

		// binding write with namespaces
		n, ok := binding[e]
		if !ok {
			n = make(map[string]ref, 8)
		}
		for _, namespace := range namespaces {
			n[namespace] = ref{t: t, v: v}
		}

		binding[e] = n
		printBinding(e, v, namespaces...)
	}
}

// MustDI is DI wrapped, if happen error will panic
func MustDI[X any](ctx context.Context) X {
	x, e := DI[X](ctx)
	if e != nil {
		panic(e)
	}
	return x
}

// DI is dependency injection method,
// Cannot dependency on oneself, otherwise circular dependency will occur
func DI[X any](ctx context.Context) (x X, e error) {
	tag := NewTag().SetSymbol(TagInvoke)
	defer tag.Free()

	t := reflect.TypeOf(x)
	v, e := di(ctx, t, tag)
	if e == nil {
		x = v.Interface().(X)
	}

	return x, e
}

// di is dependency injection method
func di(ctx context.Context, t reflect.Type, tag *Tag) (v reflect.Value, e error) {
	// try provide
	if v, e = provide(ctx, tag, t); e != nil || v.IsValid() {
		return
	}

	// try invoke
	if v, e = invoke(ctx, tag, t); e != nil || !v.IsValid() {
		return
	}

	// pointer kind need to invoke and set layer by layer
	//p := v
	//for p.Kind() == reflect.Pointer {
	//	// invoke the pointer real type
	//	x := invoke(ctx, tag, p.Elem().Type())
	//	// the struct kind inject the value of the field
	//	if x.Kind() == reflect.Struct {
	//		if e = inject(ctx, tag, x); e != nil {
	//			return v, e
	//		}
	//	}
	//	// set the real value to root type
	//	p.Elem().Set(x)
	//	// reassign of p
	//	p = x
	//}

	// the previous code block can be optimized to this
	x := v
	for x.Kind() == reflect.Pointer {
		x = x.Elem()
	}

	// struct kind need to inject the value of the field
	if x.Kind() == reflect.Struct {
		return v, inject(ctx, tag, x)
	}

	return v, e
}

// namespace get the namespace tag, if notfound use the default namespace
func namespace(ctx context.Context, tag *Tag) (namespace string) {
	if namespace = tag.GetNamespace(); namespace == "" {
		namespace = DefNamespace
	}
	return
}

// provide is take a Provider with Provider’Symbol and call Provider’Provide method
func provide(ctx context.Context, tag *Tag, t reflect.Type) (v reflect.Value, e error) {
	if symbol := tag.GetSymbol(); len(symbol) > 0 {
		if namespaces, ok := provider[symbol]; ok {
			if provider, ok := namespaces[namespace(ctx, tag)]; ok && provider != nil {
				switch x, e := provider.Provide(ctx, tag); {
				case e != nil:
					return v, e
				case x:
					return reflect.Zero(t), nil
				default:
					return reflect.ValueOf(x), nil
				}
			}
		}
	}
	return v, e
}

// invoke is invoked with reflect
func invoke(ctx context.Context, tag *Tag, t reflect.Type) (v reflect.Value, e error) {
	if tag.GetSymbol() != TagInvoke {
		return
	}

	if namespaces, ok := binding[t]; ok {
		if bind, ok := namespaces[namespace(ctx, tag)]; ok {
			return bind.v, nil
		}
	}

	switch t.Kind() {
	case reflect.Invalid:
		return
	case reflect.Uintptr,
		reflect.UnsafePointer,
		reflect.Func,
		reflect.Bool,
		reflect.String,
		reflect.Interface,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return reflect.Zero(t), nil
	case reflect.Chan:
		return reflect.MakeChan(t, tag.GetChanBuf()), nil
	case reflect.Map:
		return reflect.MakeMapWithSize(t, tag.GetMapSize()), nil
	case reflect.Slice:
		return reflect.MakeSlice(t, tag.GetSliceLen(), tag.GetSliceCap()), nil
	case reflect.Array:
		return reflect.New(t).Elem(), nil
	case reflect.Struct:
		return reflect.New(t).Elem(), nil
	case reflect.Pointer:
		return reflect.New(t.Elem()), nil
	default:
		return reflect.Zero(t), nil
	}
}

// inject is injecting instantiated values into fields
func inject(ctx context.Context, tag *Tag, v reflect.Value) (e error) {
	t := v.Type()

	// check is cycled dependency
	if ctx, e = cycled(ctx, t); e != nil {
		return
	}

	// build type fields cache
	sfs, ok := cacheTF[t]
	if !ok {
		nf := v.NumField()
		sfs = make([]reflect.StructField, nf)
		for i := 0; i < nf; i++ {
			sfs[i] = t.Field(i) // many alloc, so need cache
		}
		cacheTF[t] = sfs
	}

	// inject from fields cache
	for i, sf := range sfs {
		val, ok := sf.Tag.Lookup(TagDix)
		if !ok {
			continue
		}

		if vf := v.Field(i); vf.CanSet() && vf.IsZero() {
			switch x, e := di(ctx, vf.Type(), tag.Reset().Unmarshal(val)); {
			case e != nil:
				return fmt.Errorf("`%s` field `%s %s` di error: %w \n", t, sf.Name, sf.Type, e)
			case x.IsValid():
				printInject(t, x, &sf)
				vf.Set(x)
			}
		}
	}

	return nil
}

// cycled is checked the cycled dependency
func cycled(ctx context.Context, t reflect.Type) (context.Context, error) {
	if len(t.PkgPath()) > 0 {
		np := strings.Join([]string{t.PkgPath(), t.Name()}, ".")
		cv := ctx.Value(ctxKeyCycled)
		if cv == nil {
			ctx = context.WithValue(ctx, ctxKeyCycled, np)
			return ctx, nil
		}

		pp := cv.(string)
		if strings.Contains(pp, np) {
			return ctx, fmt.Errorf("cycled dependency(%s)", np)
		}

		np = strings.Join([]string{pp, np}, ",")
		ctx = context.WithValue(ctx, ctxKeyCycled, np)
	}

	return ctx, nil
}

// printProvider will be print Provider information
func printProvider(x Provider, namespaces ...string) {
	if logging {
		log.Printf("[dix] `dix.Provider` binding `%#v` using %s\n", x, namespaces)
	}
}

// printBinding will be print Binding information
func printBinding(t reflect.Type, v reflect.Value, namespaces ...string) {
	if logging {
		log.Printf("[dix] `%s` binding `%#v` using %s\n", t, v, namespaces)
	}
}

// printInject will be print struct fields inject information
func printInject(t reflect.Type, v reflect.Value, sf *reflect.StructField) {
	if logging {
		log.Printf("[dix] `%s` field `%s %s` di -> %#v\n", t, sf.Name, sf.Type, v)
	}
}
