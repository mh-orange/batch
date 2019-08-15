package trackablehttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/mh-orange/batch"
)

func testServer(paths map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if content, found := paths[r.URL.Path]; found {
			fmt.Fprint(w, content)
		} else {
			http.NotFound(w, r)
		}
	}))
}

func testProgress(want string, progress *batch.ProgressValues) func(t *testing.T) {
	return func(t *testing.T) {
		if progress.Total != int64(len(want)) {
			t.Errorf("Wanted progress total to be %d got %d", len(want), progress.Total)
		}

		if progress.Current != progress.Total {
			t.Errorf("Wanted current to be %d got %d", progress.Total, progress.Current)
		}
	}
}

func TestGet(t *testing.T) {
	want := "Content to be served"
	ts := testServer(map[string]string{"/": want})
	defer ts.Close()

	progress := &batch.ProgressValues{}
	url := url.URL{}
	url.UnmarshalBinary([]byte(ts.URL))
	reader, err := Get(progress, url)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	content, _ := ioutil.ReadAll(reader)
	got := string(content)

	if want != got {
		t.Errorf("Wanted content to be %q got %q", want, got)
	}
	t.Run("progress", testProgress(want, progress))
}

func TestGetFile(t *testing.T) {
	want := "Content to be served"
	ts := testServer(map[string]string{"/": want})
	defer ts.Close()

	progress := &batch.ProgressValues{}
	url := url.URL{}
	url.UnmarshalBinary([]byte(ts.URL))
	filename := filepath.Join(os.TempDir(), "testGetFile.txt")
	err := GetFile(progress, url, filename)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	defer os.Remove(filename)

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	got := string(content)

	if want != got {
		t.Errorf("Wanted content to be %q got %q", want, got)
	}
	t.Run("progress", testProgress(want, progress))
}

func TestGetList(t *testing.T) {
	tests := []struct {
		name      string
		content   map[string]string
		inputUrls []string
		want      string
		wantErr   error
	}{
		{"works", map[string]string{"/path1": "foo", "/path2": "bar", "/path3": "BOO!"}, []string{"/path1", "/path2", "/path3"}, "foobarBOO!", nil},
		{"bad url", map[string]string{"/path1": "foo", "/path2": "bar", "/path3": "BOO!"}, []string{"/path1", "/path2", "/path4"}, "", Error(404)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := testServer(test.content)
			defer ts.Close()

			urls := []url.URL{}
			for _, path := range test.inputUrls {
				if path == "" {
					continue
				}
				u := url.URL{}
				u.UnmarshalBinary([]byte(ts.URL))
				u.Path = path
				urls = append(urls, u)
			}

			progress := &batch.ProgressValues{}
			reader, gotErr := GetList(progress, urls)
			if gotErr == test.wantErr {
				if gotErr == nil {
					content, err := ioutil.ReadAll(reader)
					if err == nil {
						got := string(content)
						if test.want != got {
							t.Errorf("Wanted %q got %q", test.want, got)
						}

						wantSize := int64(len(test.want))
						if wantSize != progress.Total {
							t.Errorf("Wanted total size %d got %d", wantSize, progress.Total)
						}
					} else {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			} else {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}
		})
	}
}
