package locate

import "testing"

func TestParseNameAndRegexp(t *testing.T) {
	for i, tc := range []struct {
		spec       string
		path, expr string
	}{
		{"", "", ".*"},
		{"a", "a", ".*"},
		{".", "", ""},
		{".a", "", "a"},
		{"a.b", "a", "b"},
		{"/x/y/z/a.b", "/x/y/z/a", "b"},
		{"/x/y/z/a.", "/x/y/z/a", ""},
		{"/x/y/z/a..*", "/x/y/z/a", ".*"},
		{"/x/y/z/a.foo.*", "/x/y/z/a", "foo.*"},
	} {
		path, expr := parseSpecAndRegexp(tc.spec)
		if got, want := path, tc.path; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		if got, want := expr, tc.expr; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
	}
}
