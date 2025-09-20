package pokecache

import (
	"testing"
	"time"
)

func TestCacheAddGetAndReap(t *testing.T) {
	// Use a short interval so the test runs fast
	interval := 50 * time.Millisecond
	c := NewCache(interval)

	key := "https://pokeapi.co/api/v2/location-area"
	val := []byte(`{"dummy":"value"}`)
	c.Add(key, val)

	// Immediately get should return the value
	got, ok := c.Get(key)
	if !ok {
		t.Fatalf("expected cached value to exist")
	}
	if string(got) != string(val) {
		t.Fatalf("expected %s got %s", string(val), string(got))
	}

	// Wait enough time for the reap loop to run and remove the entry
	time.Sleep(2 * interval)

	_, ok = c.Get(key)
	if ok {
		t.Fatalf("expected entry to be reaped after interval")
	}
}
