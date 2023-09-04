package dep

import (
	"bytes"
	"context"
	"depviz/internal/services/dep/test_utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"
)

func Test_parsePackageDeps(t *testing.T) {
	t.Run("test parsing valid json", func(t *testing.T) {
		data := []byte(`{
			"info": {
				"requires_dist": ["a", "b", "c"]
			}
		}`)
		reader := bytes.NewReader(data)

		deps, err := parsePackageDeps(reader)
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, deps)
	})

	t.Run("test parsing json if requires_dist is null", func(t *testing.T) {
		data := []byte(`{
			"info": {
				"requires_dist": null
			}
		}`)
		reader := bytes.NewReader(data)

		deps, err := parsePackageDeps(reader)
		assert.NoError(t, err)
		assert.Empty(t, deps)
	})

	t.Run("test correct error if json is invalid", func(t *testing.T) {
		data := []byte(`{_}`)
		reader := bytes.NewReader(data)

		deps, err := parsePackageDeps(reader)
		assert.ErrorIs(t, err, ErrInvalidJson)
		assert.Empty(t, deps)
	})

	t.Run("test correct error if json schema is invalid", func(t *testing.T) {
		data := []byte(`{"info": null}`)
		reader := bytes.NewReader(data)

		deps, err := parsePackageDeps(reader)
		assert.ErrorIs(t, err, ErrInvalidJson)
		assert.Empty(t, deps)
	})
}

func TestDownloader_FetchPackageDeps(t *testing.T) {
	t.Run("test fetching if package not found", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/requests/json",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(404)
			})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "requests")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, ErrPackageNotFound)
	})

	t.Run("test fetching if server returns invalid json", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/requests/json",
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`not_valid_json`))
				w.WriteHeader(200)
			})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "requests")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, ErrInvalidJson)
	})

	t.Run("test fetching if server returns json with invalid schema", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/requests/json",
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"info": {}}`))
				w.WriteHeader(200)
			})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "requests")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, ErrInvalidJson)
	})

	t.Run("test fetching if server returns json with invalid schema", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/requests/json",
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"info": {}}`))
				w.WriteHeader(200)
			})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "requests")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, ErrInvalidJson)
	})

	t.Run("test fetching package with no dependencies", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/requests/json",
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"info": {"requires_dist": null}}`))
				w.WriteHeader(200)
			})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "requests")
		assert.Empty(t, deps)
		assert.NoError(t, err)
	})

	t.Run("test fetching package", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		mux := http.NewServeMux()
		mux.HandleFunc("/fastapi/json",
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(test_utils.NewServerResponse("starlette=2.0.0", "pydantic>=3", "pyqt; extra"))
				w.WriteHeader(200)

			})
		//mux.HandleFunc("/starlette/json",
		//	func(w http.ResponseWriter, r *http.Request) {
		//		_, _ = w.Write(makeSrvResponse("asyncio"))
		//		w.WriteHeader(200)
		//	})
		//mux.HandleFunc("/pydantic/json",
		//	func(w http.ResponseWriter, r *http.Request) {
		//		_, _ = w.Write(makeSrvResponse())
		//		w.WriteHeader(200)
		//	})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		d := Service{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "fastapi")
		sort.Strings(deps)
		assert.Equal(t, []string{"pydantic", "starlette"}, deps)
		assert.NoError(t, err)
	})
}

func TestDefault(t *testing.T) {
	d := Default()
	assert.NotNil(t, d)
	assert.NotNil(t, d.Client)
	assert.Equal(t, d.BaseURL, "https://pypi.python.org/pypi")
}
