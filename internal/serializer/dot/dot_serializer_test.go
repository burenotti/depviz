package dot

import (
	"bytes"
	"depviz/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DotSerializer_Serialize(t *testing.T) {
	t.Run("test DotSerializer", func(t *testing.T) {
		edges := []models.Edge{
			{"x", "y"},
			{"x", "z"},
			{"y", "z"},
		}
		expected := `digraph dependencies {
	1 [label="x"];
	2 [label="y"];
	3 [label="z"];
	1 -> 2;
	1 -> 3;
	2 -> 3;
}`
		s := DotSerializer{}
		var buf bytes.Buffer
		err := s.Serialize(edges, &buf)
		assert.NoError(t, err)
		assert.Equal(t, expected, buf.String())
	})
	t.Run("test serialization of empty graph", func(t *testing.T) {
		expected := "digraph dependencies {\n}"
		s := DotSerializer{}
		var buf bytes.Buffer
		err := s.Serialize(nil, &buf)
		assert.NoError(t, err)
		assert.Equal(t, expected, buf.String())
	})
}
