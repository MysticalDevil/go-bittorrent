package bencode

import (
	"bufio"
	"io"
)

func Parse(r io.Reader) (*BObject, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	var ret BObject
	switch {
	case b[0] >= '0' && b[0] <= '9':
		// parse string
		val, err := DecodeString(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BSTR
		ret.val_ = val
	case b[0] == 'i':
		// parse int
		val, err := DecodeInt(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BINT
		ret.val_ = val
	case b[0] == 'l':
		_, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		var list []*BObject
		for {
			p, err := br.Peek(1)
			if err != nil {
				return nil, err
			}
			if p[0] == 'e' {
				_, err := br.ReadByte()
				if err != nil {
					return nil, err
				}
				break
			}
			elem, err := Parse(br)
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		ret.type_ = BLIST
		ret.val_ = list

	case b[0] == 'd':
		_, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		dict := make(map[string]*BObject)
		for {
			p, err := br.Peek(1)
			if err != nil {
				return nil, err
			}
			if p[0] == 'e' {
				_, err := br.ReadByte()
				if err != nil {
					return nil, err
				}
				break
			}
			key, err := DecodeString(br)
			if err != nil {
				return nil, err
			}
			val, err := Parse(br)
			if err != nil {
				return nil, err
			}
			dict[key] = val
			ret.type_ = BDICT
			ret.val_ = dict
		}
	default:
		return nil, ErrType
	}
	return &ret, nil
}
