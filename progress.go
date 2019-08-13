package batch

import (
	"io"
)

// Progress is anything which can monitor progress (such as
// a progress bar)
type Progress interface {
	Set(int64)
	Add(int64)
	Update(int64)
	Finish()
}

// ProgressValues is a type that implements Progress and simply
// stores the current and total values
type ProgressValues struct {
	Total   int64 // Total expected value for the progress
	Current int64 // Current value for the progress
}

// Set the total value for the progress
func (pv *ProgressValues) Set(total int64) { pv.Total = total }

// Add an amount to the current value
func (pv *ProgressValues) Add(addend int64) { pv.Current += addend }

// Update the current value
func (pv *ProgressValues) Update(current int64) { pv.Current = current }

// Finish does nothing since nothing is in the background.  Finish shouldn't
// update the current value, since a job might finish with a failure at
// some point in the middle
func (pv *ProgressValues) Finish() {}

type proxyReader struct {
	io.Reader
	progress Progress
}

// ProxyReader returns an io.Reader that will update
// progress as it reads
func ProxyReader(p Progress, reader io.Reader) io.ReadCloser {
	return &proxyReader{reader, p}
}

func (pr *proxyReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.progress.Add(int64(n))
	}

	if err == io.EOF {
		pr.progress.Finish()
	}
	return
}

func (pr *proxyReader) Close() error {
	if closer, ok := pr.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
