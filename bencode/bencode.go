package bencode

import (
	"bufio"
	"io"
)

type BType uint8

const (
	BSTR  BType = 0x01
	BINT  BType = 0x02
	BLIST BType = 0x03
	BDICT BType = 0x04
)

type BValue any

type BObject struct {
	type_ BType
	val_  BValue
}

func (o *BObject) Str() (string, error) {
	if o.type_ != BSTR {
		return "", ErrType
	}
	return o.val_.(string), nil
}

func (o *BObject) Int() (int, error) {
	if o.type_ != BINT {
		return 0, ErrType
	}
	return o.val_.(int), nil
}

func (o *BObject) List() ([]*BObject, error) {
	if o.type_ != BLIST {
		return nil, ErrType
	}
	return o.val_.([]*BObject), nil
}

func (o *BObject) Dict() (map[string]*BObject, error) {
	if o.type_ != BDICT {
		return nil, ErrType
	}
	return o.val_.(map[string]*BObject), nil
}

func (o *BObject) Bencode(w io.Writer) int {
	bw, ok := w.(*bufio.Writer)
	var err error
	nLen := 0
	if !ok {
		bw = bufio.NewWriter(w)
	}
	wLen := 0
	switch o.type_ {
	case BSTR:
		str, _ := o.Str()
		nLen, err = EncodeString(bw, str)
		wLen += nLen
	case BINT:
		val, _ := o.Int()
		nLen, err = EncodeInt(bw, val)
		wLen += nLen
	case BLIST:
		err = bw.WriteByte('l')
		list, _ := o.List()
		for _, elem := range list {
			wLen += elem.Bencode(bw)
		}
		_ = bw.WriteByte('e')
		wLen += 2
	case BDICT:
		err = bw.WriteByte('d')
		dict, _ := o.Dict()
		for k, v := range dict {
			nLen, err = EncodeString(bw, k)
			wLen += nLen
			wLen += v.Bencode(bw)
		}
		err = bw.WriteByte('e')
		wLen += 2
	}
	err = bw.Flush()
	if err != nil {
		return 0
	}
	return wLen
}
