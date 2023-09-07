package dot

import (
	"depviz/internal/models"
	"fmt"
	"io"
	"regexp"
	"sort"
)

var escapeRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

type DotSerializer struct {
}

type labelsMap map[string]int

type pair struct {
	label string
	value int
}

func (m labelsMap) AsSortedPairs() []pair {
	result := make([]pair, 0, len(m))
	for label, value := range m {
		result = append(result, pair{label, value})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].label < result[j].label
	})
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].value < result[j].value
	})
	return result
}

func (s *DotSerializer) Serialize(graph []models.Edge, out io.Writer) error {
	idx := 0
	labels := make(labelsMap)

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

	pairs := labels.AsSortedPairs()
	for _, pair := range pairs {
		if _, err := fmt.Fprintf(out, "\t%d [label=\"%s\"];\n", pair.value, pair.label); err != nil {
			return nil
		}
	}
	for _, edge := range graph {
		from := labels[edge.From]
		to := labels[edge.To]
		if _, err := fmt.Fprintf(out, "\t%d -> %d;\n", from, to); err != nil {
			return nil
		}
	}
	_, err = fmt.Fprintf(out, "}")
	return err
}
