package dix

import "testing"

func TestTag(t *testing.T) {
	tag := NewTag("from:?;namespace:ns1;kind:x1;slice_len:10;slice_cap:10;map_size:10;chan_buf:10")
	defer tag.Free()

	t.Log(tag.GetSymbol())
	t.Log(tag.GetNamespace())
	t.Log(tag.GetSliceLen())
	t.Log(tag.GetSliceCap())
	t.Log(tag.GetMapSize())
	t.Log(tag.GetChanBuf())
	t.Log(tag.GetCustomize("kind"))
	t.Log(tag.GetCustomize("x"))
}
