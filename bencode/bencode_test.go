package bencode

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FailingWriter struct{}

func (w *FailingWriter) Write([]byte) (n int, err error) {
	return 0, errors.New("write error")
}

type FailingReader struct{}

func (r *FailingReader) Read([]byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestObject(t *testing.T) {
	strObj := &BObject{type_: BSTR, val_: "test"}
	intObj := &BObject{type_: BINT, val_: 42}
	listObj := &BObject{type_: BLIST, val_: []*BObject{
		strObj,
		intObj,
	}} // ""
	dictObj := &BObject{type_: BDICT, val_: map[string]*BObject{"key": strObj}}

	t.Run("TestStr", func(t *testing.T) {
		val, err := strObj.Str()
		assert.NoError(t, err)
		assert.Equal(t, "test", val)

		_, err = intObj.Str()
		assert.Error(t, err)
		assert.Equal(t, ErrType, err)
	})

	t.Run("TestInt", func(t *testing.T) {
		val, err := intObj.Int()
		assert.NoError(t, err)
		assert.Equal(t, 42, val)

		_, err = strObj.Int()
		assert.Error(t, err)
		assert.Equal(t, ErrType, err)
	})

	t.Run("TestList", func(t *testing.T) {
		val, err := listObj.List()
		assert.NoError(t, err)
		assert.Equal(t, strObj, val[0])
		assert.Equal(t, intObj, val[1])

		_, err = dictObj.List()
		assert.Error(t, err)
		assert.Equal(t, ErrType, err)
	})

	t.Run("TestDict", func(t *testing.T) {
		val, err := dictObj.Dict()
		assert.NoError(t, err)
		assert.Len(t, val, 1)
		assert.Equal(t, strObj, val["key"])

		_, err = listObj.Dict()
		assert.Error(t, err)
		assert.Equal(t, ErrType, err)
	})
}

func TestEncodeString(t *testing.T) {
	buf := new(bytes.Buffer)

	t.Run("SuccessfulEncodeString", func(t *testing.T) {
		wLen, err := EncodeString(buf, "hello")
		assert.NoError(t, err)
		assert.Equal(t, 7, wLen) // "5:hello"
		assert.Equal(t, "5:hello", buf.String())
		buf.Reset()
	})

	t.Run("WriterFailsOnWriteDecimal", func(t *testing.T) {
		writer := &FailingWriter{}
		wLen, err := EncodeString(writer, "hello")
		assert.Equal(t, 0, wLen)
		assert.Error(t, err)
		assert.Equal(t, ErrWriteFailed, err)
	})

	t.Run("FailedEncodeStringOnWrite", func(t *testing.T) {
		writer := &FailingWriter{}
		wLen, err := EncodeString(writer, "hello")
		assert.Zero(t, wLen)
		assert.Error(t, err)
	})
}

func TestDecodeString(t *testing.T) {
	t.Run("SuccessfulDecodeString", func(t *testing.T) {
		input := bytes.NewBufferString("5:hello")
		val, err := DecodeString(input)
		assert.NoError(t, err)
		assert.Equal(t, "hello", val)
	})

	t.Run("FailedDecodeStringOnInvalidNumber", func(t *testing.T) {
		input := bytes.NewBufferString("x:hello")
		_, err := DecodeString(input)
		assert.Error(t, err)
		assert.Equal(t, ErrNum, err)
	})

	t.Run("FailedDecodeStringOnShortInput", func(t *testing.T) {
		input := bytes.NewBufferString("5:hel")
		_, err := DecodeString(input)
		assert.Error(t, err)
	})

	t.Run("FailedDecodeStringOnReadError", func(t *testing.T) {
		reader := &FailingReader{}
		_, err := DecodeString(reader)
		assert.Error(t, err)
	})
}

func TestEncodeInt(t *testing.T) {
	buf := new(bytes.Buffer)

	t.Run("SuccessfulEncodeInt", func(t *testing.T) {
		wLen, err := EncodeInt(buf, 123) // "i123e"
		assert.NoError(t, err)
		assert.Equal(t, 5, wLen)
		assert.Equal(t, "i123e", buf.String())
		buf.Reset()
	})

	t.Run("EncodeIntWithNegativeNumber", func(t *testing.T) {
		wLen, err := EncodeInt(buf, -123) // "i-123e"
		assert.NoError(t, err)
		assert.Equal(t, 6, wLen)
		assert.Equal(t, "i-123e", buf.String())
		buf.Reset()
	})

	t.Run("EncodeIntWithZero", func(t *testing.T) {
		wLen, err := EncodeInt(buf, 0) // "i0e"
		assert.NoError(t, err)
		assert.Equal(t, 3, wLen)
		assert.Equal(t, "i0e", buf.String())
		buf.Reset()
	})

	t.Run("FailedEncodeIntOnWriteBytes", func(t *testing.T) {
		writer := &FailingWriter{}
		wLen, err := EncodeInt(writer, 123)
		assert.Zero(t, wLen)
		assert.Error(t, err)
		assert.Equal(t, ErrWriteFailed, err)
	})
}

func TestDecodeInt(t *testing.T) {
	t.Run("SuccessfulDecodeInt", func(t *testing.T) {
		input := bytes.NewBufferString("i123e")
		val, err := DecodeInt(input)
		assert.NoError(t, err)
		assert.Equal(t, 123, val)
	})

	t.Run("DecodeIntWithNegativeNumber", func(t *testing.T) {
		input := bytes.NewBufferString("i-123e")
		val, err := DecodeInt(input)
		assert.NoError(t, err)
		assert.Equal(t, -123, val)
	})

	t.Run("DecodeIntWithZero", func(t *testing.T) {
		input := bytes.NewBufferString("i0e")
		val, err := DecodeInt(input)
		assert.NoError(t, err)
		assert.Equal(t, 0, val)
	})

	t.Run("FailedDecodeIntOnMissingChatI", func(t *testing.T) {
		input := bytes.NewBufferString("123e")
		_, err := DecodeInt(input)
		assert.Error(t, err)
		assert.Equal(t, ErrExpCharI, err)
	})

	t.Run("FailedDecodeIntOnMissingChatE", func(t *testing.T) {
		input := bytes.NewBufferString("i123")
		_, err := DecodeInt(input)
		assert.Error(t, err)
		assert.Equal(t, ErrExpCharE, err)
	})

	t.Run("FailedDecodeIntOnInvalidInput", func(t *testing.T) {
		reader := &FailingReader{}
		_, err := DecodeInt(reader)
		assert.Error(t, err)
		assert.Equal(t, ErrReadFailed, err)
	})
}

func TestBencode(t *testing.T) {
	testCases := []struct {
		name      string
		input     *BObject
		wantError error
		wantLen   int
	}{
		{
			name:      "empty string",
			input:     &BObject{type_: BSTR, val_: ""},
			wantError: nil,
			wantLen:   2, // "0:"
		},
		{
			name:      "string",
			input:     &BObject{type_: BSTR, val_: "Hello, World!"},
			wantError: nil,
			wantLen:   16, // "13:Hello, World!"
		},
		{
			name:      "empty list",
			input:     &BObject{type_: BLIST, val_: []*BObject{}},
			wantError: nil,
			wantLen:   2, // "le"
		},
		{
			name: "list",
			input: &BObject{
				type_: BLIST,
				val_: []*BObject{
					{type_: BSTR, val_: "hello"},
					{type_: BINT, val_: 123},
				},
			},
			wantError: nil,
			wantLen:   14, // "l5:helloi123ee"
		},
		{
			name: "empty dict",
			input: &BObject{
				type_: BDICT,
				val_:  map[string]*BObject{},
			},
			wantError: nil,
			wantLen:   2, // "de"
		},
		{
			name: "dict",
			input: &BObject{
				type_: BDICT,
				val_: map[string]*BObject{
					"hello": {type_: BSTR, val_: "world"},
					"num":   {type_: BINT, val_: 123},
				},
			},
			wantError: nil,
			wantLen:   26, // "d5:hello5:world3:numi123ee"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bb := &bytes.Buffer{}
			gotLen := tc.input.Bencode(bb)
			if gotLen != tc.wantLen {
				t.Errorf("Bencode() = %d, want %d", gotLen, tc.wantLen)
			}
		})
	}
}
