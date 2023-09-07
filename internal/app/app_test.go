package app

import (
	"bytes"
	"context"
	"depviz/internal/dependency_provider/pip"
	"depviz/internal/dependency_provider/pip/test_utils"
	"depviz/internal/models"
	"depviz/internal/serializer/dot"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"
)

func TestApp_GetDependencyGraph(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastapi/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse("starlette=2.0.0", "pydantic>=3", "pyqt; extra"))
			w.WriteHeader(200)

		})
	mux.HandleFunc("/starlette/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse("asyncio"))
			w.WriteHeader(200)
		})
	mux.HandleFunc("/asyncio/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse())
			w.WriteHeader(200)
		})
	mux.HandleFunc("/pydantic/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse())
			w.WriteHeader(200)
		})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("test fetching pip graph with sub dependencies", func(t *testing.T) {
		expected := []models.Edge{
			{"fastapi", "pydantic"},
			{"fastapi", "starlette"},
			{"starlette", "asyncio"},
		}
		sortEdges(expected)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := &pip.DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		app := New(d, nil)
		deps, err := app.GetDependencyGraph(ctx, "fastapi")
		sortEdges(deps)

		assert.NoError(t, err)
		assert.Equal(t, deps, expected)
	})
}

func TestApp_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastapi/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse("starlette=2.0.0", "pydantic>=3", "pyqt; extra"))
			w.WriteHeader(200)

		})
	mux.HandleFunc("/starlette/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse("asyncio"))
			w.WriteHeader(200)
		})
	mux.HandleFunc("/asyncio/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse())
			w.WriteHeader(200)
		})
	mux.HandleFunc("/pydantic/json",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(test_utils.NewServerResponse())
			w.WriteHeader(200)
		})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("test running app on valid data", func(t *testing.T) {
		expected := `digraph dependencies {
	1 [label="fastapi"];
	2 [label="starlette"];
	3 [label="pydantic"];
	4 [label="asyncio"];
	1 -> 2;
	1 -> 3;
	2 -> 4;
}`
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		d := &pip.DependencyProvider{
			BaseURL: srv.URL,
			Client:  &http.Client{},
		}
		s := &dot.DotSerializer{}

		buf := &bytes.Buffer{}
		app := New(d, s)
		err := app.Run(ctx, "fastapi", buf)
		assert.NoError(t, err)
		assert.Equal(t, expected, buf.String())
	})
}

func edgeLess(left models.Edge, right models.Edge) bool {
	if left.From < right.From {
		return false
	}
	return left.To < right.To
}

func sortEdges(edges []models.Edge) {
	sort.Slice(edges, func(i, j int) bool {
		return edgeLess(edges[i], edges[j])
	})
}
