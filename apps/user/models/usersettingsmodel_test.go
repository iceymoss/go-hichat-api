package models

import "testing"

func TestParseBoolField(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		field  string
		def    bool
		expect bool
	}{
		{"empty json returns default-true", "{}", "readReceiptEnabled", true, true},
		{"empty json returns default-false", "{}", "readReceiptEnabled", false, false},
		{"malformed json returns default", "not-json", "x", true, true},
		{"field true", `{"readReceiptEnabled":true}`, "readReceiptEnabled", false, true},
		{"field false", `{"readReceiptEnabled":false}`, "readReceiptEnabled", true, false},
		{"numeric 1", `{"x":1}`, "x", false, true},
		{"numeric 0", `{"x":0}`, "x", true, false},
		{"string true", `{"x":"true"}`, "x", false, true},
		{"string false", `{"x":"false"}`, "x", true, false},
		{"unrecognized string falls to default", `{"x":"maybe"}`, "x", true, true},
		{"missing field uses default", `{"y":true}`, "x", false, false},
		{"ignores unrelated fields", `{"themeMode":"dark","readReceiptEnabled":false}`, "readReceiptEnabled", true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseBoolField(c.raw, c.field, c.def)
			if got != c.expect {
				t.Errorf("parseBoolField(%q, %q, %v) = %v, want %v", c.raw, c.field, c.def, got, c.expect)
			}
		})
	}
}
