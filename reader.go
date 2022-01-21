package mimedropreader

import (
	"errors"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

var ErrMIMENotAllowed = errors.New("MIME not allowed")

var _ io.Reader = (*Reader)(nil)

type Reader struct {
	r       io.Reader
	mime    *mimetype.MIME
	allowed []string
}

func New(r io.Reader, mime1 string, mimeN ...string) *Reader {
	if r == nil {
		panic("r is nil")
	}
	return &Reader{
		r:       r,
		allowed: append([]string{mime1}, mimeN...),
	}
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if r.mime != nil {
		return n, err
	}

	if n == 0 {
		return n, err
	}

	mime := mimetype.Detect(p[:n])
	r.mime = mime
	if err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if !mimetype.EqualsAny(mime.String(), r.allowed...) {
		return n, ErrMIMENotAllowed
	}

	return n, err
}

func (r *Reader) MIME() *mimetype.MIME {
	return r.mime
}

func (r *Reader) Unwrap() io.Reader {
	return r.r
}

var _ io.ReadCloser = (*ReadCloser)(nil)

type ReadCloser struct {
	*Reader
}

func NewReadCloser(r io.Reader, mime1 string, mimeN ...string) *ReadCloser {
	return &ReadCloser{New(r, mime1, mimeN...)}
}

func (rc *ReadCloser) Close() error {
	closer, _ := rc.Reader.r.(io.Closer)
	if closer != nil {
		return closer.Close()
	}
	return nil
}
