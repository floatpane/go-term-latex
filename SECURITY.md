# Security Policy

## Supported Versions

Only the latest release of go-term-latex is supported with security updates.

## Reporting a Vulnerability

Please report vulnerabilities privately — **do not open a public issue.**

Email [us@floatpane.com](mailto:us@floatpane.com) with a description, steps to
reproduce, potential impact, and optional suggested fix. We'll acknowledge
within 48 hours and aim for a fix within 7 days.

## Shell Escape

`pdflatex` and `latex` are invoked **without** `-shell-escape`. This means the
compiled document cannot execute arbitrary shell commands via `\write18` or
`\immediate\write18`.

> [!CAUTION]
> **Do not pass untrusted LaTeX to this library without sanitization.** Even
> without `-shell-escape`, TeX can read local files (`\input`, `\include`,
> `\openin`) and exfiltrate their contents into the PDF if an attacker controls
> the equation. Never render LaTeX from untrusted sources (user input, network
> data) without first validating or restricting the document.

## Scope

Of particular interest:

- **Shell injection** — any path where the equation string or a caller-supplied
  option is interpolated unsanitized into a shell command or file path,
  enabling command injection.
- **Path traversal** — temp-file naming that allows an attacker-controlled
  equation to overwrite files outside the temp directory.
- **TeX file read exfiltration** — document templates that allow `\input` of
  arbitrary local files.
- **Panics on malformed input** — equation strings that crash the library
  before the TeX backend is even invoked.
- **Temp-dir cleanup failures** — conditions where large temp files (malicious
  or accidental) are not cleaned up, causing disk exhaustion.

## Disclosure

We ask for reasonable time to address issues before public disclosure. We will
credit reporters in release notes unless you prefer to remain anonymous.
