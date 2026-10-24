// SPDX-License-Identifier: MIT
// Provenance: original
//
// SvekGolden renders the class diagram cases of internal/svek/testdata/cases
// with the upstream jar in the reference image (Graphviz 14.0.1), for the
// svek golden tests (internal/svek/golden_test.go). For each case NAME.puml it
// writes, in the output directory:
//
//   NAME.svg       the SVG of the jar
//   NAME.dot       the DOT graph svek gave to dot (upstream's svek trace)
//   NAME.dot14.svg the SVG that dot 14.0.1 returned for it
//
// The trace comes from a second render with "!pragma svek_trace on" added
// after the start line; upstream writes svek.dot and svek.svg in the current
// directory, which this program renames. From the repository root:
//
//   mkdir -p ~/scratch-svek && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-svek internal/svek/testdata/javagen/SvekGolden.java
//   mkdir -p .cache/svek-work && docker run --rm --platform linux/amd64 -u "$(id -u):$(id -g)" \
//     -v "$PWD:/repo" -v "$HOME/scratch-svek:/classes" -w /repo/.cache/svek-work \
//     plantuml-go-reference java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions \
//     -XX:-InlineMathNatives -cp /opt/plantuml.jar:/classes SvekGolden \
//     /repo/internal/svek/testdata/cases /repo/internal/svek/testdata/golden

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.FileFormatOption;
import net.sourceforge.plantuml.SourceStringReader;

public class SvekGolden {

	public static void main(String[] args) throws IOException {
		final Path in = Paths.get(args[0]);
		final Path out = Paths.get(args[1]);
		Files.createDirectories(out);
		final List<Path> cases = new ArrayList<>();
		try (DirectoryStream<Path> ds = Files.newDirectoryStream(in, "*.puml")) {
			for (Path p : ds)
				cases.add(p);
		}
		Collections.sort(cases);
		for (Path p : cases) {
			final String name = p.getFileName().toString().replaceFirst("\\.puml$", "");
			final String source = new String(Files.readAllBytes(p), StandardCharsets.UTF_8);
			Files.write(out.resolve(name + ".svg"), render(source));

			final Path dot = Paths.get("svek.dot");
			final Path svg = Paths.get("svek.svg");
			Files.deleteIfExists(dot);
			Files.deleteIfExists(svg);
			final int eol = source.indexOf('\n');
			render(source.substring(0, eol + 1) + "!pragma svek_trace on\n" + source.substring(eol + 1));
			if (Files.exists(dot))
				Files.move(dot, out.resolve(name + ".dot"), StandardCopyOption.REPLACE_EXISTING);
			else
				Files.deleteIfExists(out.resolve(name + ".dot"));
			if (Files.exists(svg))
				Files.move(svg, out.resolve(name + ".dot14.svg"), StandardCopyOption.REPLACE_EXISTING);
			else
				Files.deleteIfExists(out.resolve(name + ".dot14.svg"));
			System.err.println(name);
		}
	}

	private static byte[] render(String source) throws IOException {
		final ByteArrayOutputStream os = new ByteArrayOutputStream();
		new SourceStringReader(source).outputImage(os, new FileFormatOption(FileFormat.SVG));
		return os.toByteArray();
	}
}
