package id

import (
	"testing"
)

func TestGenerateULID(t *testing.T) {
	id1 := GenerateULID()
	id2 := GenerateULID()

	if len(id1) != 26 {
		t.Errorf("Expected ULID length 26, got %d", len(id1))
	}

	if id1 == id2 {
		t.Error("Generated identical ULIDs")
	}
}

func TestGetHashID(t *testing.T) {
	input := "hello world"
	// echo -n "hello world" | shasum -a 256
	// b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
	expected := "b94d27b9"

	got := GetHashID(input)
	if got != expected {
		t.Errorf("Expected hash '%s', got '%s'", expected, got)
	}

	// Idempotency check
	if GetHashID(input) != got {
		t.Error("Hash generation is not idempotent")
	}
}
