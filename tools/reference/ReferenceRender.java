// SPDX-License-Identifier: MIT
// Provenance: original

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import net.sourceforge.plantuml.BlockUml;
import net.sourceforge.plantuml.ErrorUml;
import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.FileFormatOption;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.TitledDiagram;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.core.ImageData;
import net.sourceforge.plantuml.error.PSystemError;
import net.sourceforge.plantuml.preproc.Defines;
import net.sourceforge.plantuml.security.SFile;

/**
 * Renders conformance cases with Java PlantUML used as a library, recording
 * per-case results that the CLI cannot report separately in batch mode.
 *
 * Input is a manifest with one case per line: id TAB sourceFile TAB formats,
 * where formats is a comma-separated subset of svg, png, txt, utxt. For each
 * case it writes <outRoot>/<id>/meta.json and block<i>-page<p>.<ext> files.
 * PNG files go to <pngRoot> instead, when given, and are skipped otherwise.
 *
 * Usage: java ReferenceRender <manifest> <outRoot> [pngRoot]
 */
public class ReferenceRender {

	public static void main(String[] args) throws IOException {
		if (args.length < 2 || args.length > 3) {
			System.err.println("usage: java ReferenceRender <manifest> <outRoot> [pngRoot]");
			System.exit(2);
		}
		// Upstream adds sponsor messages to error images during some minutes of
		// each hour; references must not depend on the clock.
		PSystemError.disableTimeBasedErrorDecorations();
		final File outRoot = new File(args[1]);
		final File pngRoot = args.length == 3 ? new File(args[2]) : null;
		int count = 0;
		for (String line : Files.readAllLines(new File(args[0]).toPath(), StandardCharsets.UTF_8)) {
			if (line.isEmpty())
				continue;
			final String[] parts = line.split("\t", -1);
			if (parts.length != 3)
				throw new IllegalArgumentException("bad manifest line: " + line);
			renderCase(parts[0], new File(parts[1]), parts[2].split(","), outRoot, pngRoot);
			count++;
		}
		System.err.println("ReferenceRender: " + count + " case(s) rendered");
	}

	static void renderCase(String id, File source, String[] formats, File outRoot, File pngRoot) throws IOException {
		final File out = new File(outRoot, id);
		out.mkdirs();
		final StringBuilder json = new StringBuilder();
		json.append("{\n  \"id\": ").append(str(id)).append(",\n  \"blocks\": [");
		try {
			final String text = new String(Files.readAllBytes(source.toPath()), StandardCharsets.UTF_8);
			final SourceStringReader reader = new SourceStringReader(Defines.createWithFileName(source), text,
					StandardCharsets.UTF_8, Collections.<String>emptyList(),
					SFile.fromFile(source.getAbsoluteFile().getParentFile()));
			final List<BlockUml> blocks = reader.getBlocks();
			for (int i = 0; i < blocks.size(); i++) {
				json.append(i == 0 ? "\n" : ",\n");
				renderBlock(json, i, blocks.get(i), formats, out, pngRoot == null ? null : new File(pngRoot, id));
			}
			json.append(blocks.isEmpty() ? "]" : "\n  ]");
		} catch (Exception e) {
			// A diagram-level failure is part of the reference. Errors such as
			// NoClassDefFoundError are not caught: they mean the setup is broken.
			json.append("],\n  \"crash\": ").append(str(e.toString()));
		}
		json.append("\n}\n");
		try (Writer w = new OutputStreamWriter(new FileOutputStream(new File(out, "meta.json")),
				StandardCharsets.UTF_8)) {
			w.write(json.toString());
		}
	}

	static void renderBlock(StringBuilder json, int index, BlockUml block, String[] formats, File out, File pngOut)
			throws IOException {
		final Diagram diagram = block.getDiagram();
		final String type;
		if (diagram instanceof TitledDiagram)
			type = ((TitledDiagram) diagram).getDiagramType().name();
		else if (diagram instanceof PSystemError)
			type = "ERROR";
		else
			type = "OTHER";

		json.append("    {\"index\": ").append(index);
		json.append(", \"name\": ").append(block.getFileOrDirname() == null ? "null" : str(block.getFileOrDirname()));
		json.append(", \"type\": ").append(str(type));
		json.append(", \"description\": ").append(str(diagram.getDescription().getDescription()));
		json.append(", \"pages\": ").append(diagram.getNbImages());
		if (diagram instanceof PSystemError) {
			final PSystemError err = (PSystemError) diagram;
			json.append(", \"error\": {\"line\": ").append(err.getLineLocation().getPosition() + 1);
			json.append(", \"messages\": [");
			final List<String> messages = new ArrayList<>();
			for (ErrorUml e : err.getErrorsUml())
				messages.add(str(e.getError()));
			json.append(String.join(", ", messages)).append("]}");
		}
		json.append(", \"renders\": [");
		boolean first = true;
		for (String format : formats) {
			final FileFormat ff = fileFormat(format);
			if (ff == null)
				continue;
			final File dir = ff == FileFormat.PNG ? pngOut : out;
			if (dir == null)
				continue;
			dir.mkdirs();
			for (int page = 0; page < diagram.getNbImages(); page++) {
				final ByteArrayOutputStream os = new ByteArrayOutputStream();
				final ImageData data = diagram.exportDiagram(os, page, new FileFormatOption(ff));
				final String name = "block" + index + "-page" + page + ff.getFileSuffix();
				// Status 503 is upstream's crash report, which shows a random
				// quote and picture: it is not reproducible, so no file is kept.
				final boolean crashReport = data.getStatus() == 503;
				if (!crashReport)
					Files.write(new File(dir, name).toPath(), os.toByteArray());
				json.append(first ? "" : ", ");
				json.append("{\"file\": ").append(crashReport ? "null" : str(name));
				if (crashReport)
					json.append(", \"crashReport\": true");
				json.append(", \"width\": ").append(data.getWidth());
				json.append(", \"height\": ").append(data.getHeight());
				json.append(", \"status\": ").append(data.getStatus()).append("}");
				first = false;
			}
		}
		json.append("]}");
	}

	static FileFormat fileFormat(String format) {
		switch (format) {
		case "svg":
			return FileFormat.SVG;
		case "png":
			return FileFormat.PNG;
		case "txt":
			return FileFormat.ATXT;
		case "utxt":
			return FileFormat.UTXT;
		default:
			return null; // preproc is produced by the CLI
		}
	}

	static String str(String s) {
		final StringBuilder sb = new StringBuilder("\"");
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '"' || c == '\\')
				sb.append('\\').append(c);
			else if (c < 0x20 || c > 0x7e)
				sb.append(String.format("\\u%04x", (int) c));
			else
				sb.append(c);
		}
		return sb.append('"').toString();
	}
}
