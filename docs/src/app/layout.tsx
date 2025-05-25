import type { Metadata } from "next";
import "./globals.css";
import "highlight.js/styles/github-dark.css";

export const metadata: Metadata = {
	title: "go-term-latex",
	description:
		"Render LaTeX math equations in the terminal. Kitty graphics, Sixel, and Unicode half-block fallback. Supports pdflatex, tectonic, and dvipng backends.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en">
			<body>
				<header className="site-header">
					<a href="/" className="brand">
						go-term-latex
					</a>
					<nav>
						<a href="https://github.com/floatpane/go-term-latex">GitHub</a>
					</nav>
				</header>
				<main>{children}</main>
			</body>
		</html>
	);
}
