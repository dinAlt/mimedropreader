package mimedropreader_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

type reader struct {
	r       io.Reader
	mime    *mimetype.MIME
	allowed []string
}

func (r *reader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if r.mime != nil {
		return n, err
	}

	if n == 0 {
		return n, err
	}

	r.mime = mimetype.Detect(p[:n])
	if err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if !mimetype.EqualsAny(r.mime.String(), r.allowed...) {
		return n, fmt.Errorf("Bad MIME type: %q", r.mime.String())
	}

	return n, err
}

type nonBranchedReader struct {
	child   io.Reader
	mime    *mimetype.MIME
	allowed []string
}

func (r *nonBranchedReader) Read(p []byte) (int, error) {
	return r.child.Read(p)
}

type nonBranchedReaderInner struct {
	parent *nonBranchedReader
	r      io.Reader
}

func (r *nonBranchedReaderInner) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)

	if n == 0 {
		return n, err
	}
	mime := mimetype.Detect(p[:n])
	r.parent.mime = mime
	if err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if !mimetype.EqualsAny(mime.String(), r.parent.allowed...) {
		return n, fmt.Errorf("Bad MIME type: %q", mime.String())
	}

	r.parent.child = r.r
	r.r = nil
	r.parent = nil

	return n, err
}

func BenchmarkNonBranchedMatched(b *testing.B) {
	bts, err := os.ReadFile(filepath.Join(".", "testdata", "noise.jpeg"))
	if err != nil {
		panic(err)
	}
	br := bytes.NewReader(bts)
	child := &nonBranchedReaderInner{
		r: br,
	}

	r := &nonBranchedReader{
		allowed: []string{"image/jpeg"},
		child:   child,
	}
	child.parent = r

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Seek(0, io.SeekStart)
		res, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		if int(br.Size()) != len(res) {
			panic("in and out len does not match")
		}
	}
}

func BenchmarkBranchedMatched(b *testing.B) {
	bts, err := os.ReadFile(filepath.Join(".", "testdata", "noise.jpeg"))
	if err != nil {
		panic(err)
	}
	br := bytes.NewReader(bts)
	r := &reader{r: br, allowed: []string{"image/jpeg"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Seek(0, io.SeekStart)
		res, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		if int(br.Size()) != len(res) {
			panic("in and out len does not match")
		}
	}
}
