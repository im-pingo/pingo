package avframe

type Writer interface {
	Write(frame *Frame) error
}

type Reader interface {
	Read() (*Frame, error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Closer interface {
	Close() error
}

type WriteCloser interface {
	Writer
	Closer
}

type ReadCloser interface {
	Reader
	Closer
}

type ReadWriteCloser interface {
	ReadWriter
	Closer
}
