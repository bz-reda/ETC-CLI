package cmd

import (
	"os"
	"testing"
)

func TestParseDotenv(t *testing.T) {
	tmp := t.TempDir() + "/.env"
	content := `# a comment
BLANK_LINE_NEXT=

# another comment
SIMPLE=value1
QUOTED_DOUBLE="hello world"
QUOTED_SINGLE='single space here'
EXPORT_PREFIX=prefix_value
export EXPORTED=exp_val
export QUOTED_EXPORTED="e and u"

# inline whitespace around equals is stripped around the key only
  WITH_LEADING_WHITESPACE  =  value_keeps_space

EMPTY_VALUE=
SYMBOLS=a=b=c
`
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := parseDotenv(tmp)
	if err != nil {
		t.Fatal(err)
	}

	byKey := map[string]string{}
	for _, e := range entries {
		byKey[e.Key] = e.Value
	}

	cases := []struct {
		key, want string
	}{
		{"SIMPLE", "value1"},
		{"QUOTED_DOUBLE", "hello world"},
		{"QUOTED_SINGLE", "single space here"},
		{"EXPORT_PREFIX", "prefix_value"},
		{"EXPORTED", "exp_val"},
		{"QUOTED_EXPORTED", "e and u"},
		{"EMPTY_VALUE", ""},
		{"SYMBOLS", "a=b=c"}, // split only on first =
		{"BLANK_LINE_NEXT", ""},
		{"WITH_LEADING_WHITESPACE", "value_keeps_space"},
	}
	for _, c := range cases {
		got, ok := byKey[c.key]
		if !ok {
			t.Errorf("%s missing from parse", c.key)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.key, got, c.want)
		}
	}

	// Ensure comments and blank lines are not parsed as keys
	for _, forbidden := range []string{"# a comment", "", "a comment"} {
		if _, ok := byKey[forbidden]; ok {
			t.Errorf("unexpected entry parsed from comment/blank: %q", forbidden)
		}
	}
}
