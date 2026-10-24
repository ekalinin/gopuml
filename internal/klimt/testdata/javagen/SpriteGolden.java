// SPDX-License-Identifier: MIT
// Provenance: original
//
// SpriteGolden dumps the golden values of the SVG sprite parser tests of
// internal/klimt from the upstream jar: the SVG path parser of OpenIconic
// (openiconic_svgpath_test.go), the emoji helpers ColorResolver and
// UGraphicWithScale (emoji_test.go), the parser factory and the drawings of
// SvgNanoParser, SvgSaxParser and SpriteSvgNanoParser
// (svgparser_golden_test.go), and the drawings of the sprite classes
// SpriteMonochrome, SpriteColor, SpriteImage, SpriteSvg and StdlibSprite
// (sprite_draw_test.go). From the repository root:
//
//   mkdir -p ~/scratch-sprite && javac -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-sprite internal/klimt/testdata/javagen/SpriteGolden.java
//   java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp .cache/upstream/plantuml-mit-1.2026.9.jar:$HOME/scratch-sprite \
//     SpriteGolden internal/klimt/testdata/golden
//
// -XX:-InlineMathNatives makes Math.sin, Math.cos and the other elementary
// functions compute as StrictMath (fdlibm), which the Go port follows
// (design D17).
//
// It reads svgpath_inputs.txt and svgsprite_inputs.txt of the golden
// directory and writes svgpath.golden, emoji.golden, svgsprite.golden and
// spritedraw.golden next to them. Each drawing is recorded with a UGraphic that logs the
// shapes, their position, colors and stroke, as KlimtGolden2 does.

import java.io.IOException;
import java.io.OutputStream;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.emoji.ColorResolver;
import net.sourceforge.plantuml.emoji.GrayLevelRange;
import net.sourceforge.plantuml.emoji.UGraphicWithScale;
import net.sourceforge.plantuml.klimt.UBackground;
import net.sourceforge.plantuml.klimt.UChange;
import net.sourceforge.plantuml.klimt.UGroup;
import net.sourceforge.plantuml.klimt.UParam;
import net.sourceforge.plantuml.klimt.UPath;
import net.sourceforge.plantuml.klimt.UPattern;
import net.sourceforge.plantuml.klimt.UShape;
import net.sourceforge.plantuml.klimt.UStroke;
import net.sourceforge.plantuml.klimt.UTranslate;
import net.sourceforge.plantuml.klimt.awt.PortableImage;
import net.sourceforge.plantuml.klimt.awt.XAffineTransform;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorLinearGradient;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColors;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontContext;
import net.sourceforge.plantuml.klimt.geom.USegment;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UImage;
import net.sourceforge.plantuml.klimt.shape.UImageSvg;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.klimt.sprite.Sprite;
import net.sourceforge.plantuml.klimt.sprite.SpriteColor;
import net.sourceforge.plantuml.klimt.sprite.SpriteImage;
import net.sourceforge.plantuml.klimt.sprite.SpriteMonochrome;
import net.sourceforge.plantuml.klimt.sprite.SpriteSvg;
import net.sourceforge.plantuml.preproc.StdlibSprite;
import net.sourceforge.plantuml.openiconic.Movement;
import net.sourceforge.plantuml.openiconic.SvgCommand;
import net.sourceforge.plantuml.openiconic.SvgPath;
import net.sourceforge.plantuml.security.SImageIO;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.svg.parser.ISvgSpriteParser;
import net.sourceforge.plantuml.svg.parser.SpriteSvgNanoParser;
import net.sourceforge.plantuml.svg.parser.SvgNanoParser;
import net.sourceforge.plantuml.svg.parser.SvgSaxParser;
import net.sourceforge.plantuml.svg.parser.SvgSpriteParserFactory;
import net.sourceforge.plantuml.utils.Base64Coder;

public class SpriteGolden {

	static Path dir;

	// The 2x2 PNGs of svgsprite_inputs.txt: four colors (one half
	// transparent), and one color.
	static final String FOUR = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAFElEQVR42mP4z8DwHwyBNBAwNAAAR0sIecMlh+sAAAAASUVORK5CYII=";
	static final String UNIFORM = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAYAAABytg0kAAAAEUlEQVR42mMQUDD4D8IMMAYAK7QFfSBd8FwAAAAASUVORK5CYII=";

	public static void main(String[] args) throws Exception {
		dir = Paths.get(args[0]);
		svgPaths();
		emoji();
		sprites();
		spriteDrawings();
	}

	static void write(String name, List<String> lines) throws IOException {
		Files.write(dir.resolve(name), (String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
	}

	static List<String> inputs(String name) throws IOException {
		final List<String> result = new ArrayList<>();
		for (String line : Files.readAllLines(dir.resolve(name), StandardCharsets.UTF_8))
			if (line.isEmpty() == false && line.startsWith("#") == false)
				result.add(line);
		return result;
	}

	// \n, \t and \\ of the input files.
	static String unesc(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '\\' && i + 1 < s.length()) {
				final char n = s.charAt(++i);
				sb.append(n == 'n' ? '\n' : n == 't' ? '\t' : n);
			} else
				sb.append(c);
		}
		return sb.toString();
	}

	static String esc(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '\t')
				sb.append("\\t");
			else if (c == '\n')
				sb.append("\\n");
			else if (c == '\\')
				sb.append("\\\\");
			else if (c < 32 || (c >= 0xE000 && c <= 0xF8FF))
				sb.append(String.format("\\u%04X", (int) c));
			else
				sb.append(c);
		}
		return sb.toString();
	}

	// An exception: its simple class name, and its message except for the
	// classes whose messages the Go port does not reproduce (the helpful
	// NullPointerException messages, the module text of ClassCastException).
	static String err(Throwable e) {
		final String name = e.getClass().getSimpleName();
		if (e instanceof NullPointerException || e instanceof ClassCastException || e.getMessage() == null)
			return "ERR " + name;
		return "ERR " + name + ": " + esc(e.getMessage());
	}

	static String col(HColor c) {
		if (c == null)
			return "null";
		if (c instanceof HColorLinearGradient) {
			final HColorLinearGradient g = (HColorLinearGradient) c;
			final List<String> stops = new ArrayList<>();
			for (HColorLinearGradient.Stop s : g.getStops())
				stops.add(s.getOffset() + ":" + col(s.getColor()) + ":" + s.getOpacity());
			return "lg(" + g.getX1() + "," + g.getY1() + "," + g.getX2() + "," + g.getY2() + ","
					+ g.isUserSpaceOnUse() + "," + g.getSpreadMethod() + "," + stops + ")";
		}
		return c.asString();
	}

	static final StringBounder BOUNDER = new StringBounder() {
		@Override
		public XDimension2D calculateDimension(UFont font, String text) {
			return new XDimension2D(text.length() * font.getSize2D() * 0.6, font.getSize2D() * 1.25);
		}

		@Override
		public FileFormat getFileFormat() {
			return FileFormat.PNG;
		}
	};

	static String uf(UFont f) {
		return f.getFamily(null, UFontContext.SVG) + "/" + f.getFontFace() + "/" + f.getSize();
	}

	static String styles(FontConfiguration fc) {
		final List<String> l = new ArrayList<>();
		for (FontStyle s : FontStyle.values())
			if (fc.containsStyle(s))
				l.add(s.name());
		return l.toString();
	}

	// The pixels of a small image, a hash of the others.
	static String pix(PortableImage img) {
		final StringBuilder sb = new StringBuilder(img.getWidth() + "x" + img.getHeight() + ":");
		if (img.getWidth() * img.getHeight() <= 16) {
			final List<String> l = new ArrayList<>();
			for (int y = 0; y < img.getHeight(); y++)
				for (int x = 0; x < img.getWidth(); x++)
					l.add(Integer.toHexString(img.getRGB(x, y)));
			return sb + String.join(",", l);
		}
		int h = 0;
		for (int y = 0; y < img.getHeight(); y++)
			for (int x = 0; x < img.getWidth(); x++)
				h = 31 * h + img.getRGB(x, y);
		return sb + "#" + Integer.toHexString(h);
	}

	static String path(UPath p) {
		final List<String> segs = new ArrayList<>();
		for (USegment s : p)
			segs.add(s.toString());
		return segs.toString();
	}

	static String describe(Object shape) {
		if (shape instanceof UPath) {
			final UPath p = (UPath) shape;
			return "path " + p.getComment() + " " + path(p);
		}
		if (shape instanceof UEllipse) {
			final UEllipse e = (UEllipse) shape;
			return "ellipse " + e.getWidth() + "," + e.getHeight() + " " + e.getStart() + "," + e.getExtend();
		}
		if (shape instanceof URectangle) {
			final URectangle r = (URectangle) shape;
			return "rect " + r.getWidth() + "," + r.getHeight() + " r=" + r.getRx() + "," + r.getRy();
		}
		if (shape instanceof UText) {
			final UText t = (UText) shape;
			final FontConfiguration fc = t.getFontConfiguration();
			return "text '" + esc(t.getText()) + "' " + uf(fc.getFont()) + " " + col(fc.getColor()) + " "
					+ styles(fc);
		}
		if (shape instanceof UImage) {
			final UImage i = (UImage) shape;
			return "image " + i.getWidth() + "," + i.getHeight() + " " + pix(i.getImage(1));
		}
		if (shape instanceof UImageSvg) {
			final UImageSvg i = (UImageSvg) shape;
			return "imagesvg " + i.getWidth() + "," + i.getHeight() + " " + esc(i.getSvg(true));
		}
		return shape.getClass().getSimpleName();
	}

	// A UGraphic that records the drawing operations; svg tells whether it
	// matches the "SVG" property (SVG output).
	static class Rec implements UGraphic {
		final List<String> log;
		final UTranslate t;
		final HColor color;
		final HColor back;
		final UStroke stroke;
		final boolean svg;
		final ColorMapper mapper;

		Rec(List<String> log, UTranslate t, HColor color, HColor back, UStroke stroke, boolean svg,
				ColorMapper mapper) {
			this.log = log;
			this.t = t;
			this.color = color;
			this.back = back;
			this.stroke = stroke;
			this.svg = svg;
			this.mapper = mapper;
		}

		static Rec create(List<String> log, boolean svg, ColorMapper mapper) {
			return new Rec(log, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple(), svg, mapper);
		}

		@Override
		public StringBounder getStringBounder() {
			return BOUNDER;
		}

		@Override
		public UParam getParam() {
			return new UParam() {
				public net.sourceforge.plantuml.klimt.UClip getClip() {
					return null;
				}

				public HColor getColor() {
					return color;
				}

				public HColor getBackcolor() {
					return back;
				}

				public UStroke getStroke() {
					return stroke;
				}

				public boolean isHidden() {
					return false;
				}

				public UPattern getPattern() {
					return UPattern.FULL;
				}
			};
		}

		@Override
		public <SHAPE extends UShape> void draw(SHAPE shape) {
			log.add("@" + t.getDx() + "," + t.getDy() + " c=" + col(color) + " b=" + col(back) + " s=" + stroke + " "
					+ describe(shape));
		}

		@Override
		public UGraphic apply(UChange change) {
			if (change instanceof UTranslate)
				return new Rec(log, ((UTranslate) change).compose(t), color, back, stroke, svg, mapper);
			if (change instanceof HColor)
				return new Rec(log, t, (HColor) change, back, stroke, svg, mapper);
			if (change instanceof UBackground)
				return new Rec(log, t, color, ((UBackground) change).getBackColor(), stroke, svg, mapper);
			if (change instanceof UStroke)
				return new Rec(log, t, color, back, (UStroke) change, svg, mapper);
			log.add("apply " + change.getClass().getSimpleName());
			return this;
		}

		@Override
		public ColorMapper getColorMapper() {
			return mapper;
		}

		@Override
		public void startUrl(net.sourceforge.plantuml.url.Url url) {
			log.add("startUrl " + url.getUrl());
		}

		@Override
		public void closeUrl() {
			log.add("closeUrl");
		}

		@Override
		public void startGroup(UGroup group) {
			log.add("startGroup " + group.asMap());
		}

		@Override
		public void closeGroup() {
			log.add("closeGroup");
		}

		@Override
		public void flushUg() {
		}

		@Override
		public boolean matchesProperty(String propertyName) {
			return svg && "SVG".equals(propertyName);
		}

		@Override
		public HColor getDefaultBackground() {
			return HColors.WHITE;
		}

		@Override
		public void writeToStream(OutputStream os, String metadata, int dpi) throws IOException {
		}
	}

	interface Drawing {
		void draw(UGraphic ug) throws Exception;
	}

	// The lines of a drawing on a recorder, then of an exception it throws.
	static void drawing(List<String> out, String prefix, boolean svg, ColorMapper mapper, Drawing d) {
		final List<String> log = new ArrayList<>();
		try {
			d.draw(Rec.create(log, svg, mapper));
		} catch (Throwable e) {
			log.add(err(e));
		}
		for (String l : log)
			out.add(prefix + "\t" + l);
	}

	@SuppressWarnings("unchecked")
	static <T> T field(Object o, String name) throws Exception {
		final Field f = o.getClass().getDeclaredField(name);
		f.setAccessible(true);
		return (T) f.get(o);
	}

	// ------------------------------------------------------------------
	// SvgPath

	static void svgPaths() throws Exception {
		final List<String> out = new ArrayList<>();
		final List<String> inputs = inputs("svgpath_inputs.txt");
		for (int n = 0; n < inputs.size(); n++) {
			final String data = unesc(inputs.get(n));
			final String p = (n + 1) + "";
			out.add(p + "\tinput\t" + esc(data));
			final SvgPath path;
			try {
				path = new SvgPath(data, UTranslate.none());
			} catch (Throwable e) {
				out.add(p + "\t" + err(e));
				continue;
			}
			final List<String> cmds = new ArrayList<>();
			for (SvgCommand c : SpriteGolden.<List<SvgCommand>>field(path, "commands"))
				cmds.add(c.toSvg());
			out.add(p + "\tcommands\t" + String.join(" ", cmds));
			final List<String> moves = new ArrayList<>();
			for (Movement m : SpriteGolden.<List<Movement>>field(path, "movements"))
				moves.add(m.toSvg());
			out.add(p + "\tmovements\t" + String.join("|", moves));
			out.add(p + "\ttoSvg\t" + path.toSvg());
			drawing(out, p + "\tid", false, ColorMapper.IDENTITY, ug -> path.drawMe(ug, new XAffineTransform()));
			drawing(out, p + "\ts2", false, ColorMapper.IDENTITY,
					ug -> path.drawMe(ug, XAffineTransform.getScaleInstance(2, 2)));
			drawing(out, p + "\trot", false, ColorMapper.IDENTITY, ug -> {
				final XAffineTransform at = XAffineTransform.getTranslateInstance(3, -2);
				at.rotate(Math.PI / 6, 5, 5);
				path.drawMe(ug, at);
			});
			drawing(out, p + "\tmat", false, ColorMapper.IDENTITY, ug -> {
				final XAffineTransform at = XAffineTransform.getScaleInstance(1.5, 1.5);
				at.concatenate(new XAffineTransform(new double[] { 1, 0.5, -0.5, 1, 10, 20 }));
				path.drawMe(ug, at);
			});
			drawing(out, p + "\tf15", false, ColorMapper.IDENTITY, ug -> path.drawMe(ug, 1.5));
			drawing(out, p + "\ttr", false, ColorMapper.IDENTITY,
					ug -> new SvgPath(data, new UTranslate(3, 4)).drawMe(ug, XAffineTransform.getScaleInstance(2, 3)));
		}
		write("svgpath.golden", out);
	}

	// ------------------------------------------------------------------
	// ColorResolver and UGraphicWithScale

	static GrayLevelRange range(int min, int max) {
		return new GrayLevelRange() {
			public int getMinGrayLevel() {
				return min;
			}

			public int getMaxGrayLevel() {
				return max;
			}
		};
	}

	static HColor color(String s) {
		return s == null ? null : HColorSet.instance().getColorOrWhite(s);
	}

	static void ugsStep(List<String> out, String name, UGraphicWithScale u) {
		out.add("ugs\t" + name + "\t" + u.getAffineTransform() + " angle=" + u.getAngle() + " scale="
				+ u.getInitialScale());
	}

	static void emoji() throws Exception {
		final List<String> out = new ArrayList<>();
		final String[] fonts = { null, "#3366CC" };
		final String[] forceds = { null, "#808080", "#FF8000", "#0000FF" };
		final int[][] ranges = { { 0, 255 }, { 50, 200 }, { 999, -1 } };
		final String[] codes = { "red", "#123456", "123456", "#FFF", "currentColor", "none", "NONE", "transparent",
				"foo", "#80FF8000", "black", "white", "#FF0000", "red-blue", "automatic", "?red:blue" };
		for (String font : fonts)
			for (String forced : forceds)
				for (int[] r : ranges) {
					final String p = "resolver\t" + font + "\t" + forced + "\t" + r[0] + "," + r[1];
					final ColorResolver cr = new ColorResolver(color(font), color(forced), range(r[0], r[1]));
					out.add(p + "\tdefault\t" + col(cr.getDefaultColor()));
					for (String code : codes) {
						String v;
						try {
							v = col(cr.getTrueColor(code));
						} catch (Throwable e) {
							v = err(e);
						}
						out.add(p + "\t" + code + "\t" + v);
					}
				}
		for (String bad : new String[] { "red-blue", "?red:blue", "automatic" }) {
			String v;
			try {
				v = col(new ColorResolver(color(bad), null, range(0, 255)).getDefaultColor());
			} catch (Throwable e) {
				v = err(e);
			}
			out.add("resolver-font\t" + bad + "\t" + v);
			try {
				v = col(new ColorResolver(null, color(bad), range(0, 255)).getDefaultColor());
			} catch (Throwable e) {
				v = err(e);
			}
			out.add("resolver-forced\t" + bad + "\t" + v);
		}

		final ColorResolver cr = new ColorResolver(color("#3366CC"), null, range(0, 255));
		drawing(out, "ugs\tdraw", false, ColorMapper.IDENTITY, ug -> {
			UGraphicWithScale u = new UGraphicWithScale(ug, cr, 1.5);
			ugsStep(out, "init", u);
			u = u.applyTranslate(3, -4);
			ugsStep(out, "translate", u);
			u = u.applyScale(2, 2);
			ugsStep(out, "scale", u);
			u = u.applyRotate(30, 5, 6);
			ugsStep(out, "rotate", u);
			u = u.applyMatrix(1, 0.5, -0.5, 1, 10, 20);
			ugsStep(out, "matrix", u);
			u = u.applyRotate(-45, 0, 0);
			ugsStep(out, "rotate0", u);
			u = u.applyScale(0.5, 0.5);
			ugsStep(out, "scale-half", u);
			try {
				u.applyScale(2, 3);
				out.add("ugs\tscale-xy\tok");
			} catch (Throwable e) {
				out.add("ugs\tscale-xy\t" + err(e));
			}
			out.add("ugs\ttruecolor\t" + col(u.getTrueColor("red")) + " " + col(u.getDefaultColor()));
			u.draw(UEllipse.build(3, 4));
			u.apply(new UTranslate(1, 2)).apply(UStroke.withThickness(2)).draw(URectangle.build(5, 6));
			u.apply(HColors.RED).apply(HColors.BLUE.bg()).draw(URectangle.build(1, 1));
		});
		final ColorResolver forced = new ColorResolver(color("#3366CC"), color("#FF8000"), range(10, 100));
		drawing(out, "ugs\tforced", false, ColorMapper.IDENTITY, ug -> {
			final UGraphicWithScale u = new UGraphicWithScale(ug, forced, 2);
			ugsStep(out, "forced-init", u);
			out.add("ugs\tforced-truecolor\t" + col(u.getTrueColor("red")) + " " + col(u.getTrueColor("#FFFFFF")));
			u.draw(UEllipse.build(1, 1));
		});
		final ColorResolver none = new ColorResolver(null, null, range(0, 255));
		drawing(out, "ugs\tnone", false, ColorMapper.IDENTITY,
				ug -> new UGraphicWithScale(ug, none, 1).draw(UEllipse.build(1, 1)));
		write("emoji.golden", out);
	}

	// ------------------------------------------------------------------
	// The parsers

	static class Config {
		final String name;
		final HColor font;
		final HColor forced;
		final double scale;
		final HColor back;
		final boolean svg;

		Config(String name, String font, String forced, double scale, String back, boolean svg) {
			this.name = name;
			this.font = color(font);
			this.forced = color(forced);
			this.scale = scale;
			this.back = color(back);
			this.svg = svg;
		}
	}

	static final Config[] CONFIGS = { //
			new Config("s1", null, null, 1, null, false), //
			new Config("s2", null, null, 2, null, false), //
			new Config("font", "#3366CC", null, 1, null, false), //
			new Config("forced", "#000000", "#FF8000", 1.5, null, false), //
			new Config("gray", null, "#808080", 2, null, false), //
			new Config("back", "red", null, 1, "#FFEEDD", false), //
			new Config("svg", null, null, 1, null, true) };

	static void sprites() throws Exception {
		final List<String> out = new ArrayList<>();

		// The selection of the factory
		final String svg = "<svg viewBox=\"0 0 1 1\"></svg>";
		out.add("factory\tdefault\t" + SvgSpriteParserFactory.create(svg).getClass().getSimpleName());
		out.add("factory\tnull\t"
				+ SvgSpriteParserFactory.create(Collections.singletonList(svg), null, null).getClass().getSimpleName());
		final String[] values = { "<empty>", "sax", " SAX ", "Sax", "nano", "", "  ", "foo", "<novalue>" };
		for (String v : values) {
			final Pragma pragma = Pragma.createEmpty();
			if (v.equals("<novalue>"))
				pragma.define("svgparser", null);
			else if (v.equals("<empty>") == false)
				pragma.define("svgparser", v);
			out.add("factory\t" + v + "\t" + SvgSpriteParserFactory
					.create(Collections.singletonList(svg), pragma, null).getClass().getSimpleName());
		}
		final Map<String, String> md5map = new HashMap<>();
		md5map.put("abc", "QUJD");
		md5map.put("def", "REVG");
		final Pragma sax = Pragma.createEmpty();
		sax.define("svgparser", "sax");
		for (Map<String, String> m : Arrays.asList(md5map, new HashMap<String, String>(), null)) {
			final ISvgSpriteParser parser = SvgSpriteParserFactory.create(
					Arrays.asList("<svg><image href=\"data:image/png;md5,abc\"/>",
							"<image href=\"data:image/png;md5,def\"/><image href=\"data:image/png;md5,xyz\"/></svg>"),
					sax, m);
			out.add("md5\t" + (m == null ? "null" : m.size()) + "\t" + esc(SpriteGolden.<String>field(parser, "svg")));
		}

		// The drawings
		for (String line : inputs("svgsprite_inputs.txt")) {
			final int tab = line.indexOf('\t');
			final String name = line.substring(0, tab);
			final String text = unesc(line.substring(tab + 1));
			for (String kind : new String[] { "nano", "sax" })
				for (Config cfg : CONFIGS)
					drawing(out, name + "\t" + kind + "\t" + cfg.name, cfg.svg, ColorMapper.IDENTITY, ug -> {
						final ISvgSpriteParser parser = kind.equals("nano") ? new SvgNanoParser(text)
								: new SvgSaxParser(text);
						final TextBlock tb = parser.asTextBlock(cfg.font, cfg.forced, cfg.scale, cfg.back);
						final XDimension2D d = tb.calculateDimension(BOUNDER);
						((Rec) ug).log.add("dim " + d.getWidth() + "," + d.getHeight());
						tb.drawU(ug);
					});
		}

		// SpriteSvgNanoParser
		final PortableImage four = SImageIO.read(Base64Coder.decode(FOUR));
		final PortableImage uniform = SImageIO.read(Base64Coder.decode(UNIFORM));
		spriteImage(out, "four-s1", four, null, 1, ColorMapper.IDENTITY);
		spriteImage(out, "four-forced", four, color("#FF8000"), 1, ColorMapper.IDENTITY);
		spriteImage(out, "four-mono", four, color("#FF8000"), 1, ColorMapper.MONOCHROME);
		spriteImage(out, "uniform-s2", uniform, null, 2, ColorMapper.IDENTITY);
		try {
			new SpriteSvgNanoParser(null);
			out.add("spriteimage\tnull\tok");
		} catch (Throwable e) {
			out.add("spriteimage\tnull\t" + err(e));
		}

		write("svgsprite.golden", out);
	}

	static void spriteImage(List<String> out, String name, PortableImage img, HColor forced, double scale,
			ColorMapper mapper) {
		drawing(out, "spriteimage\t" + name, false, mapper, ug -> {
			final Sprite sprite = new SpriteSvgNanoParser(img);
			final TextBlock tb = sprite.asTextBlock(null, forced, scale, color("#FFEEDD"));
			final XDimension2D d = tb.calculateDimension(BOUNDER);
			((Rec) ug).log.add("dim " + d.getWidth() + "," + d.getHeight());
			tb.drawU(ug);
		});
	}

	// ------------------------------------------------------------------
	// The sprite classes

	// The sprites of spriteDrawings, by name; the "uniform" ones have one
	// color, so that their scaled images do not depend on the interpolation.
	static Map<String, Sprite> drawnSprites() throws Exception {
		final Map<String, Sprite> result = new java.util.LinkedHashMap<>();
		final SpriteMonochrome m2 = new SpriteMonochrome(3, 1, 2);
		m2.setGray(1, 0, 1);
		result.put("mono2", m2);
		final SpriteMonochrome m4 = new SpriteMonochrome(4, 4, 4);
		for (int i = 0; i < 4; i++)
			for (int j = 0; j < 4; j++)
				m4.setGray(i, j, (i + j) % 4);
		result.put("mono4", m4);
		final SpriteMonochrome m8 = new SpriteMonochrome(4, 2, 8);
		for (int i = 0; i < 4; i++) {
			m8.setGray(i, 0, i * 2);
			m8.setGray(i, 1, 7 - i);
		}
		result.put("mono8", m8);
		final SpriteMonochrome m16 = new SpriteMonochrome(4, 4, 16);
		for (int i = 0; i < 4; i++)
			for (int j = 0; j < 4; j++)
				m16.setGray(i, j, i * 4 + j);
		result.put("mono16", m16);
		result.put("mono-zero", new SpriteMonochrome(2, 2, 16));
		final SpriteMonochrome full = new SpriteMonochrome(2, 2, 16);
		for (int i = 0; i < 2; i++)
			for (int j = 0; j < 2; j++)
				full.setGray(i, j, 15);
		result.put("mono-uniform", full);
		final SpriteColor c = new SpriteColor(3, 2);
		c.setColor(0, 0, 0xFF0000);
		c.setColor(1, 0, 0x00FF00);
		c.setGray(2, 0, 8);
		c.setGray(0, 1, 15);
		c.setColor(1, 1, 0x123456);
		result.put("color", c);
		final SpriteColor cu = new SpriteColor(2, 2);
		for (int i = 0; i < 2; i++)
			for (int j = 0; j < 2; j++)
				cu.setColor(i, j, 0x336699);
		result.put("color-uniform", cu);
		result.put("image", new SpriteImage(SImageIO.read(Base64Coder.decode(FOUR))));
		result.put("image-uniform", new SpriteImage(SImageIO.read(Base64Coder.decode(UNIFORM))));
		result.put("svgfile", new SpriteSvg(
				"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"10\" height=\"6\"><rect width=\"10\" height=\"6\" fill=\"#00F\"/></svg>"));
		result.put("stdlib", new StdlibSprite(3, 3,
				new byte[] { (byte) 0x0F, (byte) 0xF0, (byte) 0x88, (byte) 0x1E, (byte) 0x2D, (byte) 0x3C }));
		return result;
	}

	static void spriteDrawings() throws Exception {
		final List<String> out = new ArrayList<>();
		for (Map.Entry<String, Sprite> ent : drawnSprites().entrySet()) {
			final String name = ent.getKey();
			final Sprite sprite = ent.getValue();
			// configuration: font color, forced color, back color of the graphic, mapper, scale
			final Object[][] cfgs = { { "plain", null, null, null, ColorMapper.IDENTITY, 1.0 },
					{ "font", color("#FF0000"), null, null, ColorMapper.IDENTITY, 1.0 },
					{ "forced", color("#FF0000"), color("#0000FF"), null, ColorMapper.IDENTITY, 1.0 },
					{ "back", color("#00FF00"), null, color("#FFEEDD"), ColorMapper.IDENTITY, 1.0 },
					{ "mono", color("#FF8000"), null, null, ColorMapper.MONOCHROME, 1.0 },
					{ "s2", color("#FF0000"), null, null, ColorMapper.IDENTITY, 2.0 } };
			for (Object[] cfg : cfgs) {
				final boolean uniform = name.endsWith("-uniform") || name.equals("svgfile");
				final double scale = (Double) cfg[5];
				drawing(out, name + "\t" + cfg[0], false, (ColorMapper) cfg[4], ug -> {
					if (cfg[3] != null)
						ug = ug.apply(((HColor) cfg[3]).bg());
					final TextBlock tb = sprite.asTextBlock((HColor) cfg[1], (HColor) cfg[2], scale, null);
					final XDimension2D d = tb.calculateDimension(BOUNDER);
					((Rec) ug).log.add("dim " + d.getWidth() + "," + d.getHeight());
					// A scaled image depends on the interpolation, which differs
					// slightly in Go: only uniform sprites are drawn scaled.
					if (scale == 1 || uniform)
						tb.drawU(ug);
				});
			}
		}
		write("spritedraw.golden", out);
	}
}
