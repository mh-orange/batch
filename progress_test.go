package batch

import (
	"bytes"
	"io"
	"testing"
)

func TestProxyReader(t *testing.T) {
	tests := []struct {
		name     string
		bufSize  int
		readSize int
		wantRead int
		wantErr  error
	}{
		{"full read", 1024, 256, 256, nil},
		{"eof", 0, 1024, 0, io.EOF},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(make([]byte, test.bufSize))
			pr := &ProgressValues{}
			pr.Set(int64(test.bufSize))
			if pr.Current != 0 {
				t.Errorf("Expected progress to start at zero")
			}

			pr.Update(2022)
			if pr.Current != 2022 {
				t.Errorf("Expected progress to be 2022, got %d", pr.Current)
			}
			pr.Update(0)

			proxy := ProxyReader(pr, buf)
			gotRead, gotErr := proxy.Read(make([]byte, test.readSize))
			if test.wantRead != gotRead {
				t.Errorf("Wanted to read %d bytes got %d", test.wantRead, gotRead)
			}

			if test.wantErr != gotErr {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}

			if pr.Current != int64(test.wantRead) {
				t.Errorf("Wanted progress to be %d got %d", test.wantRead, pr.Current)
			}

			gotErr = proxy.Close()
			if gotErr != nil {
				t.Errorf("Wanted nil error when closing, got %v", gotErr)
			}
		})
	}
}

type testCloser struct{ error }

func (tc testCloser) Read([]byte) (int, error) { return 0, tc.error }
func (tc testCloser) Close() error             { return tc.error }

func TestProxyReaderClose(t *testing.T) {
	wantErr := io.EOF
	proxy := ProxyReader(&ProgressValues{}, testCloser{wantErr})
	gotErr := proxy.Close()
	if wantErr != gotErr {
		t.Errorf("Wanted error %v got %v", wantErr, gotErr)
	}
}
