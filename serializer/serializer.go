// Package serializer provides data serialization abstractions for CacheFlow.
//
// Serializer defines a simple interface for converting Go values to
// and from byte slices. Implementations are used by the CacheFlow
// orchestration layer to serialize typed values before storing them
// and to deserialize them upon retrieval.
package serializer

// Serializer defines the interface for encoding and decoding Go values.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type Serializer interface {
	// Marshal encodes v into a byte slice.
	//
	// Returns an error if v cannot be serialized (e.g., contains
	// unsupported types).
	Marshal(v any) ([]byte, error)

	// Unmarshal decodes data into the value pointed to by v.
	//
	// v must be a non-nil pointer. Returns an error if the data
	// is malformed or cannot be decoded into the target type.
	Unmarshal(data []byte, v any) error
}
