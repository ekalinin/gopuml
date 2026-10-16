// SPDX-License-Identifier: MIT
// Provenance: original

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.font.FontRenderContext;
import java.awt.font.LineMetrics;
import java.awt.image.BufferedImage;
import java.io.BufferedOutputStream;
import java.io.DataOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

/**
 * Dumps the text metrics that Java PlantUML uses, so the Go implementation can
 * measure text identically without a JVM.
 *
 * Measurement follows upstream's FileFormat.getJavaDimension: a Graphics2D on a
 * dummy BufferedImage with text antialiasing and fractional metrics on, and
 * FontMetrics.getStringBounds on Java's logical fonts.
 *
 * Output, for each logical family and style, a big-endian binary table
 * "<family>-<style>.bin":
 *
 *   magic "PUMF", u16 version (1), UTF family, UTF style, f32 refSize,
 *   f32 ascent, f32 descent, f32 leading, f32 boundsHeight, f32 missingAdvance,
 *   i32 rangeCount, then per range: i32 firstCodePoint, i32 count, count x f32.
 *
 * Advances are measured at refSize and scale linearly with the font size. Code
 * points not in any range use missingAdvance. The tool also writes
 * summary.json and a samples file of strings measured directly at their size,
 * used to test the Go implementation.
 *
 * Usage: java FontMetricsDump <tableDir> <samplesFile>
 */
public class FontMetricsDump {

	static final float REF_SIZE = 100f;
	static final String[] FAMILIES = { "SansSerif", "Serif", "Monospaced", "Dialog" };
	static final int[] STYLES = { Font.PLAIN, Font.BOLD, Font.ITALIC, Font.BOLD | Font.ITALIC };
	static final String[] STYLE_NAMES = { "plain", "bold", "italic", "bolditalic" };
	static final int SAMPLE_COUNT = 1000;
	static final long SAMPLE_SEED = 20260905L;

	static Graphics2D gg;

	public static void main(String[] args) throws IOException {
		if (args.length != 2) {
			System.err.println("usage: java FontMetricsDump <tableDir> <samplesFile>");
			System.exit(2);
		}
		final File tableDir = new File(args[0]);
		final File samplesFile = new File(args[1]);
		tableDir.mkdirs();
		samplesFile.getParentFile().mkdirs();

		final BufferedImage im = new BufferedImage(100, 100, BufferedImage.TYPE_INT_RGB);
		gg = im.createGraphics();
		gg.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
		gg.setRenderingHint(RenderingHints.KEY_FRACTIONALMETRICS, RenderingHints.VALUE_FRACTIONALMETRICS_ON);

		final List<Table> tables = new ArrayList<>();
		for (String family : FAMILIES)
			for (int i = 0; i < STYLES.length; i++) {
				final Table t = measureTable(family, STYLES[i], STYLE_NAMES[i]);
				t.write(new File(tableDir, family.toLowerCase() + "-" + STYLE_NAMES[i] + ".bin"));
				tables.add(t);
			}

		final double maxError = writeSamples(samplesFile, tables);
		writeSummary(new File(tableDir, "summary.json"), tables, maxError);
	}

	static final class Range {
		final int first;
		final List<Float> advances = new ArrayList<>();

		Range(int first) {
			this.first = first;
		}
	}

	static final class Table {
		String family, style;
		float ascent, descent, leading, boundsHeight, missingAdvance;
		int missingCodePoint;
		final List<Range> ranges = new ArrayList<>();
		int glyphs;

		float advance(int cp) {
			for (Range r : ranges)
				if (cp >= r.first && cp < r.first + r.advances.size())
					return r.advances.get(cp - r.first);
			return missingAdvance;
		}

		void write(File f) throws IOException {
			try (DataOutputStream out = new DataOutputStream(new BufferedOutputStream(new FileOutputStream(f)))) {
				out.writeBytes("PUMF");
				out.writeShort(1);
				out.writeUTF(family);
				out.writeUTF(style);
				out.writeFloat(REF_SIZE);
				out.writeFloat(ascent);
				out.writeFloat(descent);
				out.writeFloat(leading);
				out.writeFloat(boundsHeight);
				out.writeFloat(missingAdvance);
				out.writeInt(ranges.size());
				for (Range r : ranges) {
					out.writeInt(r.first);
					out.writeInt(r.advances.size());
					for (float a : r.advances)
						out.writeFloat(a);
				}
			}
		}
	}

	static float width(Font font, String text) {
		final FontMetrics fm = gg.getFontMetrics(font);
		return (float) fm.getStringBounds(text, gg).getWidth();
	}

	static Table measureTable(String family, int style, String styleName) {
		final Font font = new Font(family, style, (int) REF_SIZE);
		final FontRenderContext frc = gg.getFontRenderContext();
		final Table t = new Table();
		t.family = family;
		t.style = styleName;

		final LineMetrics lm = font.getLineMetrics("Xg", frc);
		t.ascent = lm.getAscent();
		t.descent = lm.getDescent();
		t.leading = lm.getLeading();
		t.boundsHeight = (float) gg.getFontMetrics(font).getStringBounds("X", gg).getHeight();

		// The advance Java uses for a code point no font can display.
		t.missingCodePoint = -1;
		for (int cp = 0xE000; cp <= 0xF8FF; cp++)
			if (!font.canDisplay(cp)) {
				t.missingCodePoint = cp;
				break;
			}
		if (t.missingCodePoint < 0)
			throw new IllegalStateException(family + " " + styleName + " displays the whole private use area");
		t.missingAdvance = width(font, new String(Character.toChars(t.missingCodePoint)));

		Range current = null;
		for (int cp = 0; cp <= Character.MAX_CODE_POINT; cp++) {
			if (cp >= Character.MIN_SURROGATE && cp <= Character.MAX_SURROGATE) {
				current = null;
				continue;
			}
			if (!font.canDisplay(cp)) {
				current = null;
				continue;
			}
			if (current == null) {
				current = new Range(cp);
				t.ranges.add(current);
			}
			current.advances.add(width(font, new String(Character.toChars(cp))));
			t.glyphs++;
		}
		return t;
	}

	// Character pools used to build sample strings.
	static final String[] POOLS = { //
			"abcdefghijklmnopqrstuvwxyz", //
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ", //
			"0123456789", //
			" .,;:!?-_()[]{}<>/\\|\"'#$%&*+=@^`~", //
			"àéîõüçñßÆØ", //
			"абвгдежзикЛМНОП", //
			"αβγδεΩΣπ", //
			"→←↔•…©®™€", //
			"中文日本가한", // usually not displayable
			"😀👍", // emoji as surrogate pairs
	};

	static String randomText(Random rnd) {
		final int length = 1 + rnd.nextInt(40);
		final StringBuilder sb = new StringBuilder();
		while (sb.length() < length) {
			final String pool = POOLS[rnd.nextInt(POOLS.length)];
			final int start = rnd.nextInt(pool.length());
			int end = start + 1;
			if (Character.isHighSurrogate(pool.charAt(start)))
				end = start + 2;
			else if (Character.isLowSurrogate(pool.charAt(start)))
				continue;
			sb.append(pool, start, end);
		}
		return sb.toString();
	}

	static double writeSamples(File f, List<Table> tables) throws IOException {
		final Random rnd = new Random(SAMPLE_SEED);
		final FontRenderContext frc = gg.getFontRenderContext();
		double maxError = 0;
		try (Writer w = new OutputStreamWriter(new FileOutputStream(f), StandardCharsets.UTF_8)) {
			w.write("{\n  \"refSize\": " + num(REF_SIZE) + ",\n  \"samples\": [\n");
			for (int i = 0; i < SAMPLE_COUNT; i++) {
				final Table t = tables.get(rnd.nextInt(tables.size()));
				final int styleIndex = indexOf(STYLE_NAMES, t.style);
				final int size = 6 + rnd.nextInt(35);
				final String text = randomText(rnd);
				final Font font = new Font(t.family, STYLES[styleIndex], size);
				final FontMetrics fm = gg.getFontMetrics(font);
				final double width = fm.getStringBounds(text, gg).getWidth();
				final double height = fm.getStringBounds(text, gg).getHeight();
				final double descent = font.getLineMetrics(text, frc).getDescent();

				double predicted = 0;
				for (int k = 0; k < text.length(); k = text.offsetByCodePoints(k, 1))
					predicted += t.advance(text.codePointAt(k));
				predicted = predicted * size / REF_SIZE;
				maxError = Math.max(maxError, Math.abs(predicted - width));

				w.write("    {\"family\": \"" + t.family + "\", \"style\": \"" + t.style + "\", \"size\": " + size
						+ ", \"text\": \"" + jsonEscape(text) + "\", \"width\": " + num(width) + ", \"height\": "
						+ num(height) + ", \"descent\": " + num(descent) + "}" + (i + 1 < SAMPLE_COUNT ? "," : "")
						+ "\n");
			}
			w.write("  ]\n}\n");
		}
		return maxError;
	}

	static void writeSummary(File f, List<Table> tables, double maxError) throws IOException {
		try (Writer w = new OutputStreamWriter(new FileOutputStream(f), StandardCharsets.UTF_8)) {
			w.write("{\n");
			w.write("  \"javaRuntime\": \"" + jsonEscape(System.getProperty("java.runtime.version")) + "\",\n");
			w.write("  \"refSize\": " + num(REF_SIZE) + ",\n");
			w.write("  \"maxAdditivityError\": " + num(maxError) + ",\n");
			w.write("  \"dialogEqualsSansSerif\": " + sameAdvances(tables, "Dialog", "SansSerif") + ",\n");
			w.write("  \"tables\": [\n");
			for (int i = 0; i < tables.size(); i++) {
				final Table t = tables.get(i);
				w.write("    {\"family\": \"" + t.family + "\", \"style\": \"" + t.style + "\", \"glyphs\": "
						+ t.glyphs + ", \"ranges\": " + t.ranges.size() + ", \"ascent\": " + num(t.ascent)
						+ ", \"descent\": " + num(t.descent) + ", \"leading\": " + num(t.leading)
						+ ", \"boundsHeight\": " + num(t.boundsHeight) + ", \"missingCodePoint\": "
						+ t.missingCodePoint + ", \"missingAdvance\": " + num(t.missingAdvance) + "}"
						+ (i + 1 < tables.size() ? "," : "") + "\n");
			}
			w.write("  ]\n}\n");
		}
	}

	static boolean sameAdvances(List<Table> tables, String a, String b) {
		for (String style : STYLE_NAMES) {
			Table ta = null, tb = null;
			for (Table t : tables) {
				if (t.style.equals(style) && t.family.equals(a))
					ta = t;
				if (t.style.equals(style) && t.family.equals(b))
					tb = t;
			}
			if (ta == null || tb == null || ta.glyphs != tb.glyphs || ta.ranges.size() != tb.ranges.size()
					|| ta.missingAdvance != tb.missingAdvance)
				return false;
			for (int i = 0; i < ta.ranges.size(); i++) {
				final Range ra = ta.ranges.get(i), rb = tb.ranges.get(i);
				if (ra.first != rb.first || !ra.advances.equals(rb.advances))
					return false;
			}
		}
		return true;
	}

	static int indexOf(String[] values, String value) {
		for (int i = 0; i < values.length; i++)
			if (values[i].equals(value))
				return i;
		throw new IllegalArgumentException(value);
	}

	static String num(double v) {
		return Double.toString(v);
	}

	static String jsonEscape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '"' || c == '\\')
				sb.append('\\').append(c);
			else if (c < 0x20 || c > 0x7e)
				sb.append(String.format("\\u%04x", (int) c));
			else
				sb.append(c);
		}
		return sb.toString();
	}
}
