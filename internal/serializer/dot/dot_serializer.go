package dot

import (
	"depviz/internal/models"
	"fmt"
	"io"
	"strings"
)

type DotSerializer struct {
}

func (s *DotSerializer) Serialize(graph []models.Edge, out io.Writer) error {
	_, err := fmt.Fprintf(out, "digraph dependencies {\n")
	if err != nil {
		return err
	}
	for _, edge := range graph {
		edge.From = strings.ReplaceAll(edge.From, "-", "_")
		edge.To = strings.ReplaceAll(edge.To, "-", "_")
		_, err = fmt.Fprintf(out, "\t%s -> %s;\n", edge.From, edge.To)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "}")
	return err
}
