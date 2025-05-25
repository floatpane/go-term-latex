package termlatex

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func renderPDFLaTeX(equation string, opts Options) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "termlatex-*")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	texFile := filepath.Join(tmpDir, "eq.tex")
	if err := os.WriteFile(texFile, []byte(buildDoc(equation, opts.Packages)), 0o600); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}

	cmd := exec.Command("pdflatex",
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-output-directory", tmpDir,
		texFile,
	)
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: pdflatex: %w\n%s", ErrRenderFailed, err, trimLog(out, 20))
	}

	pdfFile := filepath.Join(tmpDir, "eq.pdf")
	outBase := filepath.Join(tmpDir, "out")

	cmd = exec.Command("pdftoppm",
		"-png",
		"-r", strconv.Itoa(opts.dpi()),
		"-singlefile",
		pdfFile, outBase,
	)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: pdftoppm: %w\n%s", ErrRenderFailed, err, trimLog(out, 10))
	}

	return readFirstPNG(outBase)
}
