package mapper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// generate runs the generator over the package in testdata/<name> and returns the
// generated source, or the error of the first interface that failed.
func generate(t *testing.T, name string) (string, error) {
	t.Helper()

	dir := filepath.Join("testdata", name)
	interfaces, err := NewParser().Parse(dir, ".")
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", dir, err)
	}

	gen := NewGenerator("gen", t.TempDir())
	gen.SetSrcDir(dir)
	for _, iface := range interfaces {
		if err := gen.GenerateInterface(iface); err != nil {
			return "", err
		}
	}

	return gen.String()
}

// mustGenerate is generate for a package that is expected to generate cleanly. The
// result is written to testdata/<name>/gen and has to type-check there.
func mustGenerate(t *testing.T, name string) string {
	t.Helper()

	out, err := generate(t, name)
	if err != nil {
		t.Fatalf("generating %s: %v", name, err)
	}

	dir := filepath.Join("testdata", name, "gen")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "gen.go"), []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes, Dir: dir}, ".")
	if err != nil {
		t.Fatalf("loading generated %s: %v", name, err)
	}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, e := range pkg.Errors {
			t.Errorf("generated %s does not compile: %v\n%s", name, e, out)
		}
	})

	return out
}

// method returns the generated source of one method, from its signature up to the
// next top-level declaration.
func method(t *testing.T, src, name string) string {
	t.Helper()

	start := strings.Index(src, ") "+name+"(")
	if start < 0 {
		t.Fatalf("method %s not generated in:\n%s", name, src)
	}
	start = strings.LastIndex(src[:start], "\nfunc ") + 1

	end := strings.Index(src[start+1:], "\nfunc ")
	if end < 0 {
		return src[start:]
	}

	return src[start : start+1+end]
}

// assertContains fails unless every line in want appears in got, once whitespace
// is collapsed.
func assertContains(t *testing.T, got string, want ...string) {
	t.Helper()

	squashed := squashSpace(got)
	for _, w := range want {
		if !strings.Contains(squashed, squashSpace(w)) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
}

func squashSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
