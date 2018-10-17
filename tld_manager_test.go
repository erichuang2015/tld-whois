package tldwhois

import (
	"testing"
)

func Test_newTldManager(t *testing.T) {
	want := 10
	m := newTldManagerWithLimitAndTimeout(tldsURL, want, defaultTimeout)
	for range m.TldC() {
	}
	if got := int(m.load()); got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}
