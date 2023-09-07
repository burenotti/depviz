package npm

import (
	"depviz/internal/dependency_provider/dep_errors"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestParsePackageDeps(t *testing.T) {
	t.Run("test can parse valid json", func(t *testing.T) {
		data := []byte(`
{
	"analyzedAt": "2022-11-11T23:56:34.902Z",
	"collected": {
		"metadata": {
			"name": "@vue/compiler-dom",
			"dependencies": {
				"@vue/shared": "3.2.45",
				"@vue/compiler-core": "3.2.45"
			}
		}
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
	"analyzedAt": "2022-11-11T23:56:34.902Z",
	"collected": {
		"metadata": {
			"name": "@vue/compiler-dom",
			"dependencies": {}
		}
	}
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
		_, err := parsePackageDeps([]byte(`{"collected": 1}`))
		assert.ErrorIs(t, err, dep_errors.ErrFetch)
	})
}

func TestService_FetchPackageDeps(t *testing.T) {

}
