package termlatex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func renderDVIPng(ctx context.Context, equation string, opts Options) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "termlatex-*")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	texFile := filepath.Join(tmpDir, "eq.tex")
	if err := os.WriteFile(texFile, []byte(buildDoc(equation, opts.Packages)), 0o600); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}

	cmd := exec.CommandContext(ctx, "latex",
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-output-directory", tmpDir,
		texFile,
	)
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: latex: %w\n%s", ErrRenderFailed, err, trimLog(out, 20))
	}

	dviFile := filepath.Join(tmpDir, "eq.dvi")
	pngFile := filepath.Join(tmpDir, "out.png")

	// dvipng -T tight crops tightly to the content; -bg Transparent works in
	// many terminals but we default to white for maximum compatibility.
	cmd = exec.CommandContext(ctx, "dvipng",
		"-T", "tight",
		"-D", strconv.Itoa(opts.dpi()),
		"-bg", "White",
		"-fg", "Black",
		"-o", pngFile,
		dviFile,
	)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: dvipng: %w\n%s", ErrRenderFailed, err, trimLog(out, 10))
	}

	return os.ReadFile(pngFile)
}
