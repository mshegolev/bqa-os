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

func TestQuoteUnquoteYAMLRoundTrip(t *testing.T) {
	cases := []string{
		"plain",
		`has "quotes"`,
		"has\nnewline\tand tab",
		`back\slash`,
		"",
	}
	for _, in := range cases {
		got := UnquoteYAML(QuoteYAML(in))
		if got != in {
			t.Fatalf("round-trip mismatch for %q: got %q", in, got)
		}
	}
}

func TestUnquoteYAMLBareScalar(t *testing.T) {
	if got := UnquoteYAML("  bareword  "); got != "bareword" {
		t.Fatalf("expected trimmed bareword, got %q", got)
	}
}

func TestQuoteYAML(t *testing.T) {
	cases := map[string]string{
		`plain`:        `"plain"`,
		`a"b`:          `"a\"b"`,
		`a\b`:          `"a\\b"`,
		`he said "hi"`: `"he said \"hi\""`,
		"line1\nline2": `"line1\nline2"`,
		"tab\there":    `"tab\there"`,
		"cr\rend":      `"cr\rend"`,
	}
	for in, want := range cases {
		if got := QuoteYAML(in); got != want {
			t.Fatalf("QuoteYAML(%q) = %q, want %q", in, got, want)
		}
	}
}
