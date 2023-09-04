package dep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidJson     = errors.New("invalid json")
	ErrPackageNotFound = errors.New("package not found")
)

type Service struct {
	BaseURL string
	Client  *http.Client
}

func Default() *Service {
	return &Service{
		BaseURL: "https://pypi.python.org/pypi",
		Client:  &http.Client{},
	}
}

func (d *Service) fetch(ctx context.Context, packageName string) (io.ReadCloser, error) {
	uri, err := url.JoinPath(d.BaseURL, packageName, "json")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == 404 {
		return nil, fmt.Errorf("%w: %s", ErrPackageNotFound, packageName)
	}

	return resp.Body, nil
}

func parsePackageDeps(reader io.Reader) ([]string, error) {
	var deps []string
	decoder := json.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidJson, err)
		}

		tokenStr, ok := token.(string)
		if ok && tokenStr == "requires_dist" {
			err := decoder.Decode(&deps)
			if err == nil {
				return deps, nil
			} else {
				return nil, fmt.Errorf("%w: %s", ErrInvalidJson, err)
			}
		}
	}
	return nil, fmt.Errorf("%w: invalid json", ErrInvalidJson)
}

func (d *Service) FetchPackageDeps(ctx context.Context, packageName string) ([]string, error) {
	data, err := d.fetch(ctx, packageName)
	if err != nil {
		return nil, err
	}
	deps, err := parsePackageDeps(data)
	if err != nil {
		return nil, err
	}
	return cleanPackageDeps(deps), nil
}

func cleanPackageDeps(deps []string) []string {
	pattern, err := regexp.Compile(`^[a-zA-Z\-_0-9.]+`)
	if err != nil {
		panic(err.Error())
	}

	result := make([]string, 0, len(deps))
	for _, dep := range deps {
		// ignore extra dependencies
		if strings.Contains(dep, "extra") {
			continue
		}

		result = append(result, string(pattern.Find([]byte(dep))))
	}
	return result
}
