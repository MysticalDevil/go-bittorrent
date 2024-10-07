package bencode

import (
	"bufio"
	"io"
)

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
