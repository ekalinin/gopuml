// SPDX-License-Identifier: MIT
// Provenance: original
//
// SvgGolden writes the golden files of the internal/klimt/svg and
// internal/export tests from the upstream jar. It interprets the drawing
// scripts of testdata/cases.txt (the language is described in
// internal/klimt/svg/svgtest/script.go, which runs the same scripts in Go)
// and records the SVG text upstream writes.
//
// Build it on the host and run it in the reference image, whose fonts
// (DejaVu) are those of the Go font metrics. From the repository root:
//
//   mkdir -p ~/scratch-svg/classes && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-svg/classes internal/klimt/svg/testdata/javagen/SvgGolden.java
//   docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" -v "$PWD:/work" \
//     -v ~/scratch-svg/classes:/gen plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/gen SvgGolden /work
//
// It writes internal/klimt/svg/testdata/golden/*.svg, format.tsv.gz and
// xml.txt, and internal/export/testdata/golden/*.svg and imagedata.txt.
// Then run it again for the cases of cases-insecure.txt, which need process
// settings (the INSECURE security profile and javascript links):
//
//   docker run ... plantuml-go-reference java -DPLANTUML_SECURITY_PROFILE=INSECURE \
//     -DPLANTUML_ALLOW_JAVASCRIPT_IN_LINK=true (same options) SvgGolden /work insecure
//
// A third run, "SvgGolden /work glyphs", writes golden/glyphs.txt.gz: the
// UnusedSpace centers and a hash of the TextLayout outlines of the printable
// ASCII characters, for each logical font, style and size, which
// centeredchar_test.go compares with the Go tables. The embedded tables
// (glyphs/*.bin) are derived from it and from the DejaVu fonts of the image
// (.cache/fonts/dejavu) with:
//
//   go test ./internal/klimt/svg -run TestGenerateGlyphData -update-glyphs
//
// -XX:-InlineMathNatives makes Math.sin and Math.cos compute as StrictMath,
// which the Go port follows (design D17).

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.SplittableRandom;
import java.util.zip.GZIPOutputStream;

import net.atmp.PixelImage;
import net.atmp.SvgOption;
import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.FileFormatOption;
import net.sourceforge.plantuml.ScaleSimple;
import net.sourceforge.plantuml.ScaleWidth;
import net.sourceforge.plantuml.ScaleMaxWidth;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.core.ImageData;
import net.sourceforge.plantuml.core.TextBlockExporter;
import net.sourceforge.plantuml.klimt.AffineTransformType;
import net.sourceforge.plantuml.klimt.CopyForegroundColorToBackgroundColor;
import net.sourceforge.plantuml.klimt.SvgAttributes;
import net.sourceforge.plantuml.klimt.UClip;
import net.sourceforge.plantuml.klimt.UGroup;
import net.sourceforge.plantuml.klimt.UGroupType;
import net.sourceforge.plantuml.klimt.UPath;
import net.sourceforge.plantuml.klimt.UStroke;
import net.sourceforge.plantuml.klimt.UTranslate;
import net.sourceforge.plantuml.klimt.awt.PortableImage;
import net.sourceforge.plantuml.klimt.awt.PortableImageFactory;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.color.ColorOrder;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorLinearGradient;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColor.TransparentFillBehavior;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.drawing.svg.SvgGraphics;
import net.sourceforge.plantuml.klimt.drawing.svg.UGraphicSvg;
import net.sourceforge.plantuml.klimt.drawing.svg.XmlDocument;
import net.sourceforge.plantuml.klimt.drawing.svg.XmlNode;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontFace;
import net.sourceforge.plantuml.klimt.font.UFontFactory;
import net.sourceforge.plantuml.klimt.font.UFontStyle;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.DotPath;
import net.sourceforge.plantuml.klimt.shape.UCenteredCharacter;
import net.sourceforge.plantuml.klimt.shape.UComment;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UEmpty;
import net.sourceforge.plantuml.klimt.shape.UHidden;
import net.sourceforge.plantuml.klimt.shape.UImage;
import net.sourceforge.plantuml.klimt.shape.UImageSvg;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPixel;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.ConfigurationStore;
import net.sourceforge.plantuml.preproc.OptionKey;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.SkinParam;
import net.sourceforge.plantuml.style.ClockwiseTopRightBottomLeft;
import net.sourceforge.plantuml.style.LengthAdjust;
import net.sourceforge.plantuml.url.Url;

public class SvgGolden {

	public static void main(String[] args) throws Exception {
		final Path root = Paths.get(args[0]);
		final Path svgDir = root.resolve("internal/klimt/svg/testdata");
		if (args.length > 1 && args[1].equals("glyphs")) {
			glyphs(svgDir.resolve("golden/glyphs.txt.gz"));
			return;
		}
		if (args.length > 1 && args[1].equals("insecure")) {
			runCases(svgDir.resolve("cases-insecure.txt"), svgDir.resolve("golden"), false);
			return;
		}
		runCases(svgDir.resolve("cases.txt"), svgDir.resolve("golden"), false);
		xmlWriter(svgDir.resolve("golden/xml.txt"));
		formatTable(svgDir.resolve("golden/format.tsv.gz"));
		final Path exportDir = root.resolve("internal/export/testdata");
		runCases(exportDir.resolve("cases.txt"), exportDir.resolve("golden"), true);
	}

	// ---------------------------------------------------------------------
	// Script reading
	// ---------------------------------------------------------------------

	// Tokens are separated by spaces; \s is a space, \t a tab, \n a newline,
	// \\ a backslash, \_ the empty string and \\uXXXX a UTF-16 unit.
	static String unescape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c != '\\' || i + 1 >= s.length()) {
				sb.append(c);
				continue;
			}
			final char n = s.charAt(++i);
			switch (n) {
			case 's':
				sb.append(' ');
				break;
			case 't':
				sb.append('\t');
				break;
			case 'n':
				sb.append('\n');
				break;
			case '_':
				break;
			case 'u':
				sb.append((char) Integer.parseInt(s.substring(i + 1, i + 5), 16));
				i += 4;
				break;
			default:
				sb.append(n);
			}
		}
		return sb.toString();
	}

	static List<List<String>> readScript(Path file) throws IOException {
		final List<List<String>> result = new ArrayList<>();
		for (String line : Files.readAllLines(file, StandardCharsets.UTF_8)) {
			line = line.trim();
			if (line.isEmpty() || line.startsWith("#"))
				continue;
			final List<String> tokens = new ArrayList<>();
			for (String t : line.split("\\s+"))
				tokens.add(unescape(t));
			result.add(tokens);
		}
		return result;
	}

	static void runCases(Path cases, Path outDir, boolean export) throws Exception {
		Files.createDirectories(outDir);
		final List<List<String>> lines = readScript(cases);
		final StringBuilder imageData = new StringBuilder();
		Case current = null;
		for (List<String> t : lines) {
			if (t.get(0).equals("case")) {
				current = new Case(t.get(1), export);
				continue;
			}
			if (t.get(0).equals("write")) {
				final byte[] out = current.write(t.size() > 1 ? t.get(1) : null);
				Files.write(outDir.resolve(current.name + ".svg"), out);
				if (export)
					imageData.append(current.name + "\t" + current.imageData + "\n");
				current = null;
				continue;
			}
			try {
				current.exec(t);
			} catch (Exception e) {
				throw new IllegalStateException("case " + current.name + ": " + t, e);
			}
		}
		if (export)
			Files.write(outDir.resolve("imagedata.txt"), imageData.toString().getBytes(StandardCharsets.UTF_8));
	}

	// ---------------------------------------------------------------------
	// Values
	// ---------------------------------------------------------------------

	static double d(String s) {
		return Double.parseDouble(s);
	}

	static HColor color(String s) throws Exception {
		if (s.equals("null"))
			return null;
		if (s.startsWith("dark(")) {
			final String[] p = s.substring(5, s.length() - 1).split(",");
			return color(p[0]).withDark(color(p[1]));
		}
		if (s.startsWith("opacity(")) {
			return color(s.substring(8, s.length() - 1))
					.withTransparentFillBehavior(TransparentFillBehavior.WITH_FILL_OPACITY);
		}
		if (s.startsWith("lg(")) {
			// lg(x1,y1,x2,y2,userSpace,SPREAD,offset:color:opacity,...)
			final String[] p = s.substring(3, s.length() - 1).split(",");
			final List<HColorLinearGradient.Stop> stops = new ArrayList<>();
			for (int i = 6; i < p.length; i++) {
				final String[] st = p[i].split(":");
				stops.add(new HColorLinearGradient.Stop(d(st[0]), color(st[1]), d(st[2])));
			}
			return new HColorLinearGradient(d(p[0]), d(p[1]), d(p[2]), d(p[3]), Boolean.parseBoolean(p[4]),
					HColorLinearGradient.SpreadMethod.valueOf(p[5]), stops);
		}
		return HColorSet.instance().getColor(s);
	}

	static ColorMapper mapper(String s) {
		switch (s) {
		case "identity":
			return ColorMapper.IDENTITY;
		case "dark":
			return ColorMapper.DARK_MODE;
		case "mono":
			return ColorMapper.MONOCHROME;
		case "monoreverse":
			return ColorMapper.MONOCHROME_REVERSE;
		case "reverse":
			return ColorMapper.reverse(ColorOrder.RGB);
		case "lightness":
			return ColorMapper.LIGTHNESS_INVERSE;
		}
		throw new IllegalArgumentException(s);
	}

	static UFontFace face(String s) {
		switch (s) {
		case "plain":
			return UFontFace.normal();
		case "bold":
			return UFontFace.bold();
		case "italic":
			return UFontFace.italic();
		case "bolditalic":
			return UFontFace.boldItalic();
		}
		if (s.startsWith("w"))
			return UFontFace.of(UFontStyle.NORMAL, Integer.parseInt(s.substring(1)));
		if (s.startsWith("iw"))
			return UFontFace.of(UFontStyle.ITALIC, Integer.parseInt(s.substring(2)));
		throw new IllegalArgumentException(s);
	}

	// ---------------------------------------------------------------------
	// A test case
	// ---------------------------------------------------------------------

	static class Case {
		final String name;
		final boolean export;
		SvgOption option = SvgOption.basic().withBackcolor(net.sourceforge.plantuml.klimt.color.HColors.WHITE);
		final ConfigurationStore<OptionKey> options = ConfigurationStore.createEmpty();
		boolean useOptions = false;
		long seed = 42;
		FileFormat format = FileFormat.SVG;
		UGraphic root;
		UGraphic ug;
		final Deque<UGraphic> stack = new ArrayDeque<>();
		String imageData;

		// Export mode
		final List<List<String>> drawing = new ArrayList<>();
		double tbWidth = 100, tbHeight = 50;
		HColor tbBack;
		boolean exportBackSet;
		HColor exportBack;
		ClockwiseTopRightBottomLeft margin;
		net.sourceforge.plantuml.Scale scale;
		DiagramType diagramType;
		boolean withSkinParam;
		final List<String[]> skinParams = new ArrayList<>();
		final Pragma pragma = Pragma.createEmpty();
		boolean withPragma;
		FileFormatOption ffo;

		Case(String name, boolean export) {
			this.name = name;
			this.export = export;
		}

		void exec(List<String> t) throws Exception {
			final String cmd = t.get(0);
			if (export && cmd.equals("export")) {
				exportOption(t);
				return;
			}
			if (cmd.equals("opt")) {
				option(t);
				return;
			}
			if (export) {
				drawing.add(t);
				return;
			}
			if (cmd.equals("build")) {
				build();
				return;
			}
			if (root == null)
				build();
			ug = draw(ug, t);
		}

		void build() {
			if (useOptions)
				option = option.withConfigurationStore(options);
			final StringBounder sb = format.getDefaultStringBounder();
			root = UGraphicSvg.build(option, false, seed, sb, format);
			ug = root;
		}

		void option(List<String> t) throws Exception {
			final String k = t.get(1);
			switch (k) {
			case "scale":
				option = option.withScale(d(t.get(2)));
				break;
			case "decimal":
				option = option.withDecimal(Integer.parseInt(t.get(2)));
				break;
			case "lengthadjust":
				option = option.withLengthAdjust(LengthAdjust.valueOf(t.get(2)));
				break;
			case "preserve":
				option = option.withPreserveAspectRatio(t.get(2));
				break;
			case "hover":
				option = option.withHoverPathColorRGB(t.get(2));
				break;
			case "dimstyle":
				option = option.withSvgDimensionStyle(Boolean.parseBoolean(t.get(2)));
				break;
			case "mindim":
				option = option.withMinDim(new XDimension2D(d(t.get(2)), d(t.get(3))));
				break;
			case "back":
				option = option.withBackcolor(color(t.get(2)));
				break;
			case "linktarget":
				option = option.withLinkTarget(t.get(2));
				break;
			case "title":
				option = option.withTitle(net.sourceforge.plantuml.klimt.creole.Display.getWithNewlines(
						net.sourceforge.plantuml.skin.Pragma.createEmpty(), t.get(2)));
				break;
			case "svgtitle":
				options.define(OptionKey.SVG_TITLE, t.get(2));
				useOptions = true;
				break;
			case "svgdesc":
				options.define(OptionKey.SVG_DESC, t.get(2));
				useOptions = true;
				break;
			case "interactive":
				option = option.withInteractive(t.get(2));
				break;
			case "rootattr":
				option = option.withRootAttribute(t.get(2), t.get(3));
				break;
			case "mapper":
				option = option.withColorMapper(mapper(t.get(2)));
				break;
			case "seed":
				seed = Long.parseLong(t.get(2));
				break;
			case "format":
				format = FileFormat.valueOf(t.get(2));
				break;
			case "insecure":
			case "allowjs":
				// Process state upstream: these cases are only run in Go.
				break;
			default:
				throw new IllegalArgumentException(k);
			}
		}

		void exportOption(List<String> t) throws Exception {
			final String k = t.get(1);
			switch (k) {
			case "format":
				format = FileFormat.valueOf(t.get(2));
				break;
			case "dim":
				tbWidth = d(t.get(2));
				tbHeight = d(t.get(3));
				break;
			case "tbback":
				tbBack = color(t.get(2));
				break;
			case "back":
				exportBackSet = true;
				exportBack = color(t.get(2));
				break;
			case "margin":
				margin = ClockwiseTopRightBottomLeft.topRightBottomLeft(d(t.get(2)), d(t.get(3)), d(t.get(4)), d(t.get(5)));
				break;
			case "seed":
				seed = Long.parseLong(t.get(2));
				break;
			case "scale":
				if (t.get(2).equals("simple"))
					scale = new ScaleSimple(d(t.get(3)));
				else if (t.get(2).equals("width"))
					scale = new ScaleWidth(d(t.get(3)));
				else if (t.get(2).equals("maxwidth"))
					scale = new ScaleMaxWidth(d(t.get(3)));
				else
					throw new IllegalArgumentException(t.get(2));
				break;
			case "diagramtype":
				diagramType = DiagramType.valueOf(t.get(2));
				break;
			case "skinparam":
				withSkinParam = true;
				if (t.size() > 2)
					skinParams.add(new String[] { t.get(2), t.get(3) });
				break;
			case "pragma":
				withPragma = true;
				pragma.define(t.get(2), t.get(3));
				break;
			case "svgtitle":
				withSkinParam = true;
				options.define(OptionKey.SVG_TITLE, t.get(2));
				break;
			case "ffo":
				ffo = fileFormatOption(t);
				break;
			default:
				throw new IllegalArgumentException(k);
			}
		}

		FileFormatOption fileFormatOption(List<String> t) {
			FileFormatOption result = ffo == null ? new FileFormatOption(format) : ffo;
			final String k = t.get(2);
			switch (k) {
			case "linktarget":
				return result.withSvgLinkTarget(t.get(3));
			case "hover":
				return result.withHoverColor(t.get(3));
			case "preserve":
				return result.withPreserveAspectRatio(t.get(3));
			case "decimal":
				return result.withDecimal(Integer.parseInt(t.get(3)));
			case "mapper":
				return result.withColorMapper(mapper(t.get(3)));
			}
			throw new IllegalArgumentException(k);
		}

		byte[] write(String metadata) throws Exception {
			final ByteArrayOutputStream os = new ByteArrayOutputStream();
			if (export) {
				exportTo(os, metadata);
				return os.toByteArray();
			}
			if (root == null)
				build();
			root.writeToStream(os, metadata, 96);
			return os.toByteArray();
		}

		void exportTo(OutputStream os, String metadata) throws Exception {
			final List<List<String>> ops = drawing;
			final double w = tbWidth, h = tbHeight;
			final HColor back = tbBack;
			final TextBlock tb = new TextBlock() {
				public void drawU(UGraphic ug) {
					try {
						for (List<String> t : ops)
							ug = draw(ug, t);
					} catch (Exception e) {
						throw new RuntimeException(e);
					}
				}

				public XDimension2D calculateDimension(StringBounder stringBounder) {
					return new XDimension2D(w, h);
				}

				@Override
				public HColor getBackcolor() {
					return back;
				}
			};
			if (ffo == null)
				ffo = new FileFormatOption(format);
			TextBlockExporter.Builder b = TextBlockExporter.builder(tb, ffo, false);
			if (exportBackSet)
				b = b.backcolor(exportBack);
			if (margin != null)
				b = b.margin(margin);
			b = b.metadata(metadata).seed(seed);
			if (scale != null)
				b = b.scale(scale);
			if (diagramType != null)
				b = b.diagramType(diagramType);
			if (withPragma)
				b = b.pragma(pragma);
			if (withSkinParam) {
				final SkinParam sp = SkinParam.create(new HashMap<String, String>(), PathSystem.fetch(),
						diagramType == null ? DiagramType.SEQUENCE : diagramType, pragma, options);
				for (String[] kv : skinParams)
					sp.setParam(kv[0], kv[1]);
				b = b.skinParam(sp);
			}
			final ImageData data = b.build().exportTo(os);
			imageData = data.getWidth() + "\t" + data.getHeight() + "\t" + data.getStatus();
		}

		UGraphic draw(UGraphic ug, List<String> t) throws Exception {
			final String cmd = t.get(0);
			switch (cmd) {
			case "reset":
				return root;
			case "push":
				stack.push(ug);
				return ug;
			case "pop":
				return stack.pop();
			case "color":
				return ug.apply(color(t.get(1)));
			case "bg":
				return ug.apply(color(t.get(1)).bg());
			case "stroke":
				if (t.size() > 2)
					return ug.apply(new UStroke(d(t.get(2)), d(t.get(3)), d(t.get(1))));
				return ug.apply(UStroke.withThickness(d(t.get(1))));
			case "translate":
				return ug.apply(new UTranslate(d(t.get(1)), d(t.get(2))));
			case "clip":
				return ug.apply(new UClip(d(t.get(1)), d(t.get(2)), d(t.get(3)), d(t.get(4))));
			case "hidden":
				return ug.apply(UHidden.HIDDEN);
			case "shown":
				return ug.apply(UHidden.NORMAL_SHOWN);
			case "fg2bg":
				return ug.apply(new CopyForegroundColorToBackgroundColor());
			case "rect": {
				URectangle r = URectangle.build(d(t.get(1)), d(t.get(2)));
				if (t.size() > 3)
					r = r.rounded(d(t.get(3)));
				if (t.size() > 4)
					r.setDeltaShadow(d(t.get(4)));
				ug.draw(r);
				return ug;
			}
			case "line": {
				final ULine l = new ULine(d(t.get(1)), d(t.get(2)));
				if (t.size() > 3)
					l.setDeltaShadow(d(t.get(3)));
				ug.draw(l);
				return ug;
			}
			case "ellipse": {
				final UEllipse e = t.size() > 3 ? new UEllipse(d(t.get(1)), d(t.get(2)), d(t.get(3)), d(t.get(4)))
						: UEllipse.build(d(t.get(1)), d(t.get(2)));
				if (t.size() > 5)
					e.setDeltaShadow(d(t.get(5)));
				ug.draw(e);
				return ug;
			}
			case "polygon": {
				final UPolygon p = new UPolygon();
				for (int i = 2; i + 1 < t.size(); i += 2)
					p.addPoint(d(t.get(i)), d(t.get(i + 1)));
				p.setDeltaShadow(d(t.get(1)));
				ug.draw(p);
				return ug;
			}
			case "path": {
				final UPath p = new UPath(nul(t.get(2)), nul(t.get(3)));
				p.setDeltaShadow(d(t.get(1)));
				for (int i = 4; i < t.size();) {
					final String s = t.get(i++);
					switch (s) {
					case "M":
						p.moveTo(d(t.get(i)), d(t.get(i + 1)));
						i += 2;
						break;
					case "L":
						p.lineTo(d(t.get(i)), d(t.get(i + 1)));
						i += 2;
						break;
					case "Q":
						p.quadTo(d(t.get(i)), d(t.get(i + 1)), d(t.get(i + 2)), d(t.get(i + 3)));
						i += 4;
						break;
					case "C":
						p.cubicTo(d(t.get(i)), d(t.get(i + 1)), d(t.get(i + 2)), d(t.get(i + 3)), d(t.get(i + 4)),
								d(t.get(i + 5)));
						i += 6;
						break;
					case "A":
						p.arcTo(d(t.get(i)), d(t.get(i + 1)), d(t.get(i + 2)), d(t.get(i + 3)), d(t.get(i + 4)),
								d(t.get(i + 5)), d(t.get(i + 6)));
						i += 7;
						break;
					case "Z":
						p.closePath();
						break;
					default:
						throw new IllegalArgumentException(s);
					}
				}
				ug.draw(p);
				return ug;
			}
			case "dotpath": {
				DotPath p = new DotPath();
				p = p.addCurve(pt(t, 1), pt(t, 3), pt(t, 5), pt(t, 7));
				for (int i = 9; i + 5 < t.size(); i += 6)
					p = p.addCurve(pt(t, i), pt(t, i + 2), pt(t, i + 4));
				ug.draw(p);
				return ug;
			}
			case "pixel":
				ug.draw(new UPixel());
				return ug;
			case "text": {
				// text FAMILY SIZE FACE STYLES COLOR EXTCOLOR ATTRS ORIENTATION TEXT
				final UFont font = UFontFactory.build(t.get(1), face(t.get(3)), Integer.parseInt(t.get(2)));
				UStroke underline = UStroke.simple();
				FontConfiguration fc = null;
				final String styles = t.get(4);
				if (styles.contains("nounderlinestroke"))
					underline = null;
				else if (styles.contains("thinunderlinestroke"))
					underline = UStroke.withThickness(0);
				fc = FontConfiguration.create(font, color(t.get(5)), color("#0000FF"), underline);
				if (!styles.equals("-"))
					for (String st : styles.split(","))
						if (!st.endsWith("underlinestroke"))
							fc = fc.add(FontStyle.valueOf(st.toUpperCase()));
				if (!t.get(6).equals("-"))
					fc = fc.changeExtendedColor(color(t.get(6)));
				if (!t.get(7).equals("-"))
					fc = fc.changeAttributes(SvgAttributes.build(t.get(7)));
				UText text = UText.build(t.get(9), fc);
				final int orientation = Integer.parseInt(t.get(8));
				if (orientation != 0)
					text = text.withOrientation(orientation);
				ug.draw(text);
				return ug;
			}
			case "image": {
				final int w = Integer.parseInt(t.get(1));
				final int h = Integer.parseInt(t.get(2));
				final PortableImage im = PortableImageFactory.build(w, h, PortableImage.TYPE_INT_ARGB);
				for (int x = 0; x < w; x++)
					for (int y = 0; y < h; y++)
						im.setRGB(x, y, pixel(x, y));
				ug.draw(new UImage(new PixelImage(im, AffineTransformType.TYPE_BILINEAR)));
				return ug;
			}
			case "imagesvg":
				ug.draw(new UImageSvg(t.get(2), d(t.get(1))));
				return ug;
			case "comment":
				ug.draw(new UComment(t.get(1)));
				return ug;
			case "centered":
				ug.draw(new UCenteredCharacter(t.get(1).charAt(0),
						UFontFactory.build(t.size() > 3 ? t.get(3) : "Monospaced",
								t.size() > 4 ? face(t.get(4)) : UFontFace.bold(), Integer.parseInt(t.get(2)))));
				return ug;
			case "empty":
				ug.draw(new UEmpty(d(t.get(1)), d(t.get(2))));
				return ug;
			case "group": {
				final UGroup g = new UGroup();
				for (int i = 1; i < t.size(); i++) {
					final int eq = t.get(i).indexOf('=');
					g.put(UGroupType.valueOf(t.get(i).substring(0, eq)), t.get(i).substring(eq + 1));
				}
				ug.startGroup(g);
				return ug;
			}
			case "endgroup":
				ug.closeGroup();
				return ug;
			case "url":
				ug.startUrl(new Url(t.get(1), t.size() > 2 ? t.get(2) : null, null));
				return ug;
			case "endurl":
				ug.closeUrl();
				return ug;
			}
			throw new IllegalArgumentException(cmd);
		}
	}

	static String nul(String s) {
		return s.equals("-") ? null : s;
	}

	static XPoint2D pt(List<String> t, int i) {
		return new XPoint2D(d(t.get(i)), d(t.get(i + 1)));
	}

	// The pixels of test images: an ARGB pattern with some transparency.
	static int pixel(int x, int y) {
		final int a = (x + y) % 3 == 0 ? 0x80 : 0xFF;
		return a << 24 | ((x * 37) & 0xFF) << 16 | ((y * 59) & 0xFF) << 8 | ((x * y * 13) & 0xFF);
	}

	// ---------------------------------------------------------------------
	// Glyphs of centered characters: UnusedSpace and outlines
	// ---------------------------------------------------------------------

	static final String[] STYLES = { "plain", "bold", "italic", "bolditalic" };

	static void glyphs(Path out) throws Exception {
		try (PrintStream ps = new PrintStream(new GZIPOutputStream(Files.newOutputStream(out)), false, "UTF-8")) {
			for (String family : new String[] { "SansSerif", "Serif", "Monospaced" })
				for (String style : STYLES)
					for (int size = 4; size <= 32; size++)
						glyphLine(ps, family, style, size);
			for (String family : new String[] { "Dialog", "DialogInput", "DejaVu Sans", "DejaVu Serif",
					"DejaVu Sans Mono" })
				for (String style : STYLES)
					for (int size : new int[] { 9, 12, 14, 17, 23 })
						glyphLine(ps, family, style, size);
		}
	}

	static void glyphLine(PrintStream ps, String family, String style, int size) {
		final UFont font = UFontFactory.build(family, face(style), size);
		final StringBuilder centers = new StringBuilder();
		final StringBuilder outlines = new StringBuilder();
		for (char c = 0x20; c <= 0x7E; c++) {
			final net.sourceforge.plantuml.klimt.font.UnusedSpace u = net.sourceforge.plantuml.klimt.font.UnusedSpace
					.getUnusedSpace(font, c);
			centers.append(" " + (int) (u.getCenterX() * 4) + "," + (int) (u.getCenterY() * 4));
			outlines.append(outline(font, c)).append('|');
		}
		ps.print(family + "\t" + style + "\t" + size + "\t" + fnv(outlines.toString()) + "\t"
				+ centers.toString().trim() + "\n");
	}

	// The outline as upstream's DriverCenteredCharacterSvg reads it: segment
	// types and coordinates in 1/64 pixel.
	static String outline(UFont font, char c) {
		final java.awt.font.TextLayout t = font.createTextLayout("" + c);
		final java.awt.geom.PathIterator it = t.getOutline(null).getPathIterator(null);
		final double[] co = new double[6];
		final StringBuilder sb = new StringBuilder();
		while (!it.isDone()) {
			final int type = it.currentSegment(co);
			if (sb.length() > 0)
				sb.append(';');
			sb.append(type);
			final int n = type == 0 || type == 1 ? 2 : type == 2 ? 4 : type == 3 ? 6 : 0;
			for (int i = 0; i < n; i++) {
				final double v = co[i] * 64;
				if (v != Math.rint(v))
					throw new IllegalStateException("not 26.6: " + co[i]);
				sb.append(' ').append((long) v);
			}
			it.next();
		}
		return sb.toString();
	}

	// FNV-1a 64 of the UTF-8 bytes.
	static String fnv(String s) {
		long h = 0xcbf29ce484222325L;
		for (byte b : s.getBytes(StandardCharsets.UTF_8)) {
			h ^= b & 0xff;
			h *= 0x100000001b3L;
		}
		return Long.toHexString(h);
	}

	// ---------------------------------------------------------------------
	// XML writer, with and without indentation
	// ---------------------------------------------------------------------

	static void xmlWriter(Path out) throws IOException {
		final StringBuilder sb = new StringBuilder();
		for (int indent : new int[] { 0, 2, 3 }) {
			final XmlDocument doc = new XmlDocument();
			final XmlNode root = doc.createElement("root");
			doc.setRoot(root);
			root.setAttribute("a", "x&y<z>\"q'");
			root.setAttribute("b", null);
			root.setAttribute("a", "again");
			root.appendProcessingInstruction("pi", "data");
			root.appendProcessingInstruction("empty", "");
			root.appendProcessingInstruction("nul", null);
			final XmlNode empty = doc.createElement("empty");
			root.appendChild(empty);
			final XmlNode text = doc.createElement("text");
			text.setTextContent("a & b < c > d \" e");
			root.appendChild(text);
			final XmlNode nulText = doc.createElement("nul");
			nulText.setTextContent(null);
			root.appendChild(nulText);
			final XmlNode mixed = doc.createElement("mixed");
			mixed.appendText("before");
			final XmlNode child = doc.createElement("child");
			child.appendComment("c--d-");
			child.appendComment(null);
			child.appendCData("x]]>y");
			child.appendCData(null);
			mixed.appendChild(child);
			mixed.appendText("after");
			mixed.appendRaw("<raw attr=\"1\"/>");
			mixed.appendRaw(null);
			root.appendChild(mixed);
			final XmlNode deep = doc.createElement("l1");
			final XmlNode l2 = doc.createElement("l2");
			final XmlNode l3 = doc.createElement("l3");
			l3.setAttribute("k", "v");
			l2.appendChild(l3);
			l2.appendText("t");
			deep.appendChild(l2);
			root.appendChild(deep);
			sb.append("indent " + indent + "\n");
			sb.append(doc.toXml(indent));
			sb.append("\n-----\n");
		}
		Files.write(out, sb.toString().getBytes(StandardCharsets.UTF_8));
	}

	// ---------------------------------------------------------------------
	// Number formatting table
	// ---------------------------------------------------------------------

	static void formatTable(Path out) throws Exception {
		final SplittableRandom rnd = new SplittableRandom(20261016L);
		final List<Double> specials = new ArrayList<>();
		final double[] sp = { 0.0, -0.0, Double.NaN, Double.POSITIVE_INFINITY, Double.NEGATIVE_INFINITY,
				Double.MIN_VALUE, -Double.MIN_VALUE, Double.MAX_VALUE, -Double.MAX_VALUE, Double.MIN_NORMAL, 1e-300,
				-1e-300, 0.5, -0.5, 1, -1, 0.0005, -0.0005, 0.0015, 1.0005, 2.675, 1.005, 9.9995, 99.9995, 999.9995,
				0.9995, -0.9995, 0.99995, 0.999995, 123456789.0005, 1e15, 1e16, 1e17, 9007199254740992.0,
				9007199254740994.0, 1152921504606846976.0, 9.223372036854776E18, 1e21, 1e22, 1e23, 1.5e300, -1e21,
				4.35, 0.1 + 0.2, 1.0 / 3, 2.0 / 3, -2.0 / 3, 1e-4, 4.9999e-4, 5.0001e-4, -4.9999e-4, -5.0001e-4,
				2147483647.5, 2147483648.0, -2147483648.5 };
		for (double v : sp)
			specials.add(v);
		final List<Double> bound3 = new ArrayList<>();
		for (int k = -1000; k <= 1000; k++) {
			final double v = k / 2000.0;
			bound3.add(v);
			bound3.add(k * 0.0005);
			bound3.add(Math.nextUp(v));
			bound3.add(Math.nextDown(v));
		}
		for (int k = -300; k <= 300; k++) {
			final double v = 100 + k / 2000.0;
			bound3.add(v);
			bound3.add(Math.nextUp(v));
			bound3.add(Math.nextDown(v));
		}
		final List<Double> uniform = new ArrayList<>();
		for (int i = 0; i < 2000; i++)
			uniform.add(rnd.nextDouble(-2000, 2000));
		for (int i = 0; i < 800; i++)
			uniform.add((rnd.nextBoolean() ? 1 : -1) * Math.pow(10, rnd.nextDouble(-8, 22)));
		for (int i = 0; i < 300; i++)
			uniform.add(Double.longBitsToDouble(rnd.nextLong()));
		for (int n = -2000; n <= 2000; n += 7)
			uniform.add(n / 64.0);
		try (PrintStream ps = new PrintStream(new GZIPOutputStream(Files.newOutputStream(out)), false, "UTF-8")) {
			for (double v : specials)
				line(ps, 3, 1.0, v);
			for (double v : bound3)
				line(ps, 3, 1.0, v);
			for (double v : uniform)
				line(ps, 3, 1.0, v);
			for (int dec : new int[] { 0, 1, 2, 4, 5, 6 }) {
				final double div = 2 * Math.pow(10, dec);
				for (int k = -100; k <= 100; k++) {
					final double v = k / div;
					line(ps, dec, 1.0, v);
					line(ps, dec, 1.0, Math.nextUp(v));
					line(ps, dec, 1.0, Math.nextDown(v));
				}
				for (double v : specials)
					line(ps, dec, 1.0, v);
				for (int i = 0; i < 200; i++)
					line(ps, dec, 1.0, uniform.get(i));
			}
			for (double scale : new double[] { 1.5, 0.75, 2.0, 96 / 72.0, 1 / 3.0, 0.1, 4.0 }) {
				for (double v : specials)
					line(ps, 3, scale, v);
				for (int i = 0; i < 600; i++)
					line(ps, 3, scale, uniform.get(i));
			}
		}
	}

	static final HashMap<String, Object[]> graphics = new HashMap<>();

	static void line(PrintStream ps, int decimal, double scale, double v) throws Exception {
		final String key = decimal + "/" + scale;
		Object[] g = graphics.get(key);
		if (g == null) {
			final SvgGraphics svg = new SvgGraphics(42, SvgOption.basic().withDecimal(decimal).withScale(scale));
			final Method format = SvgGraphics.class.getDeclaredMethod("format", double.class);
			final Method percent = SvgGraphics.class.getDeclaredMethod("formatPercent", double.class);
			final Method opacity = SvgGraphics.class.getDeclaredMethod("formatOpacity", double.class);
			final Method bool = SvgGraphics.class.getDeclaredMethod("formatBoolean", double.class);
			format.setAccessible(true);
			percent.setAccessible(true);
			opacity.setAccessible(true);
			bool.setAccessible(true);
			g = new Object[] { svg, format, percent, opacity, bool };
			graphics.put(key, g);
		}
		final Object svg = g[0];
		ps.print(decimal + "\t" + Long.toHexString(Double.doubleToRawLongBits(scale)) + "\t"
				+ Long.toHexString(Double.doubleToRawLongBits(v)) + "\t" + ((Method) g[1]).invoke(svg, v) + "\t"
				+ ((Method) g[2]).invoke(svg, v) + "\t" + ((Method) g[3]).invoke(svg, v) + "\t"
				+ ((Method) g[4]).invoke(svg, v) + "\n");
	}
}
