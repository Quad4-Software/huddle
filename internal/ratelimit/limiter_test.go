package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowsUpToLimit(t *testing.T) {
	l := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("client") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
	if l.Allow("client") {
		t.Fatal("expected deny after limit")
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	l := New(1, 20*time.Millisecond)
	if !l.Allow("client") {
		t.Fatal("expected first allow")
	}
	if l.Allow("client") {
		t.Fatal("expected deny inside window")
	}
	time.Sleep(25 * time.Millisecond)
	if !l.Allow("client") {
		t.Fatal("expected allow after window reset")
	}
}

func TestLimiterDisabledWhenLimitZero(t *testing.T) {
	l := New(0, time.Minute)
	for i := 0; i < 5; i++ {
		if !l.Allow("client") {
			t.Fatal("expected disabled limiter to always allow")
		}
	}
}

func TestLimiterKeysAreIndependent(t *testing.T) {
	l := New(1, time.Minute)
	if !l.Allow("a") {
		t.Fatal("expected allow for a")
	}
	if !l.Allow("b") {
		t.Fatal("expected allow for b")
	}
}
