package event

import (
	"encoding/json"
	"sort"
	"strconv"
)

// Field is one scalar value from a JSON document and its path, written as
// dotted keys and [index] for arrays (for example "content[0].text").
type Field struct {
	Path  string
	Value string
}

// Flatten walks a decoded JSON value and returns every string and
// json.Number in it as a Field, with object keys visited in sorted order.
// Booleans and nulls are skipped. A top-level scalar has an empty Path.
func Flatten(v any) []Field {
	fields := make([]Field, 0)
	flattenInto("", v, &fields)
	return fields
}

func flattenInto(path string, v any, fields *[]Field) {
	switch val := v.(type) {
	case string:
		*fields = append(*fields, Field{Path: path, Value: val})
	case json.Number:
		*fields = append(*fields, Field{Path: path, Value: val.String()})
	case map[string]any:
		flattenObject(path, val, fields)
	case []any:
		flattenArray(path, val, fields)
	}
}

func flattenObject(path string, val map[string]any, fields *[]Field) {
	for _, key := range SortedKeys(val) {
		flattenInto(childPath(path, key), val[key], fields)
	}
}

func flattenArray(path string, val []any, fields *[]Field) {
	for i, item := range val {
		flattenInto(indexPath(path, i), item, fields)
	}
}

func indexPath(parent string, i int) string {
	return parent + "[" + strconv.Itoa(i) + "]"
}

// SortedKeys returns the keys of obj in ascending order.
func SortedKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func childPath(parent string, key string) string {
	if parent == "" {
		return key
	}

	return parent + "." + key
}
