package bencode

import "errors"

var (
	ErrNum         = errors.New("expect num")
	ErrColon       = errors.New("expect colon")
	ErrExpCharI    = errors.New("expect char i")
	ErrExpCharE    = errors.New("expect char e")
	ErrType        = errors.New("wrong type")
	ErrWriteFailed = errors.New("write failed")
	ErrReadFailed  = errors.New("read failed")
)
