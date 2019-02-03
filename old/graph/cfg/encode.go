package cfg

import (
	"sort"

	"gonum.org/v1/gonum/graph/encoding"
)

// Attrs specifies a set of DOT attributes as key-value pairs.
type Attrs map[string]string

// Attributes returns the DOT attributes of a node or edge.
func (a Attrs) Attributes() []encoding.Attribute {
	var keys []string
	for key := range a {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var attrs []encoding.Attribute
	for _, key := range keys {
		attr := encoding.Attribute{
			Key:   key,
			Value: a[key],
		}
		attrs = append(attrs, attr)
	}
	return attrs
}
