// SPDX-License-Identifier: MIT
// Provenance: original
//
// ErrGolden writes the golden files of the internal/errdiagram tests from the
// upstream jar: for each case of testdata/cases.txt (the format is described
// there), the SVG image upstream writes (golden/<case>.svg) and its image
// data (golden/imagedata.txt: name, width, height, status).
//
// Build it on the host and run it in the reference image, whose fonts
// (DejaVu) are those of the Go font metrics. From the repository root:
//
//   mkdir -p ~/scratch-err/classes && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-err/classes internal/errdiagram/testdata/javagen/ErrGolden.java
//   docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" -v "$PWD:/work" \
//     -v ~/scratch-err/classes:/gen plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/gen ErrGolden /work/internal/errdiagram/testdata
//
// The time-based decorations of error images are disabled, as for the
// conformance references (design D16).

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

import net.sourceforge.plantuml.ErrorUml;
import net.sourceforge.plantuml.ErrorUmlType;
import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.FileFormatOption;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.core.ImageData;
import net.sourceforge.plantuml.core.UmlSource;
import net.sourceforge.plantuml.eggs.PSystemWelcome;
import net.sourceforge.plantuml.error.PSystemError;
import net.sourceforge.plantuml.error.PSystemErrorPreprocessor;
import net.sourceforge.plantuml.error.PSystemErrorV2;
import net.sourceforge.plantuml.error.PSystemUnsupported;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.geom.GraphicPosition;
import net.sourceforge.plantuml.preproc.PreprocessingArtifact;
import net.sourceforge.plantuml.text.StringLocated;
import net.sourceforge.plantuml.utils.LineLocation;
import net.sourceforge.plantuml.utils.LineLocationImpl;

public class ErrGolden {

	static class Case {
		String name;
		String kind = "error";
		boolean dark, red, noMetadata, noRaw;
		final List<StringLocated> lines = new ArrayList<>();
		final List<StringLocated> raw = new ArrayList<>();
		int trace = -1;
		String error;
		String assumed;
		// The location of the next source line.
		LineLocationImpl location = new LineLocationImpl("string", null);
		LineLocationImpl main;
	}

	public static void main(String[] args) throws IOException {
		PSystemError.disableTimeBasedErrorDecorations();
		final Path dir = Paths.get(args[0]);
		final Path golden = dir.resolve("golden");
		Files.createDirectories(golden);
		final StringBuilder imageData = new StringBuilder();
		for (Case c : parse(Files.readAllLines(dir.resolve("cases.txt"), StandardCharsets.UTF_8))) {
			final ByteArrayOutputStream os = new ByteArrayOutputStream();
			final ImageData data = render(c, os);
			Files.write(golden.resolve(c.name + ".svg"), os.toByteArray());
			imageData.append(c.name).append(' ').append(data.getWidth()).append(' ').append(data.getHeight())
					.append(' ').append(data.getStatus()).append('\n');
		}
		Files.write(golden.resolve("imagedata.txt"), imageData.toString().getBytes(StandardCharsets.UTF_8));
	}

	static ImageData render(Case c, ByteArrayOutputStream os) throws IOException {
		FileFormatOption option = new FileFormatOption(FileFormat.SVG, !c.noMetadata);
		if (c.red)
			option = option.withUseRedForError();
		if (c.dark)
			option = option.withColorMapper(ColorMapper.DARK_MODE);
		if (c.kind.equals("novalid"))
			return new SourceStringReader("").noValidStartFound(os, option);

		final boolean sequence = DiagramType.findStartTypes(c.lines.get(0).getString())
				.contains(DiagramType.SEQUENCE);
		final List<StringLocated> raw = c.noRaw ? new ArrayList<StringLocated>()
				: c.raw.isEmpty() ? c.lines : c.raw;
		final UmlSource source = UmlSource.createWithRaw(c.lines, sequence, raw);
		final PreprocessingArtifact artifact = new PreprocessingArtifact();
		final Diagram diagram;
		switch (c.kind) {
		case "error": {
			final List<StringLocated> trace = c.lines.subList(0, c.trace < 0 ? c.lines.size() : c.trace);
			final DiagramType assumed = c.assumed == null ? null : DiagramType.valueOf(c.assumed);
			final ErrorUml error = new ErrorUml(ErrorUmlType.SYNTAX_ERROR, c.error, 0, trace.get(trace.size() - 1),
					assumed);
			diagram = new PSystemErrorV2(source, trace, error, artifact, null);
			break;
		}
		case "preprocerror": {
			final List<StringLocated> input = new ArrayList<>(c.lines);
			final int last = input.size() - 1;
			input.set(last, new StringLocated(input.get(last).getString(), input.get(last).getLocation(), c.error));
			diagram = new PSystemErrorPreprocessor(input, input, artifact);
			break;
		}
		case "unsupported":
			diagram = new PSystemUnsupported(source, artifact);
			break;
		case "welcome":
			diagram = new PSystemWelcome(source, GraphicPosition.BACKGROUND_CORNER_BOTTOM_RIGHT, artifact);
			break;
		default:
			throw new IllegalArgumentException("unknown kind " + c.kind);
		}
		return diagram.exportDiagram(os, 0, option);
	}

	static List<Case> parse(List<String> file) {
		final List<Case> cases = new ArrayList<>();
		Case c = null;
		for (String line : file) {
			if (line.isEmpty() || line.startsWith("#"))
				continue;
			final int sp = line.indexOf(' ');
			final String key = sp < 0 ? line : line.substring(0, sp);
			final String arg = sp < 0 ? "" : unescape(line.substring(sp + 1));
			switch (key) {
			case "case":
				c = new Case();
				c.name = arg;
				c.main = c.location;
				break;
			case "kind":
				c.kind = arg;
				break;
			case "option":
				if (arg.equals("dark"))
					c.dark = true;
				else if (arg.equals("red"))
					c.red = true;
				else if (arg.equals("nometadata"))
					c.noMetadata = true;
				else
					throw new IllegalArgumentException("unknown option " + arg);
				break;
			case "line":
				addLine(c, arg);
				break;
			case "repeat": {
				final int n = Integer.parseInt(arg.substring(0, arg.indexOf(' ')));
				final String text = arg.substring(arg.indexOf(' ') + 1);
				for (int i = 1; i <= n; i++)
					addLine(c, text.replace("%d", Integer.toString(i)));
				break;
			}
			case "include": {
				final String[] parts = arg.split(" ");
				final LineLocation parent = c.lines.get(Integer.parseInt(parts[1])).getLocation();
				c.location = new LineLocationImpl(parts[0], parent);
				break;
			}
			case "endinclude":
				c.location = c.main;
				break;
			case "raw":
				c.raw.add(new StringLocated(arg, null));
				break;
			case "noraw":
				c.noRaw = true;
				break;
			case "trace":
				c.trace = Integer.parseInt(arg);
				break;
			case "error":
				c.error = arg;
				break;
			case "assumed":
				c.assumed = arg;
				break;
			case "end":
				cases.add(c);
				c = null;
				break;
			default:
				throw new IllegalArgumentException("unknown line " + line);
			}
		}
		return cases;
	}

	static void addLine(Case c, String text) {
		final LineLocationImpl next = c.location.oneLineRead();
		if (c.location == c.main)
			c.main = next;
		c.location = next;
		c.lines.add(new StringLocated(text, next));
	}

	static String unescape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char ch = s.charAt(i);
			if (ch != '\\' || i + 1 >= s.length()) {
				sb.append(ch);
				continue;
			}
			final char next = s.charAt(++i);
			if (next == 't') {
				sb.append('\t');
			} else if (next == 'u') {
				sb.append((char) Integer.parseInt(s.substring(i + 1, i + 5), 16));
				i += 4;
			} else {
				sb.append(next);
			}
		}
		return sb.toString();
	}
}
