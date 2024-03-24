package dix

import (
	"strconv"
	"sync"
)

var poolTag = &sync.Pool{New: func() any { return &Tag{x: make(map[string]string, 10)} }}

// Customizable Tags
var (
	// TagDix the di struct tag
	TagDix = "dix"
	// TagChanBuf set invoke chan buf
	TagChanBuf = "chan_buf"
	// TagMapSize set invoke map size
	TagMapSize = "map_size"
	// TagSliceLen set invoke slice len
	TagSliceLen = "slice_len"
	// TagSliceCap set invoke slice cap
	TagSliceCap = "slice_cap"
	// TagSymbol set di provider Tag, Is Provider’s method Symbol
	TagSymbol = "from"
	// TagNamespace set the di working space
	TagNamespace = "namespace"
	// TagInvoke invoke type and set zero value
	TagInvoke = "?"
)

// Customizable Default Variables
var (
	// DefNamespace is the namespace tag default value
	DefNamespace = "def"
)

type Tag struct {
	// namespace is di working space
	namespace string
	// symbol is Provider’s method Symbol
	symbol string
	// buff is channel buffer
	buff int
	// size is map size
	size int
	// len is slice len
	len int
	// cap is slice cap
	cap int
	// x is custom Tags
	x map[string]string
}

func NewTag(tag ...string) *Tag {
	tag_ := poolTag.Get().(*Tag)
	return tag_.Unmarshal(tag...)
}

func (tag *Tag) Free() {
	tag.Reset()
	poolTag.Put(tag)
}

func (tag *Tag) Reset() *Tag {
	tag.namespace = ""
	tag.symbol = ""
	tag.buff = 0
	tag.size = 0
	tag.len = 0
	tag.cap = 0
	for k := range tag.x {
		delete(tag.x, k)
	}
	return tag
}

func (tag *Tag) Unmarshal(x ...string) *Tag {
	for _, f := range x {
		var k, v string
		var idx int
		for i := 0; i < len(f); i++ {
			if f[i] == ':' {
				k = f[idx:i]
				idx = i + 1
				for x := idx; x < len(f); x++ {
					if f[x] == ';' {
						v = f[idx:x]
						tag.set(k, v)
						k, v = "", ""
						idx = x + 1
						i = x
						break
					} else if x == len(f)-1 {
						v = f[idx:]
						tag.set(k, v)
						i = x
						break
					}
				}
			}
		}
	}
	return tag
}

func (tag *Tag) set(k, v string) {
	switch k {
	case TagNamespace:
		tag.SetNamespace(v)
	case TagSymbol:
		tag.SetSymbol(v)
	case TagChanBuf:
		v_, _ := strconv.Atoi(v)
		tag.SetChanBuf(v_)
	case TagMapSize:
		v_, _ := strconv.Atoi(v)
		tag.SetMapSize(v_)
	case TagSliceLen:
		v_, _ := strconv.Atoi(v)
		tag.SetSliceLen(v_)
	case TagSliceCap:
		v_, _ := strconv.Atoi(v)
		tag.SetSliceCap(v_)
	default:
		tag.SetCustomize(k, v)
	}
}

func (tag *Tag) SetNamespace(x string) *Tag {
	tag.namespace = x
	return tag
}

func (tag *Tag) SetSymbol(x string) *Tag {
	tag.symbol = x
	return tag
}

func (tag *Tag) SetChanBuf(x int) *Tag {
	tag.buff = x
	return tag
}

func (tag *Tag) SetMapSize(x int) *Tag {
	tag.size = x
	return tag
}

func (tag *Tag) SetSliceLen(x int) *Tag {
	tag.len = x
	return tag
}

func (tag *Tag) SetSliceCap(x int) *Tag {
	tag.cap = x
	return tag
}

func (tag *Tag) SetCustomize(k, v string) *Tag {
	tag.x[k] = v
	return tag
}

func (tag *Tag) GetNamespace() string {
	return tag.namespace
}

func (tag *Tag) GetSymbol() string {
	return tag.symbol
}

func (tag *Tag) GetChanBuf() int {
	return tag.buff
}

func (tag *Tag) GetMapSize() int {
	return tag.size
}

func (tag *Tag) GetSliceLen() int {
	return tag.len
}

func (tag *Tag) GetSliceCap() int {
	return tag.cap
}

func (tag *Tag) GetCustomize(k string) (v string, ok bool) {
	v, ok = tag.x[k]
	return
}
