package tldwhois

import (
	"testing"
	"time"
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

func TestStop(t *testing.T) {
	m := newTldManager(tldsURL)

	m.stop()

	select {
	case <-m.donec:
	case <-time.After(time.Second * 5):
		t.Fatal("stop too long")
	}
}

func TestTimeout(t *testing.T) {
	m := newTldManagerWithLimitAndTimeout(tldsURL, defaultLimit, time.Nanosecond)
	select {
	case <-m.donec:
	case <-time.After(time.Second * 5):
		t.Fatal("stop too long")
	}
}
