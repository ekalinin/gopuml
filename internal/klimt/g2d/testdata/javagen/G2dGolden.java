// SPDX-License-Identifier: MIT
// Provenance: original
//
// G2dGolden writes the golden PNG files of the internal/klimt/g2d and
// internal/export raster tests from the upstream jar. It interprets the
// drawing scripts of internal/klimt/g2d/testdata/cases.txt (the language of
// internal/klimt/svg/svgtest/script.go, plus the settings listed in that
// file) with upstream's UGraphicG2d, and the export scripts of
// internal/export/testdata/png-cases.txt with upstream's TextBlockExporter.
//
// Build it on the host and run it in the reference image, whose fonts
// (DejaVu) are those embedded by the Go driver. From the repository root:
//
//   mkdir -p ~/scratch-g2d/classes && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-g2d/classes internal/klimt/g2d/testdata/javagen/G2dGolden.java
//   docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" -v "$PWD:/work" \
//     -v ~/scratch-g2d/classes:/gen plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/gen G2dGolden /work
//
// It writes internal/klimt/g2d/testdata/golden/*.png and
// internal/export/testdata/golden-png/*.png and imagedata.txt.

import java.awt.Graphics2D;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;

import javax.imageio.ImageIO;

import net.atmp.PixelImage;
import net.sourceforge.plantuml.EmptyImageBuilder;
import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.FileFormatOption;
import net.sourceforge.plantuml.ScaleMaxWidth;
import net.sourceforge.plantuml.ScaleSimple;
import net.sourceforge.plantuml.ScaleWidth;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.core.ImageData;
import net.sourceforge.plantuml.core.TextBlockExporter;
import net.sourceforge.plantuml.klimt.AffineTransformType;
import net.sourceforge.plantuml.klimt.CopyForegroundColorToBackgroundColor;
import net.sourceforge.plantuml.klimt.SvgAttributes;
import net.sourceforge.plantuml.klimt.UAntiAliasing;
import net.sourceforge.plantuml.klimt.UClip;
import net.sourceforge.plantuml.klimt.UGroup;
import net.sourceforge.plantuml.klimt.UGroupType;
import net.sourceforge.plantuml.klimt.UPath;
import net.sourceforge.plantuml.klimt.UPattern;
import net.sourceforge.plantuml.klimt.UStroke;
import net.sourceforge.plantuml.klimt.UTranslate;
import net.sourceforge.plantuml.klimt.awt.PortableImage;
import net.sourceforge.plantuml.klimt.awt.PortableImageFactory;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorLinearGradient;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColors;
import net.sourceforge.plantuml.klimt.color.HColor.TransparentFillBehavior;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.drawing.g2d.UGraphicG2d;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontFace;
import net.sourceforge.plantuml.klimt.font.UFontFactory;
import net.sourceforge.plantuml.klimt.font.UFontStyle;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.shape.DotPath;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
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
import net.sourceforge.plantuml.url.Url;

public class G2dGolden {

	public static void main(String[] args) throws Exception {
		final Path root = Paths.get(args[0]);
		final Path g2dDir = root.resolve("internal/klimt/g2d/testdata");
		runCases(g2dDir.resolve("cases.txt"), g2dDir.resolve("golden"), false);
		final Path exportDir = root.resolve("internal/export/testdata");
		runCases(exportDir.resolve("png-cases.txt"), exportDir.resolve("golden-png"), true);
	}

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
				Files.write(outDir.resolve(current.name + ".png"), out);
				if (export) {
					final java.awt.image.BufferedImage im = out.length == 0 ? null
							: ImageIO.read(new java.io.ByteArrayInputStream(out));
					final String size = im == null ? "0\t0" : im.getWidth() + "\t" + im.getHeight();
					imageData.append(current.name + "\t" + current.imageData + "\t" + size + "\n");
				}
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

	static class Case {
		final String name;
		final boolean export;
		int width = 200, height = 100;
		HColor back = HColors.WHITE;
		double dpi = 1;
		FileFormat format = FileFormat.PNG;
		UGraphicG2d root;
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
		final ConfigurationStore<OptionKey> options = ConfigurationStore.createEmpty();
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
			if (root == null)
				build();
			ug = draw(ug, t);
		}

		void option(List<String> t) throws Exception {
			final String k = t.get(1);
			switch (k) {
			case "size":
				width = Integer.parseInt(t.get(2));
				height = Integer.parseInt(t.get(3));
				break;
			case "back":
				back = color(t.get(2));
				break;
			case "dpi":
				dpi = d(t.get(2));
				break;
			case "format":
				format = FileFormat.valueOf(t.get(2));
				break;
			default:
				throw new IllegalArgumentException(k);
			}
		}

		void build() {
			final StringBounder sb = FileFormat.PNG.getDefaultStringBounder();
			final EmptyImageBuilder builder = new EmptyImageBuilder(null, width, height,
					back.toColor(ColorMapper.IDENTITY), sb);
			final Graphics2D g2d = builder.getGraphics2D();
			root = new UGraphicG2d(HColors.WHITE, ColorMapper.IDENTITY, sb, g2d, dpi, format);
			root.setPortableImage(builder.getPortableImage());
			ug = root;
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
				margin = ClockwiseTopRightBottomLeft.topRightBottomLeft(d(t.get(2)), d(t.get(3)), d(t.get(4)),
						d(t.get(5)));
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
			case "mapper":
				return result.withColorMapper(mapper(t.get(3)));
			case "watermark":
				return result.withWartermark(t.get(3));
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
			final HColor tbb = tbBack;
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
					return tbb;
				}
			};
			if (ffo == null)
				ffo = new FileFormatOption(format);
			TextBlockExporter.Builder b = TextBlockExporter.builder(tb, ffo, false);
			if (exportBackSet)
				b = b.backcolor(exportBack);
			if (margin != null)
				b = b.margin(margin);
			b = b.metadata(metadata);
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
			case "aa":
				return ug.apply(t.get(1).equals("off") ? UAntiAliasing.ANTI_ALIASING_OFF
						: UAntiAliasing.ANTI_ALIASING_ON);
			case "pattern":
				return ug.apply(UPattern.valueOf(t.get(1)));
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
				final String styles = t.get(4);
				if (styles.contains("nounderlinestroke"))
					underline = null;
				else if (styles.contains("thinunderlinestroke"))
					underline = UStroke.withThickness(0);
				FontConfiguration fc = FontConfiguration.create(font, color(t.get(5)), color("#0000FF"), underline);
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
}
