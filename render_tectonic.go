package termlatex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func renderTectonic(ctx context.Context, equation string, opts Options) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "termlatex-*")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	texFile := filepath.Join(tmpDir, "eq.tex")
	if err := os.WriteFile(texFile, []byte(buildDoc(equation, opts.Packages)), 0o600); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}

	// Tectonic writes output next to the input file; --outdir puts it in tmpDir.
	cmd := exec.CommandContext(ctx, "tectonic", "--outdir", tmpDir, texFile)
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: tectonic: %w\n%s", ErrRenderFailed, err, trimLog(out, 20))
	}

	pdfFile := filepath.Join(tmpDir, "eq.pdf")
	outBase := filepath.Join(tmpDir, "out")

	cmd = exec.CommandContext(ctx, "pdftoppm",
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
