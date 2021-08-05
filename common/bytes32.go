package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
)

type Bytes8 = [8]byte
type Bytes32 = [32]byte
type Bytes16 = [16]byte

var ErrStringTooLong = errors.New("string to long")

func B8(s string) (r Bytes8) {
	copy(r[:], s)
	return
}

func ToString8(b Bytes8) string {
	return string(bytes.Trim(b[:], "\000"))
}

func B16(s string) (r Bytes16) {
	copy(r[:], s)
	return
}

func ToBytes32(s string) (r Bytes32, err error) {
	if len(s) > 32 {
		err = ErrStringTooLong
	} else {
		copy(r[:], s)
	}
	return
}

func B32(s string) (r Bytes32) {
	copy(r[:], s)
	return
}

func ToString(b Bytes32) string {
	return string(bytes.Trim(b[:], "\000"))
}

func ToHex(b Bytes32) string {
	return "0x" + hex.EncodeToString(b[:])
}

func BytesToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b[:])
}

var ErrHexTooShort = errors.New("hex string is shorter than bytes32")

func FromHex(s string) (Bytes32, error) {
	var b Bytes32
	n, err := hex.Decode(b[:], []byte(strings.TrimPrefix(s, "0x")))
	if err == nil && n != 32 {
		return b, ErrHexTooShort
	}
	return b, err
}

func Bytes32Enum(in interface{}) interface{} {
	out := reflect.ValueOf(in).Elem()
	for i := out.NumField() - 1; i >= 0; i-- {
		fieldValue := B32(out.Type().Field(i).Name)
		out.Field(i).Set(reflect.ValueOf(fieldValue))
	}
	return out.Interface()
}

func BytesToBytes32(b []byte) (r Bytes32, err error) {
	if len(b) > 32 {
		err = ErrStringTooLong
	} else {
		copy(r[:], b)
	}
	return
}
