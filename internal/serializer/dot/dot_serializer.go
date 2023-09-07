package dot

import (
	"depviz/internal/models"
	"fmt"
	"io"
	"regexp"
)

var escapeRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type DotSerializer struct {
}

func (s *DotSerializer) Serialize(graph []models.Edge, out io.Writer) error {
	idx := 0
	labels := make(map[string]int)

	for _, edge := range graph {
		if _, ok := labels[edge.From]; !ok {
			idx++
			labels[edge.From] = idx
		}
		if _, ok := labels[edge.To]; !ok {
			idx++
			labels[edge.To] = idx
		}
	}

	_, err := fmt.Fprintf(out, "digraph dependencies {\n")
	if err != nil {
		return err
	}
	for label, idx := range labels {
		if _, err := fmt.Fprintf(out, "\t %d[label=\"%s\"];\n", idx, label); err != nil {
			return nil
		}
	}
	for _, edge := range graph {
		from := labels[edge.From]
		to := labels[edge.To]
		if _, err := fmt.Fprintf(out, "\t %d -> %d;\n", from, to); err != nil {
			return nil
		}
	}
	_, err = fmt.Fprintf(out, "}")
	return err
}
