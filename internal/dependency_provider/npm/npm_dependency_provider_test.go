package npm

import (
	"context"
	"depviz/internal/dependency_provider/dep_errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"
)

func TestParsePackageDeps(t *testing.T) {
	t.Run("test can parse valid json", func(t *testing.T) {
		data := []byte(`
{
	"name": "@vue/compiler-dom",
	"dependencies": {
		"@vue/shared": "3.2.45",
		"@vue/compiler-core": "3.2.45"
	}
}
`)
		deps, err := parsePackageDeps(data)
		assert.NoError(t, err)

		sort.Strings(deps)
		expected := []string{"@vue/shared", "@vue/compiler-core"}
		sort.Strings(expected)
		assert.Equal(t, expected, deps)
	})

	t.Run("test correctly parses packages with no dependencies", func(t *testing.T) {
		data := []byte(`
{
	"name": "@vue/compiler-dom",
	"dependencies": {}
}
`)
		deps, err := parsePackageDeps(data)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(deps))
	})

	t.Run("test parsing if json is invalid", func(t *testing.T) {
		_, err := parsePackageDeps([]byte(`{`))
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})

	t.Run("test parsing if json schema is invalid", func(t *testing.T) {
		_, err := parsePackageDeps([]byte(`{"dependencies": 1}`))
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})
}

func TestDownloader_FetchPackageDeps(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/valid-dependencies/latest",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"dependencies": {"router": "1.0.0", "shared": "^2.0.0"}}`))
			w.WriteHeader(200)
		})
	mux.HandleFunc("/empty-dependencies/latest",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"dependencies": {}}`))
			w.WriteHeader(200)
		})
	mux.HandleFunc("/invalid-json/latest",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{`))
			w.WriteHeader(200)
		})

	mux.HandleFunc("/invalid-schema/latest",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{}`))
			w.WriteHeader(200)
		})

	mux.HandleFunc("/not-found/latest",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})

	t.Run("test fetching if package not found", func(t *testing.T) {
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "not-found")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, dep_errors.ErrPackageNotFound)
	})

	t.Run("test fetching if server returns invalid json", func(t *testing.T) {
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "invalid-json")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})

	t.Run("test fetching if server returns json with invalid schema", func(t *testing.T) {
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "invalid-schema")
		assert.Empty(t, deps)
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})

	t.Run("test fetching package with no dependencies", func(t *testing.T) {
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "empty-dependencies")
		assert.Empty(t, deps)
		assert.NoError(t, err)
	})

	t.Run("test fetching package", func(t *testing.T) {
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		deps, err := d.FetchPackageDeps(ctx, "valid-dependencies")
		sort.Strings(deps)
		assert.Equal(t, []string{"router", "shared"}, deps)
		assert.NoError(t, err)
	})

	t.Run("test fetching if server is unreachable", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: "http://localhost",
			Client:  &http.Client{},
		}

		_, err := d.FetchPackageDeps(ctx, "valid-dependencies")
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})
	t.Run("test if fetching url is invalid", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := DependencyProvider{
			BaseURL: ":::",
			Client:  &http.Client{},
		}

		_, err := d.FetchPackageDeps(ctx, "valid-dependencies")
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})
}

func TestDefault(t *testing.T) {
	d := Default()
	assert.NotNil(t, d)
	assert.NotNil(t, d.Client)
	assert.Equal(t, "https://registry.npmjs.com/", d.BaseURL)
}
