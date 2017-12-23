package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"

	"github.com/golang/protobuf/proto"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

func main() {
	e := flag.Bool("e", false, "encode the string back to b64 protobuf")
	flag.Parse()

	var input string
	fmt.Scanln(&input)

	if *e {
		var r rspb.Release
		// multi-line input
		err := yaml.Unmarshal([]byte(input), &r)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("%v", r)
		// s, _ := encodeRelease(input)
		// fmt.Print(s)
	} else {
		r, _ := decodeRelease(input)
		d, err := yaml.Marshal(&r)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("%s", string(d))
		// fmt.Print(d.Manifest)
	}
}

// everything below taken directly from helm/pkg/storage/driver/util.go
// https://github.com/kubernetes/helm/blob/master/pkg/storage/driver/util.go

var b64 = base64.StdEncoding
var magicGzip = []byte{0x1f, 0x8b, 0x08}

// decodeRelease decodes the bytes in data into a release
// type. Data must contain a base64 encoded string of a
// valid protobuf encoding of a release, otherwise
// an error is returned.
func decodeRelease(data string) (*rspb.Release, error) {
	// base64 decode string
	b, err := b64.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls rspb.Release
	// unmarshal protobuf bytes
	if err := proto.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

// encodeRelease encodes a release returning a base64 encoded
// gzipped binary protobuf encoding representation, or error.
func encodeRelease(rls *rspb.Release) (string, error) {
	b, err := proto.Marshal(rls)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err = w.Write(b); err != nil {
		return "", err
	}
	w.Close()

	return b64.EncodeToString(buf.Bytes()), nil
}
