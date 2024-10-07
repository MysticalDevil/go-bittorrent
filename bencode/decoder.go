package bencode

import (
	"bufio"
	"io"
)

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
