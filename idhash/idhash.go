package idhash

import (
	"fmt"
	"log"

	hashids "github.com/speps/go-hashids"
)

var hasher *hashids.HashID

func InitHasher(salt string, minLength int) (err error) {
	hd := hashids.NewData()
	hd.Salt = salt
	if minLength > 0 {
		hd.MinLength = minLength
	}

	hasher, err = hashids.NewWithData(hd)

	return err
}

func Encode(typ string, id int64) string {
	out, err := hasher.EncodeInt64([]int64{typeToInt(typ), id})
	if err != nil {
		log.Fatal(err)
	}

	return out
}

func Decode(in string) (string, int64, error) {
	out, err := hasher.DecodeInt64WithError(in)
	if err != nil {
		return "", 0, err
	}

	if len(out) != 2 {
		return "", 0, fmt.Errorf("Expected 2 output integers, got %#v", out)
	}

	return intToType(out[0]), out[1], nil
}
