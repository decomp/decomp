package cfg

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/graph/encoding"
)

// Attrs specifies a set of DOT attributes as key-value pairs.
type Attrs map[string]string

// DOTAttributes returns the DOT attributes of a node or edge.
func (a Attrs) DOTAttributes() []encoding.Attribute {
	var attrs []encoding.Attribute
	var keys []string
	for key := range a {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := a[key]
		if key == "label" {
			val = fmt.Sprintf("%q", val)
		}
		attr := encoding.Attribute{
			Key:   key,
			Value: val,
		}
		attrs = append(attrs, attr)
	}
	return attrs
}
