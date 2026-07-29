package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testBDF = `STARTFONT 2.1
FONT cli-test
SIZE 8 75 75
FONTBOUNDINGBOX 8 6 0 -1
STARTPROPERTIES 4
FONT_ASCENT 5
FONT_DESCENT 1
CHARSET_REGISTRY "ISO10646"
CHARSET_ENCODING "1"
ENDPROPERTIES
CHARS 3
STARTCHAR B
ENCODING 66
DWIDTH 8 0
BBX 8 1 0 0
BITMAP
42
ENDCHAR
STARTCHAR A
ENCODING 65
DWIDTH 4 0
BBX 2 5 1 -1
BITMAP
FF
80
40
20
00
ENDCHAR
STARTCHAR JAPANESE_DAY
ENCODING 26085
DWIDTH 8 0
BBX 8 1 0 0
BITMAP
FF
ENDCHAR
ENDFONT
`

func TestRunGeneratesSubsetGoSource(t *testing.T) {
	dir := t.TempDir()
	bdfPath := writeTestFile(t, dir, "font.bdf", []byte(testBDF))
	subsetPath := writeTestFile(t, dir, "subset.txt", []byte("A\u65e5"))
	outputPath := writeTestFile(t, dir, "font.go", []byte("old output"))
	var stderr bytes.Buffer

	code := run(validArgs(bdfPath, subsetPath, outputPath), &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}

	generated, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(generated, []byte("old output")) {
		t.Fatal("existing output was not overwritten")
	}
	if _, err := parser.ParseFile(token.NewFileSet(), outputPath, generated, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, generated)
	}
	source := string(generated)
	compact := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(source)
	for _, expected := range []string{
		"packagegenerated", "varHeadlineFont", "Rune:'A'", `Rune:'\u65e5'`,
		"Ascent:5", "Descent:1", "LineGap:0",
		"Width:2", "Height:5", "AdvanceX:4", "BearingX:1", "BearingY:4",
	} {
		if !strings.Contains(compact, expected) {
			t.Errorf("generated source does not contain %q\n%s", expected, source)
		}
	}
	if strings.Contains(source, "'B'") {
		t.Errorf("generated source contains glyph B omitted from subset\n%s", source)
	}
}

func TestRunMissingRequiredFlags(t *testing.T) {
	dir := t.TempDir()
	bdfPath := writeTestFile(t, dir, "font.bdf", []byte(testBDF))
	subsetPath := writeTestFile(t, dir, "subset.txt", []byte("A"))
	outputPath := filepath.Join(dir, "font.go")
	all := validArgs(bdfPath, subsetPath, outputPath)

	for _, flagName := range []string{"-bdf", "-subset", "-package", "-var", "-o"} {
		t.Run(flagName, func(t *testing.T) {
			args := removeFlag(all, flagName)
			var stderr bytes.Buffer
			if code := run(args, &stderr); code != 2 {
				t.Fatalf("run() code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "missing required flag "+flagName) {
				t.Errorf("stderr = %q, want missing %s", stderr.String(), flagName)
			}
		})
	}
}

func TestRunFlagErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown flag", args: []string{"-unknown"}, want: "flag provided but not defined"},
		{name: "positional argument", args: append(validArgs("font.bdf", "subset.txt", "font.go"), "extra"), want: "unexpected positional"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := run(test.args, &stderr); code != 2 {
				t.Fatalf("run() code = %d, want 2", code)
			}
			if !strings.Contains(stderr.String(), test.want) {
				t.Errorf("stderr = %q, want substring %q", stderr.String(), test.want)
			}
		})
	}
}

func TestRunProcessingErrors(t *testing.T) {
	tests := []struct {
		name         string
		prepareBDF   func(*testing.T, string) string
		prepareSet   func(*testing.T, string) string
		packageName  string
		variableName string
		outputPath   func(string) string
		want         string
	}{
		{
			name:       "BDF does not exist",
			prepareBDF: func(_ *testing.T, dir string) string { return filepath.Join(dir, "missing.bdf") },
			want:       "open BDF",
		},
		{
			name:       "subset does not exist",
			prepareSet: func(_ *testing.T, dir string) string { return filepath.Join(dir, "missing.txt") },
			want:       "read subset",
		},
		{
			name:       "invalid BDF",
			prepareBDF: func(t *testing.T, dir string) string { return writeTestFile(t, dir, "bad.bdf", []byte("not BDF\n")) },
			want:       "parse BDF",
		},
		{
			name:       "invalid UTF-8 subset",
			prepareSet: func(t *testing.T, dir string) string { return writeTestFile(t, dir, "bad.txt", []byte{0xff}) },
			want:       "valid UTF-8",
		},
		{
			name: "missing requested glyph",
			prepareSet: func(t *testing.T, dir string) string {
				return writeTestFile(t, dir, "missing-glyph.txt", []byte("\u96ea"))
			},
			want: "missing glyph U+96EA",
		},
		{
			name:        "invalid package name",
			packageName: "bad-name",
			want:        "PackageName",
		},
		{
			name:         "invalid variable name",
			variableName: "var",
			want:         "VariableName",
		},
		{
			name:       "output directory does not exist",
			outputPath: func(dir string) string { return filepath.Join(dir, "missing", "font.go") },
			want:       "write output",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			bdfPath := writeTestFile(t, dir, "font.bdf", []byte(testBDF))
			if test.prepareBDF != nil {
				bdfPath = test.prepareBDF(t, dir)
			}
			subsetPath := writeTestFile(t, dir, "subset.txt", []byte("A"))
			if test.prepareSet != nil {
				subsetPath = test.prepareSet(t, dir)
			}
			packageName := test.packageName
			if packageName == "" {
				packageName = "generated"
			}
			variableName := test.variableName
			if variableName == "" {
				variableName = "Example"
			}
			outputPath := filepath.Join(dir, "font.go")
			if test.outputPath != nil {
				outputPath = test.outputPath(dir)
			}
			args := []string{
				"-bdf", bdfPath, "-subset", subsetPath,
				"-package", packageName, "-var", variableName, "-o", outputPath,
			}
			var stderr bytes.Buffer
			if code := run(args, &stderr); code != 1 {
				t.Fatalf("run() code = %d, want 1; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "modgadget-fonts:") || !strings.Contains(stderr.String(), test.want) {
				t.Errorf("stderr = %q, want command prefix and %q", stderr.String(), test.want)
			}
			if strings.Count(strings.TrimSuffix(stderr.String(), "\n"), "\n") != 0 {
				t.Errorf("processing error spans multiple lines: %q", stderr.String())
			}
		})
	}
}

func TestRunFailureDoesNotChangeOrCreateOutput(t *testing.T) {
	dir := t.TempDir()
	bdfPath := writeTestFile(t, dir, "font.bdf", []byte(testBDF))
	subsetPath := writeTestFile(t, dir, "subset.txt", []byte("\u96ea"))
	existingPath := writeTestFile(t, dir, "existing.go", []byte("keep this"))
	newPath := filepath.Join(dir, "new.go")

	for _, outputPath := range []string{existingPath, newPath} {
		var stderr bytes.Buffer
		if code := run(validArgs(bdfPath, subsetPath, outputPath), &stderr); code != 1 {
			t.Fatalf("run() code = %d, want 1", code)
		}
	}
	existing, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(existing) != "keep this" {
		t.Fatalf("existing output changed to %q", existing)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatalf("new output exists or stat returned unexpected error: %v", err)
	}
}

func validArgs(bdfPath, subsetPath, outputPath string) []string {
	return []string{
		"-bdf", bdfPath,
		"-subset", subsetPath,
		"-package", "generated",
		"-var", "HeadlineFont",
		"-o", outputPath,
	}
}

func removeFlag(args []string, name string) []string {
	result := make([]string, 0, len(args)-2)
	for i := 0; i < len(args); i += 2 {
		if args[i] != name {
			result = append(result, args[i], args[i+1])
		}
	}
	return result
}

func writeTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
