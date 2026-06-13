package hashutil

import "testing"

func TestComputeHash(t *testing.T) {
	body := []byte("hello")
	key := "secret"

	got := ComputeHash(body, key)

	if got == "" {
		t.Fatal("expected non-empty hash")
	}

	gotAgain := ComputeHash(body, key)
	if got != gotAgain {
		t.Fatal("expected same hash for same body and key")
	}
}

func TestComputeHashEmptyKey(t *testing.T) {
	got := ComputeHash([]byte("hello"), "")

	if got != "" {
		t.Fatalf("expected empty hash, got %q", got)
	}
}
