package serializer

import (
	"testing"
)

// testStruct is a helper type used in serializer tests.
type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

// testNested is a helper type with nested fields.
type testNested struct {
	ID   int        `json:"id"`
	Data testStruct `json:"data"`
	Tags []string   `json:"tags"`
}

func TestJSONSerializer_MarshalStruct(t *testing.T) {
	s := NewJSONSerializer()

	input := testStruct{Name: "Alice", Age: 30, Email: "alice@example.com"}
	data, err := s.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Marshal() returned empty data")
	}

	var result testStruct
	err = s.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if result != input {
		t.Errorf("Unmarshal() = %+v, want %+v", result, input)
	}
}

func TestJSONSerializer_MarshalSlice(t *testing.T) {
	s := NewJSONSerializer()

	input := []string{"a", "b", "c"}
	data, err := s.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var result []string
	err = s.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if len(result) != len(input) {
		t.Fatalf("Unmarshal() len = %d, want %d", len(result), len(input))
	}
	for i, v := range result {
		if v != input[i] {
			t.Errorf("Unmarshal()[%d] = %q, want %q", i, v, input[i])
		}
	}
}

func TestJSONSerializer_MarshalMap(t *testing.T) {
	s := NewJSONSerializer()

	input := map[string]int{"a": 1, "b": 2, "c": 3}
	data, err := s.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var result map[string]int
	err = s.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if len(result) != len(input) {
		t.Fatalf("Unmarshal() len = %d, want %d", len(result), len(input))
	}
	for k, v := range input {
		if result[k] != v {
			t.Errorf("Unmarshal()[%q] = %d, want %d", k, result[k], v)
		}
	}
}

func TestJSONSerializer_MarshalPrimitives(t *testing.T) {
	s := NewJSONSerializer()

	tests := []struct {
		name  string
		input any
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool_true", true},
		{"bool_false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := s.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal(%v) unexpected error: %v", tt.input, err)
			}
			if len(data) == 0 {
				t.Fatal("Marshal() returned empty data")
			}
		})
	}
}

func TestJSONSerializer_MarshalNested(t *testing.T) {
	s := NewJSONSerializer()

	input := testNested{
		ID:   1,
		Data: testStruct{Name: "Bob", Age: 25},
		Tags: []string{"admin", "user"},
	}
	data, err := s.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var result testNested
	err = s.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if result.ID != input.ID {
		t.Errorf("ID = %d, want %d", result.ID, input.ID)
	}
	if result.Data != input.Data {
		t.Errorf("Data = %+v, want %+v", result.Data, input.Data)
	}
}

func TestJSONSerializer_MarshalZeroValues(t *testing.T) {
	s := NewJSONSerializer()

	input := testStruct{} // all zero values
	data, err := s.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var result testStruct
	err = s.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if result != input {
		t.Errorf("Unmarshal() = %+v, want %+v", result, input)
	}
}

func TestJSONSerializer_MarshalNil(t *testing.T) {
	s := NewJSONSerializer()

	data, err := s.Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal(nil) unexpected error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal(nil) = %q, want %q", string(data), "null")
	}
}

func TestJSONSerializer_UnmarshalEmptyBytes(t *testing.T) {
	s := NewJSONSerializer()

	var result testStruct
	err := s.Unmarshal([]byte{}, &result)
	if err == nil {
		t.Fatal("Unmarshal(empty) expected error, got nil")
	}
}

func TestJSONSerializer_UnmarshalInvalidJSON(t *testing.T) {
	s := NewJSONSerializer()

	var result testStruct
	err := s.Unmarshal([]byte("{not valid json"), &result)
	if err == nil {
		t.Fatal("Unmarshal(invalid) expected error, got nil")
	}
}

func TestJSONSerializer_UnmarshalTypeMismatch(t *testing.T) {
	s := NewJSONSerializer()

	// Marshal a string, then try to unmarshal into int pointer
	data, _ := s.Marshal("hello")
	var result int
	err := s.Unmarshal(data, &result)
	if err == nil {
		t.Fatal("Unmarshal(type mismatch) expected error, got nil")
	}
}

func TestJSONSerializer_MarshalUnmarshalableType(t *testing.T) {
	s := NewJSONSerializer()

	// Channels cannot be marshaled to JSON
	ch := make(chan int)
	_, err := s.Marshal(ch)
	if err == nil {
		t.Fatal("Marshal(chan) expected error, got nil")
	}
}

func TestJSONSerializer_ImplementsInterface(t *testing.T) {
	// Compile-time check that JSONSerializer implements Serializer
	var _ Serializer = (*JSONSerializer)(nil)
}
