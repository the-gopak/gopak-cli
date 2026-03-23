package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExecutable_UnmarshalYAML(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantSet bool
		binary  string
		args    []string
	}{
		{
			name:    "string form",
			input:   `executable: mytool`,
			wantSet: true,
			binary:  "mytool",
			args:    nil,
		},
		{
			name:    "array with binary only",
			input:   `executable: ["npx"]`,
			wantSet: true,
			binary:  "npx",
			args:    nil,
		},
		{
			name:    "array with binary and args",
			input:   `executable: ["npx", "-y", "prettier"]`,
			wantSet: true,
			binary:  "npx",
			args:    []string{"-y", "prettier"},
		},
		{
			name:    "empty string",
			input:   `executable: ""`,
			wantSet: false,
			binary:  "",
			args:    nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s struct {
				Executable Executable `yaml:"executable"`
			}
			if err := yaml.Unmarshal([]byte(c.input), &s); err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			if s.Executable.IsSet() != c.wantSet {
				t.Fatalf("IsSet: want %v, got %v", c.wantSet, s.Executable.IsSet())
			}
			if s.Executable.Binary() != c.binary {
				t.Fatalf("Binary: want %q, got %q", c.binary, s.Executable.Binary())
			}
			gotArgs := s.Executable.Args()
			if len(gotArgs) != len(c.args) {
				t.Fatalf("Args len: want %d, got %d (%v)", len(c.args), len(gotArgs), gotArgs)
			}
			for i, a := range c.args {
				if gotArgs[i] != a {
					t.Fatalf("Args[%d]: want %q, got %q", i, a, gotArgs[i])
				}
			}
		})
	}
}

func TestExecutable_UnmarshalYAML_InvalidType(t *testing.T) {
	input := `executable: {key: val}`
	var s struct {
		Executable Executable `yaml:"executable"`
	}
	if err := yaml.Unmarshal([]byte(input), &s); err == nil {
		t.Fatal("expected error for mapping node, got nil")
	}
}

func TestExecutable_JSON_RoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantJSON string
		binary   string
		args     []string
	}{
		{
			name:     "string form",
			input:    `"mytool"`,
			wantJSON: `"mytool"`,
			binary:   "mytool",
			args:     nil,
		},
		{
			name:     "array with binary only compacts to string",
			input:    `["npx"]`,
			wantJSON: `"npx"`,
			binary:   "npx",
			args:     nil,
		},
		{
			name:     "array with args",
			input:    `["npx","-y","prettier"]`,
			wantJSON: `["npx","-y","prettier"]`,
			binary:   "npx",
			args:     []string{"-y", "prettier"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var e Executable
			if err := json.Unmarshal([]byte(c.input), &e); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if e.Binary() != c.binary {
				t.Fatalf("Binary: want %q, got %q", c.binary, e.Binary())
			}
			gotArgs := e.Args()
			if len(gotArgs) != len(c.args) {
				t.Fatalf("Args len: want %d, got %d", len(c.args), len(gotArgs))
			}

			out, err := json.Marshal(e)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			if string(out) != c.wantJSON {
				t.Fatalf("marshaled JSON: want %s, got %s", c.wantJSON, out)
			}
		})
	}
}

func TestExecutable_IsSet(t *testing.T) {
	cases := []struct {
		name string
		e    Executable
		want bool
	}{
		{"nil slice", nil, false},
		{"empty slice", Executable{}, false},
		{"empty string element", Executable{""}, false},
		{"valid binary", Executable{"mytool"}, true},
		{"binary with args", Executable{"npx", "-y", "prettier"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.e.IsSet() != c.want {
				t.Fatalf("IsSet: want %v, got %v", c.want, c.e.IsSet())
			}
		})
	}
}
