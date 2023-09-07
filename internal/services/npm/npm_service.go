package npm

import (
	"bytes"
	"context"
	"depviz/internal/services/dep_errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Service struct {
	BaseURL string
	Client  *http.Client
}

func Default() *Service {
	return &Service{
		BaseURL: "https://api.npms.io/v2",
		Client:  &http.Client{},
	}
}

func (s *Service) fetch(ctx context.Context, packageName string) ([]byte, error) {
	uri, err := url.JoinPath(s.BaseURL, "package", url.PathEscape(packageName))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	}

	response, err := s.Client.Do(req)
	defer func() {
		_ = response.Body.Close()
	}()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err.Error())
	} else if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: package %s does not exist", dep_errors.ErrPackageNotFound, packageName)
	}
	result := &bytes.Buffer{}
	if _, err := io.Copy(result, response.Body); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func parsePackageDeps(data []byte) ([]string, error) {
	var schema struct {
		Collected struct {
			Metadata struct {
				Dependencies map[string]string `json:"dependencies"`
			} `json:"metadata"`
		} `json:"collected"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("%w: %s", dep_errors.ErrFetch, err)
	}
	dependencies := schema.Collected.Metadata.Dependencies

	names := make([]string, 0, len(dependencies))
	for name := range dependencies {
		names = append(names, name)
	}
	return names, nil
}

func (s *Service) FetchPackageDeps(ctx context.Context, packageName string) ([]string, error) {
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
