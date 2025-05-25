package termlatex

import (
	"strings"
	"testing"
)

// All tests here are pure — no TeX backend, no TTY, no hardware required.
// Integration tests (actual rendering) live under the integration build tag and
// are skipped in CI unless a TeX backend is installed.

func TestBuildDoc_ContainsEquation(t *testing.T) {
	eq := `\frac{1}{2}`
	doc := buildDoc(eq, nil)
	if !strings.Contains(doc, eq) {
		t.Errorf("document does not contain equation %q\ndoc:\n%s", eq, doc)
	}
}

func TestBuildDoc_AlwaysIncoreAMSMath(t *testing.T) {
	doc := buildDoc("x", nil)
	for _, pkg := range []string{`\usepackage{amsmath}`, `\usepackage{amssymb}`, `\usepackage{amsfonts}`} {
		if !strings.Contains(doc, pkg) {
			t.Errorf("document missing %s", pkg)
		}
	}
}

func TestBuildDoc_ExtraPackages(t *testing.T) {
	doc := buildDoc("x", []string{"physics", "siunitx"})
	for _, pkg := range []string{`\usepackage{physics}`, `\usepackage{siunitx}`} {
		if !strings.Contains(doc, pkg) {
			t.Errorf("document missing extra package %s", pkg)
		}
	}
}

func TestBuildDoc_StartsWithDocumentClass(t *testing.T) {
	doc := buildDoc("x", nil)
	if !strings.HasPrefix(doc, `\documentclass`) {
		t.Errorf("document should start with \\documentclass, got: %q", doc[:40])
	}
}

func TestBuildDoc_HasBeginEndDocument(t *testing.T) {
	doc := buildDoc("x", nil)
	if !strings.Contains(doc, `\begin{document}`) || !strings.Contains(doc, `\end{document}`) {
		t.Errorf("document missing \\begin{document} / \\end{document}")
	}
}

func TestTrimLog_ShortOutput(t *testing.T) {
	out := []byte("line1\nline2\nline3")
	got := trimLog(out, 10)
	if got != "line1\nline2\nline3" {
		t.Errorf("trimLog short: got %q", got)
	}
}

func TestTrimLog_LongOutput(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	lines[49] = "last"
	out := []byte(strings.Join(lines, "\n"))
	got := trimLog(out, 5)
	if !strings.HasSuffix(got, "last") {
		t.Errorf("trimLog should end with last line, got: %q", got)
	}
	if len(strings.Split(got, "\n")) > 5 {
		t.Errorf("trimLog should return at most 5 lines")
	}
}

func TestOptions_DPIDefault(t *testing.T) {
	var o Options
	if o.dpi() != 150 {
		t.Errorf("default DPI = %d, want 150", o.dpi())
	}
}

func TestOptions_DPIExplicit(t *testing.T) {
	o := Options{DPI: 300}
	if o.dpi() != 300 {
		t.Errorf("explicit DPI = %d, want 300", o.dpi())
	}
}

func TestBackendString(t *testing.T) {
	cases := map[Backend]string{
		PDFLaTeX: "pdflatex",
		Tectonic: "tectonic",
		DVIPng:   "latex+dvipng",
		Auto:     "auto",
	}
	for b, want := range cases {
		if got := b.String(); got != want {
			t.Errorf("Backend(%d).String() = %q, want %q", b, got, want)
		}
	}
}

func TestDisplayWrapsEquation(t *testing.T) {
	eq := `x^2`
	// buildDoc must contain the display-wrapped equation
	wrapped := `\[\displaystyle ` + eq + `\]`
	doc := buildDoc(wrapped, nil)
	if !strings.Contains(doc, wrapped) {
		t.Errorf("Display-wrapped equation not found in document")
	}
}

func TestInlineWrapsEquation(t *testing.T) {
	eq := `x^2`
	wrapped := `$` + eq + `$`
	doc := buildDoc(wrapped, nil)
	if !strings.Contains(doc, wrapped) {
		t.Errorf("Inline-wrapped equation not found in document")
	}
}
