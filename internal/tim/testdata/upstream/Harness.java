// SPDX-License-Identifier: MIT
// Provenance: original
//
// Harness runs the preprocessor of the reference PlantUML jar (v1.2026.9, MIT
// variant) on the case files of this directory and writes the results as JSON.
// The Go tests compare the Go preprocessor with those results.
//
// Regenerate upstream.json in the reference image with:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/tim
//
// Usage: java -cp plantuml.jar:<classes> Harness <out.json> <workdir> <case files...>
//
// Case file format: "=== name" starts a case. Header lines start with "--- ":
//
//	--- profile: SANDBOX       security profile (default: unset, so LEGACY)
//	--- define: NAME=value     a -D definition
//	--- includepath: dir       plantuml.include.path, relative to the case dir
//	--- dir: sub               current directory, relative to the case dir
//	--- escaped                \\uXXXX escapes are decoded in the following lines
//	--- file: path             the following lines are a fixture file
//	--- main                   the following lines are the diagram source
//
// A main source without an @start line is wrapped in @startuml/@enduml.
//
// Files ending with .lines hold one line per line (with \\uXXXX escapes):
// their TLineType goes to the "lineTypes" list of the output.

import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.io.StringReader;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import net.sourceforge.plantuml.BlockUml;
import net.sourceforge.plantuml.BlockUmlBuilder;
import net.sourceforge.plantuml.FileSystem;
import net.sourceforge.plantuml.log.Logme;
import net.sourceforge.plantuml.preproc.Defines;
import net.sourceforge.plantuml.preproc.OptionKey;
import net.sourceforge.plantuml.preproc.PreprocessingArtifact;
import net.sourceforge.plantuml.security.SFile;
import net.sourceforge.plantuml.security.SecurityUtils;
import net.sourceforge.plantuml.text.StringLocated;
import net.sourceforge.plantuml.text.TLineType;
import net.sourceforge.plantuml.utils.LineLocation;
import net.sourceforge.plantuml.warning.Warning;

public class Harness {

	static final String WORK = "/work";

	static class Case {
		String name;
		String profile;
		String includePath;
		String dir = "";
		final Map<String, String> defines = new LinkedHashMap<>();
		final Map<String, List<String>> files = new LinkedHashMap<>();
		final List<String> main = new ArrayList<>();
	}

	public static void main(String[] args) throws Exception {
		Logme.MODE_VEGA = true;
		final Path out = Paths.get(args[0]);
		final Path work = Paths.get(args[1]).toAbsolutePath();
		final List<Case> cases = new ArrayList<>();
		final List<String> lines = new ArrayList<>();
		for (int i = 2; i < args.length; i++)
			if (args[i].endsWith(".lines"))
				for (String line : Files.readAllLines(Paths.get(args[i]), StandardCharsets.UTF_8))
					lines.add(unescape(line));
			else
				cases.addAll(parse(Paths.get(args[i])));
		final StringBuilder sb = new StringBuilder();
		sb.append("{\"lineTypes\": [");
		for (int i = 0; i < lines.size(); i++) {
			if (i > 0)
				sb.append(",");
			final String type = TLineType.getFromLineInternal(new StringLocated(lines.get(i), null)).name();
			sb.append("\n {\"s\": ").append(q(lines.get(i))).append(", \"type\": ").append(q(type)).append("}");
		}
		sb.append("],\n\"cases\": [\n");
		for (int i = 0; i < cases.size(); i++) {
			if (i > 0)
				sb.append(",\n");
			run(sb, cases.get(i), work.resolve(String.format("c%04d", i)));
		}
		sb.append("\n]}\n");
		Files.write(out, sb.toString().getBytes(StandardCharsets.UTF_8));
	}

	static List<Case> parse(Path file) throws Exception {
		final List<Case> result = new ArrayList<>();
		Case current = null;
		List<String> target = null;
		boolean escaped = false;
		for (String line : Files.readAllLines(file, StandardCharsets.UTF_8)) {
			if (line.startsWith("=== ")) {
				current = new Case();
				current.name = line.substring(4).trim();
				result.add(current);
				target = null;
				escaped = false;
				continue;
			}
			if (current == null)
				continue;
			if (line.startsWith("--- ")) {
				final String h = line.substring(4);
				if (h.startsWith("profile:"))
					current.profile = h.substring(8).trim();
				else if (h.startsWith("define:")) {
					final String d = h.substring(7).trim();
					final int eq = d.indexOf('=');
					current.defines.put(d.substring(0, eq), d.substring(eq + 1));
				} else if (h.startsWith("includepath:"))
					current.includePath = h.substring(12).trim();
				else if (h.startsWith("dir:"))
					current.dir = h.substring(4).trim();
				else if (h.equals("escaped"))
					escaped = true;
				else if (h.startsWith("file:")) {
					target = new ArrayList<>();
					current.files.put(h.substring(5).trim(), target);
				} else if (h.equals("main"))
					target = current.main;
				else
					throw new IllegalArgumentException("bad header " + line);
				continue;
			}
			if (target == null)
				throw new IllegalArgumentException("line outside of a section in " + current.name + ": " + line);
			target.add(escaped ? unescape(line) : line);
		}
		for (Case c : result) {
			boolean hasStart = false;
			for (String s : c.main)
				if (s.trim().startsWith("@start"))
					hasStart = true;
			if (!hasStart) {
				c.main.add(0, "@startuml");
				c.main.add("@enduml");
			}
		}
		return result;
	}

	static String unescape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '\\' && i + 1 < s.length() && s.charAt(i + 1) == '\\') {
				sb.append('\\');
				i++;
			} else if (c == '\\' && i + 5 < s.length() && s.charAt(i + 1) == 'u') {
				sb.append((char) Integer.parseInt(s.substring(i + 2, i + 6), 16));
				i += 5;
			} else {
				sb.append(c);
			}
		}
		return sb.toString();
	}

	static String join(List<String> lines) {
		final StringBuilder sb = new StringBuilder();
		for (String s : lines)
			sb.append(s).append('\n');
		return sb.toString();
	}

	static void run(StringBuilder sb, Case c, Path dir) throws Exception {
		Files.createDirectories(dir);
		for (Map.Entry<String, List<String>> ent : c.files.entrySet()) {
			final Path p = dir.resolve(ent.getKey());
			Files.createDirectories(p.getParent());
			Files.write(p, join(ent.getValue()).getBytes(StandardCharsets.UTF_8));
		}
		final Path cwd = dir.resolve(c.dir);
		Files.createDirectories(cwd);
		final String real = dir.toRealPath().toString();
		final String abs = dir.toAbsolutePath().toString();

		if (c.profile == null)
			System.clearProperty("PLANTUML_SECURITY_PROFILE");
		else
			System.setProperty("PLANTUML_SECURITY_PROFILE", c.profile);
		if (c.includePath == null)
			System.clearProperty("plantuml.include.path");
		else
			System.setProperty("plantuml.include.path", dir.resolve(c.includePath).toString());
		final Field current = SecurityUtils.class.getDeclaredField("current");
		current.setAccessible(true);
		current.set(null, null);

		FileSystem.getInstance().setCurrentDir(new SFile(cwd.toString()));
		final Defines defines = Defines.createEmpty();
		for (Map.Entry<String, String> ent : c.defines.entrySet())
			defines.define(ent.getKey(), ent.getValue());

		final ByteArrayOutputStream err = new ByteArrayOutputStream();
		final PrintStream oldErr = System.err;
		System.setErr(new PrintStream(err, true, "UTF-8"));
		List<BlockUml> blocks = Collections.emptyList();
		String crash = null;
		try {
			final BlockUmlBuilder builder = new BlockUmlBuilder(Collections.<String>emptyList(),
					StandardCharsets.UTF_8, defines, new StringReader(join(c.main)), new SFile(cwd.toString()),
					"string");
			blocks = builder.getBlockUmls();
		} catch (Throwable t) {
			crash = t.getClass().getName();
		} finally {
			System.setErr(oldErr);
		}

		sb.append("{\"name\": ").append(q(c.name));
		if (c.profile != null)
			sb.append(", \"profile\": ").append(q(c.profile));
		if (c.includePath != null)
			sb.append(", \"includePath\": ").append(q(c.includePath));
		if (!c.dir.isEmpty())
			sb.append(", \"dir\": ").append(q(c.dir));
		if (!c.defines.isEmpty()) {
			sb.append(", \"defines\": {");
			boolean first = true;
			for (Map.Entry<String, String> ent : c.defines.entrySet()) {
				if (!first)
					sb.append(", ");
				first = false;
				sb.append(q(ent.getKey())).append(": ").append(q(ent.getValue()));
			}
			sb.append("}");
		}
		if (!c.files.isEmpty()) {
			sb.append(",\n \"files\": {");
			boolean first = true;
			for (Map.Entry<String, List<String>> ent : c.files.entrySet()) {
				if (!first)
					sb.append(", ");
				first = false;
				sb.append(q(ent.getKey())).append(": ").append(q(join(ent.getValue())));
			}
			sb.append("}");
		}
		sb.append(",\n \"main\": ").append(q(join(c.main)));
		if (crash != null)
			sb.append(",\n \"crash\": ").append(q(crash));
		final String log = new String(err.toByteArray(), StandardCharsets.UTF_8);
		if (!log.isEmpty())
			sb.append(",\n \"stderr\": ").append(q(clean(log, real, abs)));
		sb.append(",\n \"blocks\": [");
		for (int i = 0; i < blocks.size(); i++) {
			if (i > 0)
				sb.append(",");
			block(sb, blocks.get(i), real, abs);
		}
		sb.append("]}");
	}

	static String clean(String s, String real, String abs) {
		if (s == null)
			return null;
		return s.replace(real, WORK).replace(abs, WORK);
	}

	@SuppressWarnings("unchecked")
	static void block(StringBuilder sb, BlockUml b, String real, String abs) throws Exception {
		final Field fDebug = BlockUml.class.getDeclaredField("debug");
		fDebug.setAccessible(true);
		final Field fErr = BlockUml.class.getDeclaredField("preprocessorError");
		fErr.setAccessible(true);
		final Field fArt = BlockUml.class.getDeclaredField("preprocessingArtifact");
		fArt.setAccessible(true);
		sb.append("\n  {\"error\": ").append(fErr.getBoolean(b));
		sb.append(",\n   \"lines\": ");
		lines(sb, b.getData(), real, abs);
		sb.append(",\n   \"debug\": ");
		lines(sb, (List<StringLocated>) fDebug.get(b), real, abs);
		final PreprocessingArtifact art = (PreprocessingArtifact) fArt.get(b);
		final StringBuilder opts = new StringBuilder();
		for (OptionKey key : OptionKey.values()) {
			final String v = art.getOption().getValue(key);
			if (v == null)
				continue;
			if (opts.length() > 0)
				opts.append(", ");
			opts.append(q(key.name())).append(": ").append(q(v));
		}
		if (opts.length() > 0)
			sb.append(",\n   \"options\": {").append(opts).append("}");
		if (!art.getWarnings().isEmpty()) {
			sb.append(",\n   \"warnings\": [");
			boolean first = true;
			for (Warning w : art.getWarnings()) {
				if (!first)
					sb.append(", ");
				first = false;
				sb.append(q(w.asSingleLine()));
			}
			sb.append("]");
		}
		sb.append("}");
	}

	static void lines(StringBuilder sb, List<StringLocated> lines, String real, String abs) {
		sb.append("[");
		if (lines != null)
			for (int i = 0; i < lines.size(); i++) {
				final StringLocated sl = lines.get(i);
				if (i > 0)
					sb.append(",");
				sb.append("\n    {\"s\": ").append(q(clean(sl.getString(), real, abs)));
				sb.append(", \"loc\": ").append(q(clean(location(sl.getLocation()), real, abs)));
				if (sl.getPreprocessorError() != null)
					sb.append(", \"err\": ").append(q(clean(sl.getPreprocessorError(), real, abs)));
				sb.append("}");
			}
		sb.append("]");
	}

	static String location(LineLocation loc) {
		if (loc == null)
			return "null";
		final StringBuilder sb = new StringBuilder();
		sb.append(loc.getDescription()).append(':').append(loc.getPosition());
		if (loc.getParent() != null)
			sb.append(" < ").append(location(loc.getParent()));
		return sb.toString();
	}

	static String q(String s) {
		final StringBuilder sb = new StringBuilder("\"");
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '"')
				sb.append("\\\"");
			else if (c == '\\')
				sb.append("\\\\");
			else if (c == '\n')
				sb.append("\\n");
			else if (c < 0x20 || c >= 0x7f)
				sb.append(String.format("\\u%04x", (int) c));
			else
				sb.append(c);
		}
		return sb.append('"').toString();
	}
}
