package hydrus

import (
	"fmt"
	"io"

	"github.com/oxtoacart/emsort"
)

// externalSorter sorts constant size []byte externally on the file system
type externalSorter struct {
	chunkSize int
	commit    func([]byte) error
	writeBuf  []byte
}

// NewSorter creates a new external []byte sorter
func NewSorter(chunkSize int, commit func([]byte) error) (
	emsort.SortedWriter, error,
) {
	e := externalSorter{
		chunkSize: chunkSize,
		commit:    commit,
	}
	return emsort.New(&e, e.chunk, e.less, 1<<29)
}

// Splits input stream into []byte chunks of chunkSize
func (e externalSorter) chunk(r io.Reader) ([]byte, error) {
	buf := make([]byte, e.chunkSize)
	read, err := r.Read(buf)
	switch {
	case err != nil:
		return nil, err
	case read != e.chunkSize:
		return nil, fmt.Errorf("incomplete input size: %d", read)
	default:
		return buf, nil
	}
}

// Write chunks the sorted stream and passes it to commit
func (e externalSorter) Write(buf []byte) (written int, err error) {
	// Consume previously buffered chunk part, if any
	if e.writeBuf != nil {
		i := e.chunkSize - len(e.writeBuf)
		err = e.commit(append(e.writeBuf, buf[:i]...))
		if err != nil {
			return
		}
		written += i
		buf = buf[i:]
	}

	for len(buf) >= e.chunkSize {
		err = e.commit(buf[:e.chunkSize])
		if err != nil {
			return
		}
		written += e.chunkSize
		buf = buf[e.chunkSize:]
	}

	// Retain incomplete chunk till next write call
	e.writeBuf = nil
	if len(buf) != 0 {
		e.writeBuf = make([]byte, len(buf))
		copy(e.writeBuf, buf)
	}
	return
}

func (e externalSorter) less(a, b []byte) bool {
	for i := 0; i < e.chunkSize; i++ {
		switch {
		case a[i] < b[i]:
			return true
		case a[i] > b[i]:
			return false
		}
	}
	return false
}
