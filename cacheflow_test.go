package cacheflow

import (
	"testing"

	"github.com/elokanugrah/go-cacheflow/serializer"
	"github.com/elokanugrah/go-cacheflow/store"
)

func TestNew_Defaults(t *testing.T) {
	cf := New()

	if cf.store == nil {
		t.Fatal("New() store is nil, want MemoryStore")
	}
	if cf.serializer == nil {
		t.Fatal("New() serializer is nil, want JSONSerializer")
	}

	// Verify types
	if _, ok := cf.store.(*store.MemoryStore); !ok {
		t.Errorf("New() store type = %T, want *store.MemoryStore", cf.store)
	}
	if _, ok := cf.serializer.(*serializer.JSONSerializer); !ok {
		t.Errorf("New() serializer type = %T, want *serializer.JSONSerializer", cf.serializer)
	}
}

func TestNew_WithCustomStore(t *testing.T) {
	customStore := store.NewMemoryStore()
	cf := New(WithStore(customStore))

	if cf.store != customStore {
		t.Error("WithStore() did not set the custom store")
	}
}

func TestNew_WithCustomSerializer(t *testing.T) {
	customSerializer := serializer.NewJSONSerializer()
	cf := New(WithSerializer(customSerializer))

	if cf.serializer != customSerializer {
		t.Error("WithSerializer() did not set the custom serializer")
	}
}

func TestNew_MultipleOptions(t *testing.T) {
	customStore := store.NewMemoryStore()
	customSerializer := serializer.NewJSONSerializer()

	cf := New(
		WithStore(customStore),
		WithSerializer(customSerializer),
	)

	if cf.store != customStore {
		t.Error("WithStore() did not set the custom store")
	}
	if cf.serializer != customSerializer {
		t.Error("WithSerializer() did not set the custom serializer")
	}
}

func TestNew_LastOptionWins(t *testing.T) {
	store1 := store.NewMemoryStore()
	store2 := store.NewMemoryStore()

	cf := New(
		WithStore(store1),
		WithStore(store2),
	)

	if cf.store != store2 {
		t.Error("Last WithStore() should override the first")
	}
}
