package data

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
)

// Construct the dictionary file, and then compress it using gzip:
// $ cat dictionary.txt | gzip -9 > dictionary.txt.gz
//go:embed dictionary.txt.gz
var dictionary []byte

func MustDictionary() io.Reader {
	r, err := gzip.NewReader(bytes.NewReader(dictionary))
	if err != nil {
		panic(err)
	}
	return r
}
