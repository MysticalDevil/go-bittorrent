package bencode

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrNum         = errors.New("expect num")
	ErrColon       = errors.New("expect colon")
	ErrExpCharI    = errors.New("expect char i")
	ErrExpCharE    = errors.New("expect char e")
	ErrType        = errors.New("wrong type")
	ErrWriteFailed = errors.New("write failed")
	ErrReadFailed  = errors.New("read failed")
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

func EncodeString(w io.Writer, val string) (int, error) {
	strLen := len(val)
	bw := bufio.NewWriter(w)
	wLen, err := writeDecimal(bw, strLen)
	if err != nil {
		return 0, err
	}
	err = bw.WriteByte(':')
	if err != nil {
		return 0, ErrWriteFailed
	}
	wLen++
	_, err = bw.WriteString(val)
	if err != nil {
		return 0, ErrWriteFailed
	}
	wLen += strLen

	err = bw.Flush()
	if err != nil {
		return 0, ErrWriteFailed
	}

	return wLen, nil
}

func DecodeString(r io.Reader) (val string, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	num, len_ := readDecimal(br)
	if len_ == 0 {
		return val, ErrNum
	}
	b, err := br.ReadByte()
	if err != nil {
		return "", err
	}
	if b != ':' {
		return val, ErrColon
	}

	buf := make([]byte, num)
	_, err = io.ReadAtLeast(br, buf, num)
	if err != nil {
		return "", err
	}
	val = string(buf)
	return
}

func EncodeInt(w io.Writer, val int) (int, error) {
	bw := bufio.NewWriter(w)
	wLen := 0
	err := bw.WriteByte('i')
	if err != nil {
		return 0, err
	}
	wLen++
	nLen, err := writeDecimal(bw, val)
	if err != nil {
		return 0, err
	}
	wLen += nLen
	err = bw.WriteByte('e')
	if err != nil {
		return 0, ErrWriteFailed
	}
	wLen++

	err = bw.Flush()
	if err != nil {
		return 0, ErrWriteFailed
	}

	return wLen, nil
}

func DecodeInt(r io.Reader) (val int, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.ReadByte()
	if err != nil {
		return 0, ErrReadFailed
	}
	if b != 'i' {
		return val, ErrExpCharI
	}
	val, _ = readDecimal(br)
	b, err = br.ReadByte()
	if b != 'e' {
		return val, ErrExpCharE
	}
	return
}

func checkNum(data byte) bool {
	return data >= '0' && data <= '9'
}

func readDecimal(r *bufio.Reader) (val int, len int) {
	sign := 1
	b, err := r.ReadByte()
	if err != nil {
		return 0, 0
	}
	len++
	if b == '-' {
		sign = -1
		b, err = r.ReadByte()
		if err != nil {
			return 0, 0
		}
		len++
	}
	for {
		if !checkNum(b) {
			err := r.UnreadByte()
			if err != nil {
				return 0, 0
			}
			len--
			return sign * val, len
		}
		val = val*10 + int(b-'0')
		b, err = r.ReadByte()
		if err != nil {
			return 0, 0
		}
		len++
	}
}

func writeDecimal(w *bufio.Writer, val int) (len int, err error) {
	if val == 0 {
		err = w.WriteByte('0')
		if err != nil {
			return 0, ErrWriteFailed
		}
		len++
		return
	}
	if val < 0 {
		err = w.WriteByte('-')
		if err != nil {
			return 0, ErrWriteFailed
		}
		len++
		val = -val
	}

	dividend := 1
	for {
		if dividend > val {
			dividend /= 10
			break
		}
		dividend *= 10
	}
	for {
		num := byte(val / dividend)
		err = w.WriteByte('0' + num)
		if err != nil {
			return 0, ErrWriteFailed
		}
		len++
		if dividend == 1 {
			return
		}
		val %= dividend
		dividend /= 10
	}
}
