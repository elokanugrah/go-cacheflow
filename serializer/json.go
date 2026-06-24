package serializer

import "encoding/json"

// JSONSerializer implements the Serializer interface using encoding/json
// from the Go standard library.
//
// JSONSerializer is safe for concurrent use — encoding/json.Marshal
// and encoding/json.Unmarshal are stateless and goroutine-safe.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSONSerializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Marshal encodes v into a JSON byte slice.
//
// Returns an error if v contains types that cannot be marshaled
// to JSON (e.g., channels, functions).
func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal decodes JSON data into the value pointed to by v.
//
// v must be a non-nil pointer. Returns an error if data is not
// valid JSON or cannot be decoded into the target type.
func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
