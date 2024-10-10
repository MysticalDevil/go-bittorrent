package bencode

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func objAssertStr(t *testing.T, expect string, o *BObject) {
	assert.Equal(t, BSTR, o.type_)
	str, err := o.Str()
	assert.NoError(t, err)
	assert.Equal(t, expect, str)
}

func objAssertInt(t *testing.T, expect int, o *BObject) {
	assert.Equal(t, BINT, o.type_)
	val, err := o.Int()
	assert.NoError(t, err)
	assert.Equal(t, expect, val)
}

func TestParse(t *testing.T) {
	var o *BObject
	var in string
	var buf *bytes.Buffer

	t.Run("ParseString", func(t *testing.T) {
		in = "3:abc"
		buf = bytes.NewBufferString(in)
		o, _ = Parse(buf)
		objAssertStr(t, "abc", o)
	})

	t.Run("FailedParseString", func(t *testing.T) {
		in := "e:abc"
		buf = bytes.NewBufferString(in)
		_, err := Parse(buf)
		assert.Error(t, err)
	})

	t.Run("ParseInt", func(t *testing.T) {
		in = "i123e"
		buf = bytes.NewBufferString(in)
		o, _ = Parse(buf)
		objAssertInt(t, 123, o)
	})

	t.Run("FailedParseInt", func(t *testing.T) {
		in = "i123"
		buf = bytes.NewBufferString(in)
		_, err := Parse(buf)
		assert.Error(t, err)
	})

	t.Run("ParseList", func(t *testing.T) {
		var list []*BObject
		in = "li123e6:archeri789ee"
		buf = bytes.NewBufferString(in)
		o, _ = Parse(buf)
		assert.Equal(t, BLIST, o.type_)
		list, err := o.List()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(list))
		objAssertInt(t, 123, list[0])
		objAssertStr(t, "archer", list[1])
		objAssertInt(t, 789, list[2])

		out := bytes.NewBufferString("")
		assert.Equal(t, len(in), o.Bencode(out))
		assert.Equal(t, in, out.String())
	})

	t.Run("FailedParseList", func(t *testing.T) {
		in = "li123"
		buf = bytes.NewBufferString(in)
		_, err := Parse(buf)
		assert.Error(t, err)
	})

	t.Run("ParseMap", func(t *testing.T) {
		var dict map[string]*BObject
		in = "d4:name6:archer3:agei29ee"
		buf = bytes.NewBufferString(in)
		o, _ = Parse(buf)
		assert.Equal(t, BDICT, o.type_)
		dict, err := o.Dict()
		assert.NoError(t, err)
		objAssertStr(t, "archer", dict["name"])
		objAssertInt(t, 29, dict["age"])

		out := bytes.NewBufferString("")
		assert.Equal(t, len(in), o.Bencode(out))
	})

	t.Run("FailedParseMap", func(t *testing.T) {
		in = "d4:name6:archer3:agei29e"
		buf = bytes.NewBufferString(in)
		_, err := Parse(buf)
		assert.Error(t, err)
	})

	t.Run("ParseComMap", func(t *testing.T) {
		var dict map[string]*BObject
		in = "d4:userd4:name6:archer3:agei29ee5:valueli80ei85ei90eee"
		buf = bytes.NewBufferString(in)
		o, _ = Parse(buf)
		assert.Equal(t, BDICT, o.type_)
		dict, err := o.Dict()
		assert.NoError(t, err)
		assert.Equal(t, BDICT, dict["user"].type_)
		assert.Equal(t, BLIST, dict["value"].type_)
	})
}
