// SPDX-License-Identifier: MIT
// Provenance: original
//
// KlimtGolden2 dumps more golden values of the internal/klimt tests: drawing
// recordings, Java hash orders, Display operations, font configurations,
// strokes, DotPath, geometry, style parser edge cases and pixel images. Build
// and run it as KlimtGolden (see there), with the class name KlimtGolden2 and
// the same JVM options (-XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives).

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.klimt.Fashion;
import net.sourceforge.plantuml.klimt.UBackground;
import net.sourceforge.plantuml.klimt.UChange;
import net.sourceforge.plantuml.klimt.UClip;
import net.sourceforge.plantuml.klimt.UGroup;
import net.sourceforge.plantuml.klimt.UGroupType;
import net.sourceforge.plantuml.klimt.UMotif;
import net.sourceforge.plantuml.klimt.UParam;
import net.sourceforge.plantuml.klimt.UPath;
import net.sourceforge.plantuml.klimt.UPattern;
import net.sourceforge.plantuml.klimt.UShape;
import net.sourceforge.plantuml.klimt.UStroke;
import net.sourceforge.plantuml.klimt.UTranslate;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColors;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.drawing.UGraphicStencil;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontPosition;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontFace;
import net.sourceforge.plantuml.klimt.font.UFontFactory;
import net.sourceforge.plantuml.klimt.geom.HorizontalAlignment;
import net.sourceforge.plantuml.klimt.geom.PointAndAngle;
import net.sourceforge.plantuml.klimt.geom.RectangleArea;
import net.sourceforge.plantuml.klimt.geom.USegment;
import net.sourceforge.plantuml.klimt.geom.VerticalAlignment;
import net.sourceforge.plantuml.klimt.geom.XCubicCurve2D;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.shape.BigFrame;
import net.sourceforge.plantuml.klimt.shape.CircledCharacter;
import net.sourceforge.plantuml.klimt.shape.DotPath;
import net.sourceforge.plantuml.klimt.shape.GraphicStrings;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.TextBlockArrow;
import net.sourceforge.plantuml.klimt.shape.TextBlockArrow2;
import net.sourceforge.plantuml.klimt.shape.TextBlockLineBefore;
import net.sourceforge.plantuml.klimt.shape.TextBlockRaw;
import net.sourceforge.plantuml.klimt.shape.TextBlockUtils;
import net.sourceforge.plantuml.klimt.shape.UCenteredCharacter;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UEmpty;
import net.sourceforge.plantuml.klimt.shape.UHorizontalLine;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.skin.Padder;
import net.sourceforge.plantuml.skin.VisibilityModifier;
import net.sourceforge.plantuml.stereo.Stereogroup;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.style.ClockwiseTopRightBottomLeft;
import net.sourceforge.plantuml.style.PName;
import net.sourceforge.plantuml.style.SName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.style.StyleBuilder;
import net.sourceforge.plantuml.style.StyleLoader;
import net.sourceforge.plantuml.style.StyleQuery;
import net.sourceforge.plantuml.klimt.LineBreakStrategy;
import net.sourceforge.plantuml.svek.GuideLine;
import net.sourceforge.plantuml.svek.PackageStyle;
import net.sourceforge.plantuml.svek.Ports;
import net.sourceforge.plantuml.url.Url;
import net.sourceforge.plantuml.utils.BlocLines;
import net.sourceforge.plantuml.utils.Direction;

public class KlimtGolden2 {

	static Path dir;

	public static void main(String[] args) throws Exception {
		dir = Paths.get(args[0]);
		drawings();
		hashOrder();
		displayOps();
		fontConfigs();
		strokes();
		dotPaths();
		geometry();
		stereoStyles();
		parserCases();
		pixels();
	}

	// ------------------------------------------------------------------
	// Pixel images

	static String pix(net.sourceforge.plantuml.klimt.awt.PortableImage img) {
		final List<String> l = new ArrayList<>();
		for (int i = 0; i < img.getWidth(); i++)
			l.add(Integer.toHexString(img.getRGB(i, 0)));
		return String.join(",", l);
	}

	static void pixels() throws Exception {
		final List<String> out = new ArrayList<>();
		final int[] values = { 0xFF102030, 0x80FFFFFF, 0x00000000, 0xFF000000, 0xFFFFFFFF, 0xFF7F7F7F, 0x40102030 };
		final net.sourceforge.plantuml.klimt.awt.PortableImage img = net.sourceforge.plantuml.klimt.awt.PortableImageFactory
				.build(values.length, 1, net.sourceforge.plantuml.klimt.awt.PortableImage.TYPE_INT_ARGB);
		for (int i = 0; i < values.length; i++)
			img.setRGB(i, 0, values[i]);
		final net.atmp.PixelImage p = new net.atmp.PixelImage(img, net.sourceforge.plantuml.klimt.AffineTransformType.TYPE_BILINEAR);
		out.add("source\t" + pix(p.getImage()));
		out.add("monochrome\t" + pix(p.monochrome().getImage()));
		out.add("mutecolor\t" + pix(p.muteColor(new net.sourceforge.plantuml.klimt.awt.XColor(10, 200, 30)).getImage()));
		out.add("mutetransparent\t"
				+ pix(p.muteTransparentColor(new net.sourceforge.plantuml.klimt.awt.XColor(1, 2, 3)).getImage()));
		write("pixels.golden", out);
	}

	static void write(String name, List<String> lines) throws IOException {
		Files.write(dir.resolve(name), (String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
	}

	static String hc(HColor c) {
		return c == null ? "null" : c.asString();
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

	static FontConfiguration fc(int size) {
		return FontConfiguration.create(UFontFactory.sansSerif(size), HColors.BLACK, HColors.BLUE, UStroke.simple());
	}

	static TextBlock raw(int size, String... lines) {
		return new TextBlockRaw(Arrays.asList(lines), fc(size));
	}

	static String uf(UFont f) {
		return f.getFamily(null, net.sourceforge.plantuml.klimt.font.UFontContext.SVG) + "/" + f.getFontFace() + "/"
				+ f.getSize();
	}

	// ------------------------------------------------------------------
	// A UGraphic that records the drawing operations.

	static class Rec implements UGraphic {
		final List<String> log;
		final UTranslate t;
		final HColor color;
		final HColor back;
		final UStroke stroke;

		Rec(List<String> log, UTranslate t, HColor color, HColor back, UStroke stroke) {
			this.log = log;
			this.t = t;
			this.color = color;
			this.back = back;
			this.stroke = stroke;
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
			log.add("@" + t.getDx() + "," + t.getDy() + " c=" + hc(color) + " b=" + hc(back) + " s=" + stroke + " "
					+ describe(shape));
		}

		@Override
		public UGraphic apply(UChange change) {
			if (change instanceof UTranslate)
				return new Rec(log, ((UTranslate) change).compose(t), color, back, stroke);
			if (change instanceof HColor)
				return new Rec(log, t, (HColor) change, back, stroke);
			if (change instanceof UBackground)
				return new Rec(log, t, color, ((UBackground) change).getBackColor(), stroke);
			if (change instanceof UStroke)
				return new Rec(log, t, color, back, (UStroke) change);
			log.add("apply " + change.getClass().getSimpleName());
			return this;
		}

		@Override
		public ColorMapper getColorMapper() {
			return ColorMapper.IDENTITY;
		}

		@Override
		public void startUrl(Url url) {
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
			return false;
		}

		@Override
		public HColor getDefaultBackground() {
			return HColors.WHITE;
		}

		@Override
		public void writeToStream(OutputStream os, String metadata, int dpi) throws IOException {
		}
	}

	static String path(UPath p) {
		final List<String> segs = new ArrayList<>();
		for (USegment s : p)
			segs.add(s.toString());
		return segs.toString();
	}

	static String describe(Object shape) {
		if (shape instanceof URectangle) {
			final URectangle r = (URectangle) shape;
			return "rect " + r.getWidth() + "," + r.getHeight() + " r=" + r.getRx() + "," + r.getRy() + " sh="
					+ r.getDeltaShadow();
		}
		if (shape instanceof UEllipse) {
			final UEllipse e = (UEllipse) shape;
			return "ellipse " + e.getWidth() + "," + e.getHeight() + " " + e.getStart() + "," + e.getExtend() + " sh="
					+ e.getDeltaShadow();
		}
		if (shape instanceof ULine) {
			final ULine l = (ULine) shape;
			return "line " + l.getDX() + "," + l.getDY();
		}
		if (shape instanceof UPolygon) {
			final UPolygon p = (UPolygon) shape;
			final List<String> pts = new ArrayList<>();
			for (XPoint2D pt : p.getPoints())
				pts.add(pt.getX() + ":" + pt.getY());
			return "polygon " + pts + " sh=" + p.getDeltaShadow();
		}
		if (shape instanceof UPath) {
			final UPath p = (UPath) shape;
			return "path " + path(p) + " sh=" + p.getDeltaShadow();
		}
		if (shape instanceof UText) {
			final UText t = (UText) shape;
			return "text '" + t.getText() + "' " + uf(t.getFontConfiguration().getFont()) + " "
					+ hc(t.getFontConfiguration().getColor());
		}
		if (shape instanceof DotPath)
			return "dotpath " + shape;
		if (shape instanceof UEmpty) {
			final UEmpty e = (UEmpty) shape;
			return "empty " + e.getWidth() + "," + e.getHeight();
		}
		if (shape instanceof UCenteredCharacter) {
			final UCenteredCharacter c = (UCenteredCharacter) shape;
			return "char " + c.getChar() + " " + uf(c.getFont());
		}
		if (shape instanceof UHorizontalLine)
			return "hline";
		return shape.getClass().getSimpleName();
	}

	static UGraphic rec(List<String> log) {
		return new Rec(log, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple());
	}

	interface Drawing {
		void draw(UGraphic ug) throws Exception;
	}

	static void drawing(List<String> out, String name, Drawing d) {
		final List<String> log = new ArrayList<>();
		try {
			d.draw(rec(log));
		} catch (Exception e) {
			log.add("ERR " + e.getClass().getSimpleName());
		}
		for (String l : log)
			out.add(name + "\t" + l);
	}

	static void drawings() throws Exception {
		final List<String> out = new ArrayList<>();
		drawing(out, "raw", ug -> raw(14, "Hello", "Tab\there", "").drawU(ug));
		drawing(out, "error", ug -> GraphicStrings.createForError(Arrays.asList("Syntax", "line 2"), false).drawU(ug));
		drawing(out, "bordered", ug -> TextBlockUtils.bordered(raw(12, "note"), UStroke.withThickness(1.5),
				HColors.RED, HColors.MY_YELLOW, 5, ClockwiseTopRightBottomLeft.topRightBottomLeft(2, 3, 4, 5), "id")
				.drawU(ug));
		drawing(out, "borderednoback", ug -> TextBlockUtils.bordered(raw(12, "n"), UStroke.withThickness(0),
				HColors.RED, HColors.WHITE, 0, ClockwiseTopRightBottomLeft.same(1), null).drawU(ug));
		drawing(out, "vertical", ug -> TextBlockUtils
				.mergeTB(TextBlockUtils.addBackcolor(raw(12, "colored"), HColors.RED), raw(10, "wider plain"),
						HorizontalAlignment.CENTER)
				.drawU(ug));
		drawing(out, "horizontal", ug -> TextBlockUtils.mergeLR(raw(20, "Big"), raw(10, "small"),
				VerticalAlignment.CENTER).drawU(ug));
		drawing(out, "margin", ug -> TextBlockUtils.withMargin(raw(12, "m"), 3, 4).drawU(ug));
		drawing(out, "minwidth", ug -> TextBlockUtils.withMinWidth(raw(12, "x"), 60, HorizontalAlignment.CENTER)
				.drawU(ug));
		drawing(out, "linebefore", ug -> new TextBlockLineBefore(1.5, raw(12, "after"), '=').drawU(
				UGraphicStencil.create(ug, new XDimension2D(80, 30))));
		drawing(out, "linebeforedotted", ug -> new TextBlockLineBefore(1.5, raw(12, "after"), '.').drawU(
				UGraphicStencil.create(ug, new XDimension2D(80, 30)).apply(new UTranslate(5, 7))));
		drawing(out, "linebeforetitle", ug -> new TextBlockLineBefore(1.5, raw(12, "after"), '-', raw(10, "Title"))
				.drawU(UGraphicStencil.create(ug, new XDimension2D(100, 30))));
		drawing(out, "linebeforedefault", ug -> new TextBlockLineBefore(2.5, raw(12, "after"), '\0')
				.drawU(UGraphicStencil.create(ug, new XDimension2D(50, 30))));
		drawing(out, "linebeforenostencil", ug -> new TextBlockLineBefore(2.5, raw(12, "after"), '_').drawU(ug));
		drawing(out, "bigframe", ug -> new BigFrame(raw(12, "frame"), raw(12, "line one", "line two"),
				ClockwiseTopRightBottomLeft.same(3), new Fashion(HColors.WHITE, HColors.BLACK).withShadow(2)
						.withCorner(4, 0))
				.drawU(ug));
		drawing(out, "bigframelongtitle", ug -> new BigFrame(raw(12, "a much longer frame title"), raw(12, "x"),
				ClockwiseTopRightBottomLeft.none(), new Fashion(null, HColors.BLUE)).drawU(ug));
		for (PackageStyle p : PackageStyle.values()) {
			drawing(out, "package " + p, ug -> p.drawU(ug, new XDimension2D(60, 40), new XDimension2D(20, 10), true));
		}
		drawing(out, "package frame notitle",
				ug -> PackageStyle.FRAME.drawU(ug, new XDimension2D(60, 40), null, false));
		for (VisibilityModifier v : VisibilityModifier.values())
			drawing(out, "visibility " + v, ug -> v.getUBlock(10, HColors.BLACK, HColors.GREEN, true).drawU(ug));
		drawing(out, "visibility drawable",
				ug -> VisibilityModifier.PRIVATE_FIELD.getUDrawable(8, HColors.RED, null).drawU(ug));
		drawing(out, "circled", ug -> new CircledCharacter('C', 9, UFontFactory.monospace(17), HColors.GREEN,
				HColors.BLACK, HColors.BLACK).drawU(ug));
		for (Direction d : Direction.values())
			drawing(out, "arrow " + d, ug -> new TextBlockArrow(d, fc(13)).drawU(ug));
		drawing(out, "arrow2", ug -> new TextBlockArrow2(new GuideLine() {
			public Direction getArrowDirection() {
				return Direction.RIGHT;
			}

			public double getArrowDirectionInRadian() {
				return 0.7;
			}
		}, fc(14)).drawU(ug));
		drawing(out, "padder", ug -> Padder.NONE.withMargin(ClockwiseTopRightBottomLeft.same(2))
				.withPadding(ClockwiseTopRightBottomLeft.margin1margin2(3, 4)).withBackgroundColor(HColors.RED)
				.withRoundCorner(6).apply(raw(12, "padded")).drawU(ug));
		drawing(out, "motif", ug -> new UMotif("AAGGHH").drawHorizontal(ug, 2, 3, 2));
		drawing(out, "motifrect", ug -> ug.draw(new UMotif(-2, 0, 0, 2, 2, 0).getRectangle(20, 10)));
		drawing(out, "diagonal", ug -> ug.draw(URectangle.build(30, 20).diagonalCorner(5)));
		drawing(out, "halfrounded", ug -> ug.draw(URectangle.build(30, 20).halfRounded(8)));
		drawing(out, "fashion", ug -> new Fashion(HColors.RED, null).withStroke(new UStroke(3, 4, 2)).apply(ug)
				.draw(URectangle.build(5, 5)));
		drawing(out, "url", ug -> net.sourceforge.plantuml.klimt.shape.TextBlockWithUrl
				.withUrl(raw(12, "link"), new Url("http://x.org", null, null)).drawU(ug));
		drawing(out, "style", ug -> {
			final StyleBuilder b = StyleLoader.loadSkin("plantuml.skin");
			final Style s = b.getMergedStyle(StyleQuery.of3(SName.root, SName.document, SName.legend));
			s.applyStrokeAndLineColor(ug, HColorSet.instance()).draw(new ULine(10, 0));
			s.getSymbolContext(HColorSet.instance()).apply(ug).draw(UEllipse.build(4, 4));
		});
		drawing(out, "rotate", ug -> {
			final UPolygon p = new UPolygon();
			p.addPoint(0, 0);
			p.addPoint(10, 0);
			p.addPoint(5, 7);
			p.rotate(Math.PI / 6);
			ug.draw(p);
			ug.draw(new ULine(10, 3).rotate(1.1));
			final UPath path = UPath.none();
			path.moveTo(1, 2);
			path.lineTo(8, 3);
			ug.draw(path.rotate(-0.4));
			ug.apply(new UTranslate(3, 4).rotate(2.5)).draw(new UEmpty(1, 1));
		});
		write("drawing.golden", out);
	}

	// ------------------------------------------------------------------
	// Java hash orders

	static void hashOrder() throws Exception {
		final List<String> out = new ArrayList<>();
		for (int n : new int[] { 3, 12, 13, 30, 70 }) {
			final Map<String, Integer> map = new HashMap<>();
			final List<String> keys = new ArrayList<>();
			for (int i = 0; i < n; i++) {
				final String k = Ports.encodePortNameToId("port" + i);
				keys.add(k);
				map.put(k, i);
			}
			out.add("ports" + n + "\t" + String.join(",", keys) + "\t" + String.join(",", map.keySet()));
		}
		final List<String> words = Arrays.asList("participant", "actor", "boundary", "a", "b", "Aa", "BB", "zz",
				"<U+00E9>", "x", "é", "日本", "ab", "ba", "polygenelubricants", "GydZG_", "DESIGNING WORKHOUSES");
		final Set<String> set = new HashSet<>();
		for (String w : words)
			set.add(w);
		out.add("words\t" + String.join(",", words) + "\t" + String.join(",", set));
		final Set<XPoint2D> pts = new HashSet<>();
		final List<String> ins = new ArrayList<>();
		for (int i = 0; i < 20; i++) {
			final XPoint2D p = new XPoint2D(i * 1.5 - 7, (i * 7) % 11 - 0.25);
			pts.add(p);
			ins.add(p.getX() + ":" + p.getY());
		}
		final List<String> ord = new ArrayList<>();
		for (XPoint2D p : pts)
			ord.add(p.getX() + ":" + p.getY());
		out.add("points\t" + String.join(",", ins) + "\t" + String.join(",", ord));
		final DotPath dp = new DotPath().addCurve(new XPoint2D(0, 0), new XPoint2D(10, -30), new XPoint2D(40, 50),
				new XPoint2D(60, 5));
		final List<String> sample = new ArrayList<>();
		for (XPoint2D p : dp.sample())
			sample.add(p.getX() + ":" + p.getY());
		out.add("sample\t\t" + String.join(",", sample));
		final Ports ports = new Ports();
		ports.add("a", 1, 10, 5);
		ports.add("b", 2, 3, 5);
		ports.add("c", 1, 10, 7);
		ports.add("a", 0, 1, 1);
		ports.add("d", 3, 10, 2);
		final Ports more = new Ports();
		more.add("a", 5, 20, 1);
		more.add("e", 1, -2, 1);
		ports.addThis(more.translateY(4));
		out.add("portsgeom\t\t" + ports.getAllPortGeometry() + " " + ports);
		write("hashorder.golden", out);
	}

	// ------------------------------------------------------------------
	// Display operations

	static String disp(Display d) {
		final List<String> parts = new ArrayList<>();
		for (CharSequence cs : d)
			parts.add(cs.toString());
		return d.size() + ":" + String.join("|", parts) + ":" + d.getNaturalHorizontalAlignment();
	}

	static void displayOps() throws Exception {
		final List<String> out = new ArrayList<>();
		final Display d = Display.create("first : type", "+field", "<<stereo>>", "page %page% of %lastpage%");
		out.add("create\t" + disp(d));
		out.add("underlined\t" + disp(d.underlined()));
		out.add("underlinedName\t" + disp(d.underlinedName()));
		out.add("guillemet\t" + disp(Display.create("+private <<a>> <<b>>", "x <<c>>").manageGuillemet(true)));
		out.add("guillemetnovis\t" + disp(Display.create("+private <<a>>", "x").manageGuillemet(false)));
		out.add("withPage\t" + disp(d.withPage(3, 7)));
		out.add("replace\t" + disp(d.replace("e", "E")));
		out.add("removeEndingStereotype\t" + disp(Display.create("a", "name <<st>>").removeEndingStereotype()));
		out.add("endingStereotype\t" + Display.create("name << st >>").getEndingStereotype());
		out.add("addFirst\t" + disp(d.addFirst("zero")));
		out.add("appendFirstLine\t" + disp(d.appendFirstLine(">>")));
		out.add("addGeneric\t" + disp(Display.create("List").addGeneric("T")));
		out.add("addGenericEmpty\t" + disp(Display.empty().addGeneric("T")));
		out.add("addAll\t" + disp(d.addAll(Display.create("x", "y"))));
		out.add("subList\t" + disp(d.subList(1, 3)));
		final List<String> splits = new ArrayList<>();
		for (Display part : Display.create("a--b", "c", "d--e--f").splitMultiline(
				net.sourceforge.plantuml.regex.Pattern2.cmpile("--")))
			splits.add(disp(part));
		out.add("splitMultiline\t" + String.join(" ; ", splits));
		out.add("replaceBackslashT\t" + disp(Display.create("a\\tb").replaceBackslashT()).replace("\t", "<TAB>"));
		out.add("contentWidth\t" + Display.create("ab", "日本語x", "😀😀").contentWidth());
		out.add("isWhite\t" + Display.create("   ").isWhite() + "," + Display.create(" ", "").isWhite() + ","
				+ Display.empty().isWhite());
		out.add("guidelines\t" + Display.hasSeveralGuideLines("a\\n< b") + "," + Display.hasSeveralGuideLines("a < b")
				+ "," + Display.create("x", "y >").hasSeveralGuideLines());
		out.add("equals\t" + d.equals(Display.create("first : type", "+field", "<<stereo>>", "page %page% of %lastpage%"))
				+ "," + d.hashCode() + "," + Display.NULL.hashCode() + "," + Display.isNull(Display.NULL));
		out.add("tooltip\t" + d.toTooltipText() + "," + Display.empty().toTooltipText());
		out.add("foo\t" + disp(Display.createFoo(Arrays.asList(
				new net.sourceforge.plantuml.text.StringLocated("a", null),
				new net.sourceforge.plantuml.text.StringLocated("}}", null),
				new net.sourceforge.plantuml.text.StringLocated("", null)))));
		write("displayops.golden", out);
	}

	// ------------------------------------------------------------------
	// Font configurations

	static String fcs(FontConfiguration f) {
		return uf(f.getFont()) + " " + hc(f.getColor()) + " " + hc(f.getExtendedColor()) + " " + f.getSpace() + " "
				+ f.containsStyle(FontStyle.UNDERLINE) + " " + f.getUnderlineStroke() + " " + f.getTabSize() + " "
				+ f.getAttributes();
	}

	static void fontConfigs() throws Exception {
		final List<String> out = new ArrayList<>();
		final FontConfiguration base = FontConfiguration.create(UFontFactory.build("Courier", UFontFace.bold(), 13),
				HColors.RED, HColors.BLUE, UStroke.withThickness(2), 4);
		out.add("base\t" + fcs(base));
		out.add("italic\t" + fcs(base.italic()));
		out.add("unbold\t" + fcs(base.unbold()));
		out.add("plain\t" + fcs(base.add(FontStyle.PLAIN)));
		out.add("bigger\t" + fcs(base.bigger(2.7)));
		out.add("smaller\t" + fcs(base.bigger(-20)));
		out.add("changeSize\t" + fcs(base.changeSize(9.9f)));
		out.add("exposant\t" + fcs(base.changeFontPosition(FontPosition.EXPOSANT)));
		out.add("indice\t" + fcs(base.changeFontPosition(FontPosition.INDICE)));
		out.add("hyperlink\t" + fcs(base.hyperlink()));
		out.add("wave\t" + fcs(base.wave(HColors.GREEN)));
		out.add("family\t" + fcs(base.changeFamily("Serif")));
		out.add("forceFontNull\t" + fcs(base.forceFont(null, null)));
		out.add("forceFont\t" + fcs(base.forceFont(UFontFactory.monospace(10), HColors.GRAY)));
		out.add("reset\t" + fcs(base.italic().bigger(5).changeColor(HColors.GREEN).resetFont()));
		out.add("attrs\t" + fcs(base.changeAttributes(net.sourceforge.plantuml.klimt.SvgAttributes
				.build("a=1 b=\"two words\" c = x"))));
		out.add("faceweight\t" + fcs(base.withFontFace(UFontFace.normal().withWeight(300))));
		out.add("equals\t" + base.equals(base.italic().unitalic()) + "," + base.equals(base.italic()));
		out.add("blackblue\t" + fcs(FontConfiguration.blackBlueTrue(UFontFactory.serif(11))));
		write("fontconfig.golden", out);
	}

	// ------------------------------------------------------------------
	// Strokes, paddings, line breaks

	static void strokes() throws Exception {
		final List<String> out = new ArrayList<>();
		final StyleBuilder b = StyleLoader.loadSkin("plantuml.skin");
		final Style root = b.getMergedStyle(StyleQuery.of3(SName.root));
		for (String ls : new String[] { "", "7;7", "1-3", "x", "5", "-", "2,8", " 3 - 4 ", "1.5;2.5;9", ";;", "1e2-3" }) {
			String r;
			try {
				r = root.eventuallyOverride(PName.LineStyle, ls).getStroke().toString();
			} catch (Exception e) {
				r = "ERR " + e.getClass().getSimpleName();
			}
			out.add("linestyle\t" + ls + "\t" + r);
		}
		out.add("overridestroke\t\t" + root.eventuallyOverride(new UStroke(4, 2, 3)).getStroke());
		for (String p : new String[] { "5", "1 2", "1 2 3", "1 2 3 4", "1 2 3 4 5", " 7 ", "", "a", "1.5",
				"99999999999", "  1   2  " })
			out.add("padding\t" + p + "\t" + ClockwiseTopRightBottomLeft.read(p));
		for (String l : new String[] { "100", "-20", "auto", "AUTO", "", "12px", "007" })
			out.add("linebreak\t" + l + "\t" + new LineBreakStrategy(l).getMaxWidth() + ","
					+ new LineBreakStrategy(l).isAuto());
		out.add("linebreak\tnull\t" + LineBreakStrategy.NONE.getMaxWidth() + "," + LineBreakStrategy.NONE.isAuto());
		final List<String> vals = Arrays.asList("12", "1.5", "abc", "a1b2", "", "  bold ", "italic", "600",
				"bolder", "true", "TRUE", "center", "1.2.3");
		for (String v : vals) {
			final Style s = root.eventuallyOverride(PName.FontSize, v);
			final net.sourceforge.plantuml.style.value.Value val = s.value(PName.FontSize);
			out.add("value\t" + v + "\t" + item(() -> val.asInt()) + "," + item(() -> val.asIntButMinusOneIfError())
					+ "," + item(() -> val.asDouble()) + "," + item(() -> val.asBoolean()) + ","
					+ item(() -> val.asFontFace()) + "," + item(() -> val.asHorizontalAlignment()) + ","
					+ item(() -> val.asColor(HColorSet.instance()).asString()));
		}
		write("strokes.golden", out);
	}

	interface Item {
		Object get() throws Exception;
	}

	static String item(Item i) {
		try {
			return String.valueOf(i.get());
		} catch (Exception e) {
			return "ERR " + e.getClass().getSimpleName();
		}
	}

	// ------------------------------------------------------------------
	// DotPath

	static String pts(DotPath p) {
		return p.toString();
	}

	static void dotPaths() throws Exception {
		final List<String> out = new ArrayList<>();
		final DotPath p = new DotPath().addCurve(new XPoint2D(0, 0), new XPoint2D(10, 0), new XPoint2D(20, 5),
				new XPoint2D(30, 5)).addCurve(new XPoint2D(40, 5), new XPoint2D(50, 20), new XPoint2D(60, 40));
		out.add("path\t" + pts(p));
		final PointAndAngle m = p.getMiddle();
		out.add("middle\t" + m.getX() + "," + m.getY() + "," + m.getAngle());
		out.add("angles\t" + p.getStartAngle() + "," + p.getEndAngle());
		out.add("minmax\t" + p.getMinMax());
		out.add("mindist\t" + p.getMinDist(new XPoint2D(25, -3)));
		out.add("reverse\t" + pts(p.reverse()));
		out.add("upath\t" + pathMM(p.toUPath()));
		out.add("isline\t" + p.isLine() + "," + new DotPath().addCurve(new XPoint2D(0, 0), new XPoint2D(1, 1),
				new XPoint2D(2, 2), new XPoint2D(3, 3)).isLine());
		final DotPath moved = p.copy();
		moved.moveStartPoint(2, 3);
		moved.moveEndPoint(-1, 4);
		out.add("moved\t" + pts(moved));
		final DotPath movedFar = p.copy();
		movedFar.moveStartPoint(40, 1);
		out.add("movedfar\t" + pts(movedFar));
		final DotPath delta = p.copy();
		delta.moveDelta(5, -5);
		out.add("delta\t" + pts(delta));
		out.add("compound\t" + pts(p.simulateCompound(new RectangleArea(50, 15, 70, 50), new RectangleArea(-5, -5, 12, 3))));
		final DotPath ortho = new DotPath().addCurve(new XPoint2D(0, 0), new XPoint2D(0, 0), new XPoint2D(0, 20),
				new XPoint2D(0, 20)).addCurve(new XPoint2D(0, 20), new XPoint2D(30, 20), new XPoint2D(30, 20))
				.addCurve(new XPoint2D(30, 20), new XPoint2D(30, 50), new XPoint2D(30, 50));
		ortho.muteToRoundOrthogonalPaths(6);
		out.add("rounded\t" + pts(ortho));
		write("dotpath.golden", out);
	}

	static String pathMM(UPath p) {
		final List<String> segs = new ArrayList<>();
		for (USegment s : p)
			segs.add(s.toString());
		return segs.toString() + " " + p.getMinX() + "," + p.getMinY() + "," + p.getMaxX() + "," + p.getMaxY();
	}

	// ------------------------------------------------------------------
	// Geometry

	static void geometry() throws Exception {
		final List<String> out = new ArrayList<>();
		final UEllipse e = new UEllipse(30, 20, 0, 0);
		for (double y : new double[] { 0, 3, 10, 17.5, 20 })
			out.add("ellipsex\t" + y + "\t" + e.getStartingX(y) + "," + e.getEndingX(y));
		for (double a : new double[] { 0, 0.3, 1, Math.PI / 2, 2.5, -1.2, 10 }) {
			final XPoint2D p = e.getPointAtAngle(a);
			out.add("ellipseangle\t" + a + "\t" + p.getX() + "," + p.getY());
		}
		for (double a : new double[] { 0.1, 0.7, 1.3, 2.9, -0.5, 4.4 }) {
			final ULine l = new ULine(13, 7).rotate(a);
			final UTranslate t = new UTranslate(5, -2).rotate(a);
			out.add("rotate\t" + a + "\t" + l.getDX() + "," + l.getDY() + "," + t.getDx() + "," + t.getDy());
		}
		out.add("linelength\t\t" + new ULine(3, 4).getLength() + "," + new ULine(0.1, 0.2).getLength());
		final net.sourceforge.plantuml.klimt.awt.XAffineTransform at = new net.sourceforge.plantuml.klimt.awt.XAffineTransform();
		at.translate(3, 4);
		at.rotate(0.9, 2, 1);
		at.scale(1.5, 0.5);
		at.concatenate(net.sourceforge.plantuml.klimt.awt.XAffineTransform.getRotateInstance(-0.3));
		out.add("affine\t\t" + at);
		final XPoint2D tp = at.transform(new XPoint2D(7, -3));
		out.add("affinept\t\t" + tp.getX() + "," + tp.getY());
		final UPath path = UPath.none();
		path.moveTo(1, 1);
		path.arcTo(new XPoint2D(9, 9), 4, 0, 1);
		path.lineTo(2, 8);
		out.add("affinepath\t\t" + pathMM(path.affine(at, 0.25, 2)));
		final UClip clip = new UClip(0, 0, 10, 20);
		out.add("clip\t\t" + line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(-5, 3, 15, 3)))
				+ " " + line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(4, -5, 4, 30))) + " "
				+ line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(-5, -5, 15, 30))) + " "
				+ line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(-5, -5, -5, 30))) + " "
				+ line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(2, 3, 4, 5))) + " "
				+ line(clip.getClippedLine(new net.sourceforge.plantuml.klimt.geom.XLine2D(3, 3, 30, 3))) + " "
				+ rect(clip.getClippedRectangle(new net.sourceforge.plantuml.klimt.geom.XRectangle2D(5, 5, 20, 20))));
		final UGroup g = new UGroup();
		g.put(UGroupType.TITLE, "a/b<c> d-e_f");
		g.put(UGroupType.CLASS, "x😀y");
		out.add("ugroup\t\t" + g.asMap());
		write("geometry.golden", out);
	}

	static String line(net.sourceforge.plantuml.klimt.geom.XLine2D l) {
		return l == null ? "null" : l.getX1() + "," + l.getY1() + "," + l.getX2() + "," + l.getY2();
	}

	static String rect(net.sourceforge.plantuml.klimt.geom.XRectangle2D r) {
		return r.getX() + "," + r.getY() + "," + r.getWidth() + "," + r.getHeight();
	}

	// ------------------------------------------------------------------
	// Style parser edge cases

	static String unesc(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '\\' && i + 1 < s.length()) {
				final char n = s.charAt(++i);
				sb.append(n == 'n' ? '\n' : n);
			} else
				sb.append(c);
		}
		return sb.toString();
	}

	// As KlimtGolden.canonical: the atoms and the level constraint of a query
	// are package-private.
	static String canonical(StyleQuery sig) {
		try {
			final java.lang.reflect.Method atoms = StyleQuery.class.getDeclaredMethod("getAtoms");
			final java.lang.reflect.Method level = StyleQuery.class.getDeclaredMethod("getLevelConstraint");
			atoms.setAccessible(true);
			level.setAccessible(true);
			final List<String> names = new ArrayList<>();
			final java.util.TreeSet<String> st = new java.util.TreeSet<>();
			for (net.sourceforge.plantuml.style.StyleAtom atom : (Iterable<net.sourceforge.plantuml.style.StyleAtom>) atoms
					.invoke(sig))
				if (atom.isName())
					names.add(atom.getSName().name());
				else
					st.add(atom.getStereotype());
			final net.sourceforge.plantuml.style.LevelConstraint lc = (net.sourceforge.plantuml.style.LevelConstraint) level
					.invoke(sig);
			return String.join(".", names) + st + (lc.hasLevel() ? "#" + lc.getLevel() : "") + (lc.isStar() ? "*" : "");
		} catch (ReflectiveOperationException e) {
			throw new IllegalStateException(e);
		}
	}

	static void parserCases() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String line : Files.readAllLines(dir.resolve("parser_inputs.txt"), StandardCharsets.UTF_8)) {
			String r;
			try {
				final BlocLines lines = BlocLines.fromArray(unesc(line).split("\n"));
				final List<String> parts = new ArrayList<>();
				for (Style st : StyleLoader.parseStyleText(lines,
						new net.sourceforge.plantuml.style.AutomaticCounterBasic()))
					parts.add(canonical(st.getQuery()) + " " + styleValues(st));
				r = String.join(" | ", parts);
			} catch (Exception e) {
				r = "ERR " + e.getClass().getSimpleName();
			}
			out.add(line + "\t" + r);
		}
		write("parser.golden", out);
	}

	// ------------------------------------------------------------------
	// Stereotype styles and stereogroups

	static String styleValues(Style style) {
		if (style == null)
			return "null";
		final StringBuilder sb = new StringBuilder();
		for (PName p : PName.values())
			if (style.hasValue(p)) {
				if (sb.length() > 0)
					sb.append(";");
				sb.append(p.name() + "=" + style.value(p).toString());
			}
		return sb.toString();
	}

	static void stereoStyles() throws Exception {
		final List<String> out = new ArrayList<>();
		final StyleBuilder b = StyleLoader.loadSkin("plantuml.skin");
		final BlocLines lines = BlocLines.load(Files.newInputStream(dir.resolve("user_style.css")), null);
		final StyleBuilder user = b.muteStyle(StyleLoader.parseStyleText(lines.subExtract(1, 1), b));
		final StyleQuery part = StyleQuery.of3(SName.root, SName.element, SName.sequenceDiagram,
				SName.participant);
		for (String st : new String[] { "<<foo>>", "<<bar>>", "<<foo>><<bar>>", "<<other>>", "<<(C,red)>>" }) {
			final Stereotype stereo = Stereotype.build(st);
			final StyleQuery sig = part.withStereotype(stereo);
			out.add("with\t" + st + "\t" + styleValues(user.getMergedStyle(sig)));
			out.add("itself\t" + st + "\t" + styleValues(user.getMergedStyle(part.forStereotypeItself(stereo))));
			final List<String> sts = new ArrayList<>();
			for (Style s : stereo.getStyles(user))
				sts.add(styleValues(s));
			out.add("getstyles\t" + st + "\t" + sts);
		}
		out.add("special\t\t" + styleValues(user.getMergedStyleSpecial(part, 5000)));
		for (String g : new String[] { "<<#red>>", "<<##blue>>", "<<###green>>", "<<##[dashed]orange>>",
				"<<#back:yellow;line:navy>>", "<<BackGroundColor:pink;FontColor:teal;>>", "<<foo>> <<#AABBCC>>",
				"<<##[bold]>>" }) {
			final Stereogroup group = Stereogroup.build(g);
			final net.sourceforge.plantuml.klimt.color.Colors colors = group.getInnerColors(HColorSet.instance());
			final Style muted = group.mute(user.getMergedStyle(part), HColorSet.instance());
			out.add("group\t" + g + "\t" + group.getLabels() + " " + hc(colors.getColor(
					net.sourceforge.plantuml.klimt.color.ColorType.BACK)) + "," + hc(colors.getColor(
							net.sourceforge.plantuml.klimt.color.ColorType.LINE)) + "," + hc(colors.getColor(
									net.sourceforge.plantuml.klimt.color.ColorType.TEXT)) + "," + colors.getLineStyle()
					+ " " + hc(muted.value(PName.BackGroundColor).asColor(HColorSet.instance())) + ","
					+ hc(muted.value(PName.LineColor).asColor(HColorSet.instance())) + ","
					+ hc(group.getHColor(part, PName.LineColor, user, HColorSet.instance())));
		}
		write("stereostyles.golden", out);
	}
}
