package mimedropreader_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/dinalt/mimedropreader"
)

func TestReader_Read(t *testing.T) {
	bts, err := os.ReadFile(filepath.Join("testdata", "noise.jpeg"))
	if err != nil {
		panic(err)
	}

	t.Run("should match", func(t *testing.T) {
		br := bytes.NewReader(bts)

		reader := mimedropreader.New(br, "image/jpeg", "image/png")

		res, err := io.ReadAll(reader)
		if err != nil {
			t.Error(err)
		}
		if len(res) != int(br.Size()) {
			t.Errorf("input size != output size (%d != %d)", br.Size(), len(res))
		}
	})

	t.Run("should not match", func(t *testing.T) {
		br := bytes.NewReader(bts)

		reader := mimedropreader.New(br, "image/*")

		res, err := io.ReadAll(reader)
		if !errors.Is(err, mimedropreader.ErrMIMENotAllowed) {
			t.Errorf("want %+v, got: %+v", mimedropreader.ErrMIMENotAllowed, err)
		}
		if len(res) == 0 {
			t.Error("no data returned from io.ReadAll")
		}
	})
}

func TestReadCloser_Close(t *testing.T) {
	t.Run("should close file", func(t *testing.T) {
		f, err := os.Open(filepath.Join("testdata", "mono.jpeg"))
		if err != nil {
			panic(err)
		}

		rc := mimedropreader.NewReadCloser(f, "image/jpeg")
		res, err := io.ReadAll(rc)
		if err != nil {
			panic(err)
		}
		if len(res) == 0 {
			t.Error("no data was read")
		}
		err = rc.Close()
		if err != nil {
			t.Error(err)
		}
		err = f.Close()
		if err == nil {
			t.Error("file was not closed by ReadCloser")
		}
	})
}
