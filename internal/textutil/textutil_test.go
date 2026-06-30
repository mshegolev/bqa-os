package textutil

import "testing"

func TestHasAny(t *testing.T) {
	if !HasAny("hello world", "nope", "world") {
		t.Fatal("expected match on second needle")
	}
	if HasAny("hello", "x", "y") {
		t.Fatal("expected no match")
	}
	if HasAny("hello") {
		t.Fatal("no needles must report false")
	}
}

func TestQuoteYAML(t *testing.T) {
	cases := map[string]string{
		`plain`:        `"plain"`,
		`a"b`:          `"a\"b"`,
		`a\b`:          `"a\\b"`,
		`he said "hi"`: `"he said \"hi\""`,
	}
	for in, want := range cases {
		if got := QuoteYAML(in); got != want {
			t.Fatalf("QuoteYAML(%q) = %q, want %q", in, got, want)
		}
	}
}
