package npm

import (
	"bytes"
	"context"
	"depviz/internal/dependency_provider/dep_errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type DependencyProvider struct {
	BaseURL string
	Client  *http.Client
}

func Default() *DependencyProvider {
	return &DependencyProvider{
		BaseURL: "https://registry.npmjs.com/",
		Client:  &http.Client{},
	}
}

func (s *DependencyProvider) fetch(ctx context.Context, packageName string) ([]byte, error) {
	uri, err := url.JoinPath(s.BaseURL, url.PathEscape(packageName), "latest")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	}

	response, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	} else if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: package %s does not exist", dep_errors.ErrPackageNotFound, packageName)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	result := &bytes.Buffer{}
	if _, err := io.Copy(result, response.Body); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func parsePackageDeps(data []byte) ([]string, error) {
	var schema struct {
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err)
	}
	if schema.Dependencies == nil {
		return nil, fmt.Errorf("%w: invalid json schema", dep_errors.ErrFetch)
	}
	dependencies := schema.Dependencies

	names := make([]string, 0, len(dependencies))
	for name := range dependencies {
		names = append(names, name)
	}
	return names, nil
}

func (s *DependencyProvider) FetchPackageDeps(ctx context.Context, packageName string) ([]string, error) {
	data, err := s.fetch(ctx, packageName)
	if err != nil {
		return nil, err
	}
	deps, err := parsePackageDeps(data)
	if err != nil {
		return nil, err
	}
	return deps, nil
}
