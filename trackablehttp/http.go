package trackablehttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/mh-orange/batch"
)

// Error is returned from DownloadList or GetTotalSize when one of the
// requests results in something other than a 200 code
type Error int

// Error will report the HTTP status code returned from the server as
// well as its status text
func (he Error) Error() string {
	return fmt.Sprintf("Received HTTP code %d: %v", int(he), http.StatusText(int(he)))
}

// Get will perform an HTTP GET request and return a reader attached
// to the http.Response.Body.  The reader will automatically update the
// progress object as it is read from
func Get(pb batch.Progress, url url.URL) (reader io.ReadCloser, err error) {
	resp, err := http.Get(url.String())
	if err == nil {
		if resp.ContentLength >= 0 {
			pb.Set(resp.ContentLength)
		}
		reader = batch.ProxyReader(pb, resp.Body)
	}
	return reader, err
}

// GetFile will perform an HTTP GET for the URL and write the response
// to the file named "filename".  The Progess object will be updated as
// the file is downloaded
func GetFile(pb batch.Progress, url url.URL, filename string) error {
	file, err := os.Create(filename)
	if err == nil {
		var reader io.ReadCloser
		reader, err = Get(pb, url)
		if err == nil {
			_, err = io.Copy(file, reader)
			reader.Close()
		}
	}
	return err
}

// GetTotalSize will perform HEAD requests on the list of
// urls and will return the sum of the Content-Length header
// values
func GetTotalSize(urls []url.URL) (total int64, err error) {
	for _, url := range urls {
		var resp *http.Response
		resp, err = http.Head(url.String())
		if err != nil {
			break
		}
		if resp.StatusCode != http.StatusOK {
			err = Error(resp.StatusCode)
			break
		}
		size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		total += int64(size)
	}
	return
}

// GetList will take a list of urls and perform a sequence of HTTP GET
// requests then writing the bodies in sequence to the given io.Writer.
// Prior to downloading the url bodies, HEAD requests will be performed
// for each URL so that a total size can be determined and the progress
// can be properly updated
func GetList(pb batch.Progress, urls []url.URL) (reader io.ReadCloser, err error) {
	total, err := GetTotalSize(urls)
	pb.Set(total)
	lr := &listReader{progress: pb, urls: urls}
	return lr, err
}

type listReader struct {
	progress batch.Progress
	urls     []url.URL
	current  io.ReadCloser
}

func (lr *listReader) Read(p []byte) (n int, err error) {
	if lr.current == nil {
		if len(lr.urls) == 0 {
			return 0, io.EOF
		}

		resp, err := http.Get(lr.urls[0].String())
		lr.urls = lr.urls[1:]
		if err == nil {
			lr.current = batch.ProxyReader(lr.progress, resp.Body)
		} else {
			return 0, err
		}
	}

	n, err = lr.current.Read(p)
	if err == io.EOF {
		if len(lr.urls) > 0 {
			err = nil
		}
		err = lr.current.Close()
		lr.current = nil
	}
	return n, err
}

func (lr *listReader) Close() error {
	if lr.current != nil {
		return lr.current.Close()
	}
	return errors.New("Already closed")
}
