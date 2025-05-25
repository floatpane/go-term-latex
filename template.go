package termlatex

import (
	"fmt"
	"strings"
)

// buildDoc wraps equation in a minimal standalone LaTeX document. The
// standalone class with the preview option crops the output tightly to the
// content with a small border, which is exactly what terminal rendering needs.
func buildDoc(equation string, extraPackages []string) string {
	var b strings.Builder
	b.WriteString(`\documentclass[preview,border=4pt]{standalone}` + "\n")
	b.WriteString(`\usepackage{amsmath}` + "\n")
	b.WriteString(`\usepackage{amssymb}` + "\n")
	b.WriteString(`\usepackage{amsfonts}` + "\n")
	for _, pkg := range extraPackages {
		fmt.Fprintf(&b, `\usepackage{%s}`+"\n", pkg)
	}
	b.WriteString(`\begin{document}` + "\n")
	b.WriteString(equation + "\n")
	b.WriteString(`\end{document}` + "\n")
	return b.String()
}

// trimLog returns the last n lines of a tool's combined stdout/stderr, useful
// for keeping error messages readable without dumping an entire TeX log.
func trimLog(out []byte, n int) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) <= n {
		return strings.TrimSpace(string(out))
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
