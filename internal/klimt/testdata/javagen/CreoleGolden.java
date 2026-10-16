// SPDX-License-Identifier: MIT
// Provenance: original
//
// CreoleGolden dumps the golden values of the creole tests (creole_test.go)
// from the upstream jar: for each case of creole_inputs.txt, the dimension
// of the text block and the recorded drawing operations, measured with the
// harness's fixed StringBounder and with the SVG StringBounder of the JVM
// fonts. The second needs the font set of the reference image, which
// internal/klimt/font reproduces, so the file is generated there. From the
// repository root:
//
//   mkdir -p ~/scratch-klimt && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.8.jar \
//     -d ~/scratch-klimt internal/klimt/testdata/javagen/CreoleGolden.java
//   docker run --rm -v "$PWD:/repo" -v "$HOME/scratch-klimt:/classes" plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/classes CreoleGolden /repo/internal/klimt/testdata/golden
//
// It writes creole.golden and creole_skin.golden (skin components: arrow
// configurations, areas, actors, circle interface, textual components).
//
// creole_inputs.txt has a case per line: name, options (size=, align=,
// max=, mode=, nl=1 for Display.getWithNewlines, bordered=1) and the text,
// separated by tabulations; in the text, "⏎" separates the lines of the
// display, "STEREO:" starts a stereotype line and "⇥" stands for a
// tabulation.

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.klimt.LineBreakStrategy;
import net.sourceforge.plantuml.klimt.UBackground;
import net.sourceforge.plantuml.klimt.UChange;
import net.sourceforge.plantuml.klimt.UGroup;
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
import net.sourceforge.plantuml.klimt.creole.CreoleMode;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontFactory;
import net.sourceforge.plantuml.klimt.geom.HorizontalAlignment;
import net.sourceforge.plantuml.klimt.geom.USegment;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.TextBlockUtils;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UImage;
import net.sourceforge.plantuml.klimt.shape.UImageSvg;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.ConfigurationStore;
import net.sourceforge.plantuml.skin.ArrowBody;
import net.sourceforge.plantuml.skin.ArrowConfiguration;
import net.sourceforge.plantuml.skin.ArrowDecoration;
import net.sourceforge.plantuml.skin.ArrowHead;
import net.sourceforge.plantuml.skin.ArrowPart;
import net.sourceforge.plantuml.skin.ComponentType;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.SkinParam;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.style.ClockwiseTopRightBottomLeft;
import net.sourceforge.plantuml.style.SName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.style.StyleBuilder;
import net.sourceforge.plantuml.style.StyleLoader;
import net.sourceforge.plantuml.style.StyleSignatureBasic;
import net.sourceforge.plantuml.klimt.shape.TextBlockTitle;
import net.sourceforge.plantuml.url.Url;

public class CreoleGolden {

	static Path dir;

	public static void main(String[] args) throws Exception {
		dir = Paths.get(args[0]);
		creole();
		skin();
		ubrex();
	}

	// ------------------------------------------------------------------
	// UBrex

	static void ubrex() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String input : Files.readAllLines(dir.resolve("ubrex_inputs.txt"), StandardCharsets.UTF_8)) {
			if (input.isEmpty())
				continue;
			final String[] parts = input.split("\t", -1);
			final com.plantuml.ubrex.UnicodeBracketedExpression ubrex = com.plantuml.ubrex.UnicodeBracketedExpression
					.build(parts[0]);
			final String[] keys = parts[parts.length - 1].isEmpty() ? new String[0]
					: parts[parts.length - 1].split(",");
			for (int i = 1; i < parts.length - 1; i++) {
				final com.plantuml.ubrex.UMatcher m = ubrex.match(parts[i], 0);
				final StringBuilder sb = new StringBuilder();
				sb.append(parts[0]).append("\t").append(parts[i]).append("\t").append(m.startMatch()).append(" ")
						.append(m.exactMatch()).append(" '").append(m.getAcceptedMatch()).append("'");
				for (String key : keys)
					sb.append(" ").append(key).append("=").append(m.findValuesByKey(key));
				if (keys.length > 0)
					sb.append(" prefix=").append(m.findFirstValuesByKeyPrefix(keys[0] + "/")).append(" extract=")
							.append(m.extractByPrefix(keys[0]));
				out.add(sb.toString());
			}
		}
		write("ubrex.golden", out);
	}

	static void write(String name, List<String> lines) throws IOException {
		Files.write(dir.resolve(name), (String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
	}

	static String hc(HColor c) {
		return c == null ? "null" : c.asString();
	}

	static final StringBounder FAKE = new StringBounder() {
		@Override
		public XDimension2D calculateDimension(UFont font, String text) {
			return new XDimension2D(text.length() * font.getSize2D() * 0.6, font.getSize2D() * 1.25);
		}

		@Override
		public FileFormat getFileFormat() {
			return FileFormat.PNG;
		}
	};

	static final StringBounder SVG = FileFormat.SVG.getDefaultStringBounder();

	static String uf(UFont f) {
		return f.getFamily(null, net.sourceforge.plantuml.klimt.font.UFontContext.SVG) + "/" + f.getFontFace() + "/"
				+ f.getSize();
	}

	static String fcString(FontConfiguration fc) {
		final StringBuilder sb = new StringBuilder();
		sb.append(uf(fc.getFont())).append(" ").append(hc(fc.getColor())).append(" [");
		boolean first = true;
		for (FontStyle s : FontStyle.values())
			if (fc.containsStyle(s)) {
				if (!first)
					sb.append(",");
				sb.append(s);
				first = false;
			}
		sb.append("] ext=").append(hc(fc.getExtendedColor())).append(" space=").append(fc.getSpace());
		final Map<String, String> attr = new TreeMap<>(fc.getAttributes());
		if (attr.size() > 0)
			sb.append(" attr=").append(attr);
		return sb.toString();
	}

	// A UGraphic that records the drawing operations.
	static class Rec implements UGraphic {
		final List<String> log;
		final StringBounder bounder;
		final UTranslate t;
		final HColor color;
		final HColor back;
		final UStroke stroke;

		Rec(List<String> log, StringBounder bounder, UTranslate t, HColor color, HColor back, UStroke stroke) {
			this.log = log;
			this.bounder = bounder;
			this.t = t;
			this.color = color;
			this.back = back;
			this.stroke = stroke;
		}

		@Override
		public StringBounder getStringBounder() {
			return bounder;
		}

		@Override
		public UParam getParam() {
			return new UParam() {
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
				return new Rec(log, bounder, ((UTranslate) change).compose(t), color, back, stroke);
			if (change instanceof HColor)
				return new Rec(log, bounder, t, (HColor) change, back, stroke);
			if (change instanceof UBackground)
				return new Rec(log, bounder, t, color, ((UBackground) change).getBackColor(), stroke);
			if (change instanceof UStroke)
				return new Rec(log, bounder, t, color, back, (UStroke) change);
			log.add("apply " + change.getClass().getSimpleName());
			return this;
		}

		@Override
		public ColorMapper getColorMapper() {
			return ColorMapper.IDENTITY;
		}

		@Override
		public void startUrl(Url url) {
			log.add("startUrl " + url.getUrl() + " tooltip=" + url.getTooltip() + " label=" + url.getLabel());
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
			return "text '" + t.getText() + "' " + fcString(t.getFontConfiguration());
		}
		if (shape instanceof UImage) {
			final UImage i = (UImage) shape;
			return "image " + i.getWidth() + "x" + i.getHeight() + " scale=" + i.getScale();
		}
		if (shape instanceof UImageSvg) {
			final UImageSvg i = (UImageSvg) shape;
			return "svg " + i.getWidth() + "x" + i.getHeight();
		}
		if (shape instanceof net.sourceforge.plantuml.klimt.shape.UEmpty) {
			final net.sourceforge.plantuml.klimt.shape.UEmpty e = (net.sourceforge.plantuml.klimt.shape.UEmpty) shape;
			return "empty " + e.getWidth() + "," + e.getHeight();
		}
		if (shape instanceof net.sourceforge.plantuml.klimt.shape.UHorizontalLine)
			return "hline";
		return shape.getClass().getSimpleName();
	}

	static SkinParam newSkinParam() {
		return SkinParam.create(new java.util.HashMap<String, String>(), PathSystem.fetch(), DiagramType.SEQUENCE,
				Pragma.createEmpty(), ConfigurationStore.createEmpty());
	}

	static FontConfiguration fc(int size) {
		return FontConfiguration.create(UFontFactory.sansSerif(size), HColors.BLACK, HColors.BLUE, UStroke.simple());
	}

	static String err(Throwable e) {
		return "ERR " + e.getClass().getSimpleName();
	}

	// ------------------------------------------------------------------
	// Creole

	static void creole() throws Exception {
		final List<String> out = new ArrayList<>();
		final List<String> inputs = Files.readAllLines(dir.resolve("creole_inputs.txt"), StandardCharsets.UTF_8);
		for (String input : inputs) {
			if (input.isEmpty())
				continue;
			final String[] parts = input.split("\t", -1);
			parts[2] = parts[2].replace("⇥", "\t");
			final String name = parts[0];
			int size = 14;
			HorizontalAlignment align = HorizontalAlignment.LEFT;
			String max = null;
			CreoleMode mode = CreoleMode.FULL;
			boolean nl = false;
			boolean bordered = false;
			for (String opt : parts[1].split(",")) {
				if (opt.isEmpty())
					continue;
				final String[] kv = opt.split("=");
				if (kv[0].equals("size"))
					size = Integer.parseInt(kv[1]);
				else if (kv[0].equals("align"))
					align = kv[1].equals("NULL") ? null : HorizontalAlignment.valueOf(kv[1]);
				else if (kv[0].equals("max"))
					max = kv[1];
				else if (kv[0].equals("mode"))
					mode = CreoleMode.valueOf(kv[1]);
				else if (kv[0].equals("nl"))
					nl = true;
				else if (kv[0].equals("bordered"))
					bordered = true;
			}
			Display display;
			if (nl) {
				display = Display.getWithNewlines(Pragma.createEmpty(), parts[2]);
			} else {
				display = Display.empty();
				for (String line : parts[2].split("⏎", -1))
					if (line.startsWith("STEREO:"))
						display = display.add(Stereotype.build(line.substring("STEREO:".length())));
					else
						display = display.add(line);
			}
			for (StringBounder bounder : Arrays.asList(FAKE, SVG)) {
				final String b = bounder == FAKE ? "fake" : "svg";
				final List<String> log = new ArrayList<>();
				try {
					TextBlock tb = display.create0(fc(size), align, newSkinParam(), new LineBreakStrategy(max), mode,
							null, null);
					if (bordered)
						tb = TextBlockUtils.bordered(tb, UStroke.withThickness(1.5), HColors.RED, HColors.MY_YELLOW,
								5, ClockwiseTopRightBottomLeft.topRightBottomLeft(2, 3, 4, 5), "id");
					final XDimension2D dim = tb.calculateDimension(bounder);
					log.add("dim " + dim.getWidth() + "x" + dim.getHeight());
					tb.drawU(new Rec(log, bounder, new UTranslate(0, 0), HColors.BLACK, HColors.none(),
							UStroke.simple()));
				} catch (Throwable e) {
					log.add(err(e));
				}
				for (String l : log)
					out.add(name + "\t" + b + "\t" + l);
			}
		}
		// Style.createTextBlockBordered and TextBlockTitle with the default skin.
		final StyleBuilder builder = StyleLoader.loadSkin("plantuml.skin");
		final Style note = StyleSignatureBasic.of(SName.root, SName.element, SName.sequenceDiagram, SName.note)
				.getMergedStyle(builder);
		final Style title = StyleSignatureBasic.of(SName.root, SName.document, SName.title).getMergedStyle(builder);
		final Display noteText = Display.create("**Note** with //style//", "----", "second line");
		final Display titleText = Display.create("The <color:red>title</color>", "line 2");
		for (StringBounder bounder : Arrays.asList(FAKE, SVG)) {
			final String b = bounder == FAKE ? "fake" : "svg";
			final List<String> log = new ArrayList<>();
			try {
				final TextBlock tb = note.createTextBlockBordered(noteText, HColorSet.instance(), newSkinParam(),
						"nid", LineBreakStrategy.NONE);
				final XDimension2D dim = tb.calculateDimension(bounder);
				log.add("dim " + dim.getWidth() + "x" + dim.getHeight());
				tb.drawU(new Rec(log, bounder, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple()));
			} catch (Throwable e) {
				log.add(err(e));
			}
			for (String l : log)
				out.add("stylebordered\t" + b + "\t" + l);
			log.clear();
			try {
				final java.lang.reflect.Constructor<TextBlockTitle> ctor = TextBlockTitle.class.getDeclaredConstructor(
						FontConfiguration.class, Display.class, net.sourceforge.plantuml.style.ISkinSimple.class);
				ctor.setAccessible(true);
				final TextBlock tb = ctor.newInstance(title.getFontConfiguration(HColorSet.instance()), titleText,
						newSkinParam());
				final XDimension2D dim = tb.calculateDimension(bounder);
				log.add("dim " + dim.getWidth() + "x" + dim.getHeight());
				tb.drawU(new Rec(log, bounder, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple()));
			} catch (Throwable e) {
				log.add(err(e));
			}
			for (String l : log)
				out.add("title\t" + b + "\t" + l);
		}
		write("creole.golden", out);
	}
	// ------------------------------------------------------------------
	// Skin components

	static String arrow(ArrowConfiguration a) {
		return a.name() + " dir=" + a.getArrowDirection() + " self=" + a.isSelfArrow() + " dotted=" + a.isDotted()
				+ " hidden=" + a.isHidden() + " head=" + a.getHead() + " async=" + a.isAsync1() + "," + a.isAsync2()
				+ " part=" + a.getPart() + " incl=" + a.getInclination1() + "," + a.getInclination2() + " rev="
				+ a.isReverseDefine() + " deco=" + a.getDecoration1() + "," + a.getDecoration2() + " d1="
				+ a.getDressing1().getHead() + "/" + a.getDressing1().getPart() + " d2=" + a.getDressing2().getHead()
				+ "/" + a.getDressing2().getPart() + " color=" + hc(a.getColor());
	}

	static class TestComponent extends net.sourceforge.plantuml.skin.AbstractTextualComponent {
		final List<String> log;

		TestComponent(List<String> log, Style style, LineBreakStrategy max, ClockwiseTopRightBottomLeft padding,
				SkinParam skinParam, Display display, boolean enhanced) {
			super(style, max, padding, skinParam, display, enhanced);
			this.log = log;
		}

		TestComponent(List<String> log, Style style, LineBreakStrategy max, ClockwiseTopRightBottomLeft padding,
				SkinParam skinParam, String label) {
			super(style, max, padding, skinParam, label);
			this.log = log;
		}

		@Override
		protected void drawInternalU(UGraphic ug, net.sourceforge.plantuml.skin.Area area) {
			log.add("internal " + area);
			getTextBlock().drawU(ug);
		}

		@Override
		protected void drawBackgroundInternalU(UGraphic ug, net.sourceforge.plantuml.skin.Area area) {
			log.add("background " + area);
			ug.draw(URectangle.build(area.getDimensionToUse()));
		}

		@Override
		public double getPaddingX() {
			return 3;
		}

		@Override
		public double getPreferredWidth(StringBounder stringBounder) {
			return getTextWidth(stringBounder) + 1;
		}

		@Override
		public double getPreferredHeight(StringBounder stringBounder) {
			return getTextHeight(stringBounder) + 2;
		}

		String sizes(StringBounder sb) {
			return "pure=" + getPureTextWidth(sb) + " text=" + getTextWidth(sb) + "x" + getTextHeight(sb) + " pad="
					+ getOldPaddingX1() + "," + getOldPaddingX2() + "," + getOldPaddingY() + " pref="
					+ getPreferredDimension(sb);
		}
	}

	static void skin() throws Exception {
		final List<String> out = new ArrayList<>();
		final List<ArrowConfiguration> arrows = new ArrayList<>();
		final ArrowConfiguration normal = ArrowConfiguration.withDirectionNormal();
		arrows.add(normal);
		arrows.add(ArrowConfiguration.withDirectionBoth());
		arrows.add(ArrowConfiguration.withDirectionSelf(false));
		arrows.add(ArrowConfiguration.withDirectionSelf(true));
		arrows.add(ArrowConfiguration.withDirectionReverse());
		arrows.add(normal.self());
		arrows.add(normal.withBody(ArrowBody.DOTTED));
		arrows.add(normal.withBody(ArrowBody.HIDDEN).withHead(ArrowHead.ASYNC));
		arrows.add(normal.reverse().withHead(ArrowHead.CROSSX));
		arrows.add(normal.withHead1(ArrowHead.ASYNC));
		arrows.add(normal.withHead2(ArrowHead.NONE));
		arrows.add(normal.withHead2(ArrowHead.NONE).withHead1(ArrowHead.NONE));
		arrows.add(normal.withPart(ArrowPart.TOP_PART));
		arrows.add(normal.withHead2(ArrowHead.NONE).withPart(ArrowPart.BOTTOM_PART));
		arrows.add(normal.withDecoration1(ArrowDecoration.CIRCLE).withDecoration2(ArrowDecoration.CIRCLE).reverse());
		arrows.add(normal.withColor(HColors.RED).withThickness(3));
		arrows.add(normal.reverseDefine().reverseDefine().reverseDefine());
		arrows.add(normal.withInclination(4));
		arrows.add(normal.withInclination(4).withHead2(ArrowHead.CROSSX).withHead1(ArrowHead.NORMAL));
		arrows.add(normal.withInclination(-2).reverse().withHead(ArrowHead.ASYNC));
		arrows.add(ArrowConfiguration.withDirectionBoth().withInclination(5).withHead1(ArrowHead.CROSSX));
		for (ArrowConfiguration a : arrows)
			out.add("arrow\t" + arrow(a));

		final StyleBuilder builder = StyleLoader.loadSkin("plantuml.skin");
		final Style arrowStyle = StyleSignatureBasic.of(SName.root, SName.element, SName.sequenceDiagram, SName.arrow)
				.getMergedStyle(builder);
		final Style noteStyle = StyleSignatureBasic.of(SName.root, SName.element, SName.sequenceDiagram, SName.note)
				.getMergedStyle(builder);
		int n = 0;
		for (ArrowConfiguration a : Arrays.asList(normal, normal.withBody(ArrowBody.DOTTED),
				normal.withThickness(2.5), normal.withBody(ArrowBody.DOTTED).withThickness(1.5))) {
			final int idx = n++;
			final List<String> log = new ArrayList<>();
			final UGraphic ug = new Rec(log, FAKE, new UTranslate(0, 0), HColors.BLACK, HColors.none(),
					UStroke.simple());
			a.applyStroke(ug, arrowStyle).draw(new ULine(10, 0));
			a.applyStroke(ug, noteStyle).draw(new ULine(0, 10));
			a.applyStroke(ug).draw(new ULine(1, 1));
			a.applyThicknessOnly(ug).draw(new ULine(2, 2));
			ArrowConfiguration.stroke(ug, 5, 3, 2).draw(new ULine(3, 3));
			for (String l : log)
				out.add("stroke " + idx + "\t" + l);
		}

		final net.sourceforge.plantuml.skin.Area area = net.sourceforge.plantuml.skin.Area.create(30.5, 12);
		out.add("area\t" + area + " " + area.getDimensionToUse() + " " + area.getDeltaX1() + " " + area.getLevel()
				+ " " + area.getLiveDeltaSize() + " " + area.getTextDeltaX());
		area.setDeltaX1(2.5);
		area.setLevel(3);
		area.setLiveDeltaSize(4.25);
		area.setTextDeltaX(-1);
		out.add("area\t" + area + " " + area.getDimensionToUse() + " " + area.getDeltaX1() + " " + area.getLevel()
				+ " " + area.getLiveDeltaSize() + " " + area.getTextDeltaX());

		final List<net.sourceforge.plantuml.klimt.Fashion> fashions = Arrays.asList(
				new net.sourceforge.plantuml.klimt.Fashion(HColors.MY_YELLOW, HColors.MY_RED),
				new net.sourceforge.plantuml.klimt.Fashion(HColors.WHITE, HColors.BLUE).withShadow(3)
						.withStroke(UStroke.withThickness(2)),
				new net.sourceforge.plantuml.klimt.Fashion(null, HColors.BLACK).withStroke(new UStroke(2, 3, 0.5)));
		for (net.sourceforge.plantuml.skin.ActorStyle style : net.sourceforge.plantuml.skin.ActorStyle.values())
			for (int i = 0; i < fashions.size(); i++) {
				final List<String> log = new ArrayList<>();
				final TextBlock tb = style.getTextBlock(fashions.get(i));
				final XDimension2D dim = tb.calculateDimension(FAKE);
				log.add("dim " + dim.getWidth() + "x" + dim.getHeight());
				tb.drawU(new Rec(log, FAKE, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple()));
				for (String l : log)
					out.add("actor " + style + " " + i + "\t" + l);
			}
		for (float thickness : new float[] { 1.5f, 0.3f }) {
			final List<String> log = new ArrayList<>();
			final net.sourceforge.plantuml.skin.CircleInterface ci = new net.sourceforge.plantuml.skin.CircleInterface(
					HColors.MY_YELLOW, HColors.MY_RED, 10.5, thickness);
			log.add("pref " + ci.getPreferredWidth(FAKE) + "x" + ci.getPreferredHeight(FAKE));
			ci.drawU(new Rec(log, FAKE, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple()));
			for (String l : log)
				out.add("circle " + thickness + "\t" + l);
		}

		for (ComponentType t : ComponentType.values()) {
			String v;
			try {
				final Style st = t.getStyleSignature().getMergedStyle(builder);
				v = st.value(net.sourceforge.plantuml.style.PName.BackGroundColor).asString() + " "
						+ st.value(net.sourceforge.plantuml.style.PName.LineColor).asString() + " "
						+ st.value(net.sourceforge.plantuml.style.PName.FontSize).asString() + " "
						+ st.value(net.sourceforge.plantuml.style.PName.LineThickness).asString();
			} catch (Throwable e) {
				v = err(e);
			}
			out.add("componenttype\t" + t + " " + t.isArrow() + " " + v);
		}

		final ClockwiseTopRightBottomLeft padding = ClockwiseTopRightBottomLeft.topRightBottomLeft(2, 7, 3, 5);
		final List<Object[]> components = new ArrayList<>();
		components.add(new Object[] { "label", null, "Some **label**\\nsecond" });
		components.add(new Object[] { "emptylabel", null, "" });
		components.add(new Object[] { "nulllabel", null, null });
		components.add(new Object[] { "display", Display.create("plain", "<b>bold</b>"), Boolean.FALSE });
		components.add(new Object[] { "enhanced", Display.create("head", "--", "after line", "==title==", "end"),
				Boolean.TRUE });
		components.add(new Object[] { "enhancedplain", Display.create("one", "two"), Boolean.TRUE });
		components.add(new Object[] { "stereo",
				Display.create("text").add(Stereotype.build("<<st>>")), Boolean.FALSE });
		for (Object[] c : components) {
			for (StringBounder bounder : Arrays.asList(FAKE, SVG)) {
				final String b = bounder == FAKE ? "fake" : "svg";
				final List<String> log = new ArrayList<>();
				try {
					final TestComponent comp;
					if (c[1] == null)
						comp = new TestComponent(log, noteStyle, LineBreakStrategy.NONE, padding, newSkinParam(),
								(String) c[2]);
					else
						comp = new TestComponent(log, noteStyle, LineBreakStrategy.NONE, padding, newSkinParam(),
								(Display) c[1], (Boolean) c[2]);
					log.add(comp.sizes(bounder));
					final UGraphic ug = new Rec(log, bounder, new UTranslate(0, 0), HColors.BLACK, HColors.none(),
							UStroke.simple());
					comp.drawU(ug, net.sourceforge.plantuml.skin.Area.create(40, 20),
							new net.sourceforge.plantuml.skin.SimpleContext2D(true));
					comp.drawU(ug, net.sourceforge.plantuml.skin.Area.create(40, 20),
							new net.sourceforge.plantuml.skin.SimpleContext2D(false));
				} catch (Throwable e) {
					log.add(err(e));
				}
				for (String l : log)
					out.add("component " + c[0] + "\t" + b + "\t" + l);
			}
		}
		write("creole_skin.golden", out);
	}
}
