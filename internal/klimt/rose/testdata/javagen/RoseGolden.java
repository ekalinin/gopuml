// SPDX-License-Identifier: MIT
// Provenance: original
//
// RoseGolden dumps the golden values of the Rose component tests
// (internal/klimt/rose/golden_test.go) and of the symbol tests
// (internal/decoration/symbol/golden_test.go) from the upstream jar: for
// each component built by Rose with several skin setups (default, skin
// parameters, a style sheet, strict UML), the preferred size, the arrow
// points and the recorded drawing operations in areas of several sizes,
// measured with the harness's fixed StringBounder and with the SVG
// StringBounder of the reference fonts. The second needs the font set of the
// reference image, which internal/klimt/font reproduces, so the files are
// generated there. From the repository root:
//
//   mkdir -p ~/scratch-rose && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.8.jar \
//     -d ~/scratch-rose internal/klimt/rose/testdata/javagen/RoseGolden.java
//   docker run --rm -v "$PWD:/repo" -v "$HOME/scratch-rose:/classes" plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/classes RoseGolden /repo/internal/klimt/rose/testdata/golden \
//     /repo/internal/decoration/symbol/testdata/golden /repo/internal/klimt/testdata/golden
//
// It writes rose.golden.gz, symbol.golden.gz and svek.golden (the svek shapes
// moved to internal/klimt: Opale, Control, Boundary, EntityDomain). The
// style sheets of the "style" and "style2" setups are rose_style.css and
// rose_style2.css, next to rose.golden.gz.

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.EnumSet;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.abel.EntityPortion;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.cucadiagram.BodyFactory;
import net.sourceforge.plantuml.decoration.symbol.USymbols;
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
import net.sourceforge.plantuml.klimt.Fashion;
import net.sourceforge.plantuml.klimt.color.ColorMapper;
import net.sourceforge.plantuml.klimt.color.ColorType;
import net.sourceforge.plantuml.klimt.color.Colors;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColors;
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
import net.sourceforge.plantuml.klimt.shape.UEmpty;
import net.sourceforge.plantuml.klimt.shape.UHorizontalLine;
import net.sourceforge.plantuml.klimt.shape.UImage;
import net.sourceforge.plantuml.klimt.shape.UImageSvg;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.ConfigurationStore;
import net.sourceforge.plantuml.sequencediagram.NotePosition;
import net.sourceforge.plantuml.sequencediagram.Participant;
import net.sourceforge.plantuml.sequencediagram.ParticipantType;
import net.sourceforge.plantuml.skin.ArrowBody;
import net.sourceforge.plantuml.skin.ArrowComponent;
import net.sourceforge.plantuml.skin.ArrowConfiguration;
import net.sourceforge.plantuml.skin.ArrowDecoration;
import net.sourceforge.plantuml.skin.ArrowHead;
import net.sourceforge.plantuml.skin.ArrowPart;
import net.sourceforge.plantuml.skin.Area;
import net.sourceforge.plantuml.skin.ColorParam;
import net.sourceforge.plantuml.skin.Component;
import net.sourceforge.plantuml.skin.ComponentType;
import net.sourceforge.plantuml.skin.LineParam;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.SimpleContext2D;
import net.sourceforge.plantuml.skin.SkinParam;
import net.sourceforge.plantuml.skin.rose.ComponentRoseSelfArrow;
import net.sourceforge.plantuml.skin.rose.Rose;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.style.SName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.style.StyleSignatureBasic;
import net.sourceforge.plantuml.style.parser.StyleParser;
import net.sourceforge.plantuml.url.Url;
import net.sourceforge.plantuml.utils.BlocLines;

public class RoseGolden {

	static Path dir;
	static final Rose ROSE = new Rose();

	public static void main(String[] args) throws Exception {
		dir = Paths.get(args[0]);
		rose();
		symbols(Paths.get(args[1]));
		svek(Paths.get(args[2]));
	}

	static void write(Path d, String name, List<String> lines) throws IOException {
		Files.createDirectories(d);
		Files.write(d.resolve(name), (String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
	}

	static void writeGzip(Path d, String name, List<String> lines) throws IOException {
		Files.createDirectories(d);
		try (OutputStream os = new java.util.zip.GZIPOutputStream(Files.newOutputStream(d.resolve(name)))) {
			os.write((String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
		}
	}

	static String hc(HColor c) {
		return c == null ? "null" : c.asString();
	}

	static String d(double v) {
		return Double.toString(v);
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

		Rec(List<String> log, StringBounder bounder) {
			this(log, bounder, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple());
		}

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
		if (shape instanceof UEmpty) {
			final UEmpty e = (UEmpty) shape;
			return "empty " + e.getWidth() + "," + e.getHeight();
		}
		if (shape instanceof UHorizontalLine)
			return "hline";
		return shape.getClass().getSimpleName();
	}

	static String err(Throwable e) {
		return "ERR " + e.getClass().getSimpleName();
	}

	interface Step {
		void run() throws Exception;
	}

	// Runs a step, recording an exception as the Go side records a panic.
	static void step(List<String> log, Step s) {
		try {
			s.run();
		} catch (Throwable e) {
			log.add(err(e));
		}
	}

	// ------------------------------------------------------------------
	// Skin setups

	static final String[] SETUPS = { "default", "params", "style", "strict", "style2" };

	static SkinParam newSkinParam() {
		return SkinParam.create(new java.util.HashMap<String, String>(), PathSystem.fetch(), DiagramType.SEQUENCE,
				Pragma.createEmpty(), ConfigurationStore.createEmpty());
	}

	static SkinParam setup(String name) throws Exception {
		final SkinParam sp = newSkinParam();
		if (name.equals("params")) {
			sp.setParam("roundcorner", "12");
			sp.setParam("shadowing", "true");
			sp.setParam("defaultFontSize", "15");
			sp.setParam("actorStyle", "awesome");
			sp.setParam("responseMessageBelowArrow", "true");
			sp.setParam("maxMessageSize", "60");
			sp.setParam("noteTextAlignment", "center");
			sp.setParam("SequenceMessagePadding", "3");
			sp.setParam("SequenceMessageMargin", "2");
			sp.setParam("SequenceMessageBorderColor", "red");
			sp.setParam("SequenceMessageBackgroundColor", "#EEEEEE");
			sp.setParam("arrowThickness", "2.5");
			sp.setParam("participantBackgroundColor", "#CCE5FF");
			sp.setParam("noteBackgroundColor", "#FFFFAA");
			sp.setParam("sequenceDividerBackgroundColor", "#F0E0D0");
			sp.setParam("sequenceLifeLineBorderThickness", "2");
			sp.setParam("sequenceGroupBorderThickness", "3");
		} else if (name.equals("style") || name.equals("style2")) {
			final String file = name.equals("style") ? "rose_style.css" : "rose_style2.css";
			final BlocLines lines = BlocLines.load(Files.newInputStream(dir.resolve(file)), null);
			sp.muteStyle(new StyleParser(sp.getCurrentStyleBuilder()).parse(lines.subExtract(1, 1)));
			sp.applyPendingStyleMigration();
		} else if (name.equals("strict")) {
			sp.setParam("style", "strictuml");
			sp.setParam("actorStyle", "hollow");
			sp.setParam("noteTextAlignment", "right");
			sp.setParam("SequenceMessageAlignment", "right");
		}
		return sp;
	}

	static Style style(SkinParam sp, SName... names) {
		return StyleSignatureBasic.of(names).getMergedStyle(sp.getCurrentStyleBuilder());
	}

	static Style seq(SkinParam sp, SName... names) {
		final SName[] all = new SName[names.length + 3];
		all[0] = SName.root;
		all[1] = SName.element;
		all[2] = SName.sequenceDiagram;
		System.arraycopy(names, 0, all, 3, names.length);
		return style(sp, all);
	}

	// ------------------------------------------------------------------
	// Recording of a component

	static String name(Component c) {
		return c.getClass().getSimpleName();
	}

	// Records the class, the preferred size, the arrow points and the
	// drawings in the areas (each a width, a height, a deltaX1, a level, a
	// live delta size and a text delta).
	static void record(List<String> out, String caseName, Component comp, StringBounder b, double[][] areas) {
		final String bn = b == FAKE ? "fake" : "svg";
		final List<String> log = new ArrayList<>();
		log.add("class " + name(comp));
		step(log, () -> {
			final XDimension2D dim = comp.getPreferredDimension(b);
			log.add("pref " + d(dim.getWidth()) + "x" + d(dim.getHeight()));
		});
		if (comp instanceof ArrowComponent) {
			final ArrowComponent a = (ArrowComponent) comp;
			final XDimension2D full = new XDimension2D(120, 40);
			step(log, () -> {
				final XPoint2D p = a.getStartPoint(b, full);
				log.add("start " + d(p.getX()) + "," + d(p.getY()));
			});
			step(log, () -> {
				final XPoint2D p = a.getEndPoint(b, full);
				log.add("end " + d(p.getX()) + "," + d(p.getY()));
			});
			step(log, () -> log.add("ypoint " + d(a.getYPoint(b))));
			step(log, () -> log.add("paddingy " + d(a.getPaddingY())));
			step(log, () -> log.add("posarrow " + d(a.getPosArrow(b))));
			if (comp instanceof ComponentRoseSelfArrow)
				step(log, () -> log.add("drawn " + d(((ComponentRoseSelfArrow) comp).getDrawnWidth(b))));
		}
		for (double[] ar : areas) {
			step(log, () -> {
				final XDimension2D pref = comp.getPreferredDimension(b);
				final Area area = Area.create(pref.getWidth() + ar[0], pref.getHeight() + ar[1]);
				area.setDeltaX1(ar[2]);
				area.setLevel((int) ar[3]);
				area.setLiveDeltaSize(ar[4]);
				area.setTextDeltaX(ar[5]);
				log.add("area " + area + " lvl=" + area.getLevel() + " live=" + d(area.getLiveDeltaSize()) + " tdx="
						+ d(area.getTextDeltaX()));
				final UGraphic ug = new Rec(log, b);
				log.add("background");
				step(log, () -> comp.drawU(ug, area, new SimpleContext2D(true)));
				log.add("foreground");
				step(log, () -> comp.drawU(ug, area, new SimpleContext2D(false)));
			});
		}
		for (String l : log)
			out.add(caseName + "\t" + bn + "\t" + l);
	}

	interface Factory {
		Component create() throws Exception;
	}

	static void comp(List<String> out, String caseName, Factory f, double[][] areas) {
		for (StringBounder b : Arrays.asList(FAKE, SVG)) {
			final Component comp;
			try {
				comp = f.create();
			} catch (Throwable e) {
				out.add(caseName + "\t" + (b == FAKE ? "fake" : "svg") + "\t" + err(e));
				continue;
			}
			record(out, caseName, comp, b, areas);
		}
	}

	static final double[][] AREA_PREF = { { 0, 0, 0, 0, 0, 0 } };
	static final double[][] AREA_WIDE = { { 0, 0, 0, 0, 0, 0 }, { 30, 12, 0, 0, 0, 0 } };
	static final double[][] AREA_NOTE = { { 0, 0, 0, 0, 0, 0 }, { 40, 6, 0, 0, 0, 0 }, { -1, 0, 0, 0, 0, 0 } };
	static final double[][] AREA_ARROW = { { 0, 0, 0, 0, 0, 0 }, { 40, 0, 0, 0, 0, 0 }, { 40, 0, 0, 0, 0, 9 },
			{ 40, 0, 0, 0, 0, -9 } };
	static final double[][] AREA_SELF = { { 0, 0, 0, 0, 0, 0 }, { 20, 0, -5, 0, 5, 0 }, { 20, 0, 5, 1, 5, 0 },
			{ 20, 0, 5, 2, 5, 0 }, { 20, 0, -5, 2, 5, 0 }, { 20, 0, 0, 3, 5, 0 } };
	static final double[][] AREA_LINE = { { 0, 0, 0, 0, 0, 0 }, { 20, 30, 0, 0, 0, 0 }, { 9, -20, 0, 0, 0, 0 } };

	// ------------------------------------------------------------------
	// Displays

	static Display display(String spec) {
		switch (spec) {
		case "simple":
			return Display.create("Alice");
		case "multi":
			return Display.getWithNewlines(Pragma.createEmpty(), "Long participant\\nsecond line");
		case "creole":
			return Display.create("**Bold** and <color:red>red</color>", "----", "<i>after</i> rule");
		case "empty":
			return Display.create("");
		case "hello":
			return Display.create("hello");
		case "message":
			return Display.getWithNewlines(Pragma.createEmpty(),
					"a much longer message text\\nwith <b>two</b> lines and some more words to wrap");
		case "note":
			return Display.create("note text");
		case "notesep":
			return Display.create("first", "----", "after line", "==title==", "end");
		case "notecreole":
			return Display.create("* bullet one", "* bullet **two**", "<u>under</u> text");
		case "notelong":
			return Display.create("This is a rather long note text that should wrap when a maximum width is set");
		}
		throw new IllegalArgumentException(spec);
	}

	// ------------------------------------------------------------------
	// Rose

	static void rose() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String setupName : SETUPS) {
			participants(out, setupName);
			arrows(out, setupName);
			notes(out, setupName);
			others(out, setupName);
		}
		alignments(out);
		misc(out);
		writeGzip(dir, "rose.golden.gz", out);
	}

	static final ParticipantType[] PTYPES = ParticipantType.values();

	static ComponentType headType(ParticipantType t, boolean head) {
		switch (t) {
		case PARTICIPANT:
			return head ? ComponentType.PARTICIPANT_HEAD : ComponentType.PARTICIPANT_TAIL;
		case ACTOR:
			return head ? ComponentType.ACTOR_HEAD : ComponentType.ACTOR_TAIL;
		case BOUNDARY:
			return head ? ComponentType.BOUNDARY_HEAD : ComponentType.BOUNDARY_TAIL;
		case CONTROL:
			return head ? ComponentType.CONTROL_HEAD : ComponentType.CONTROL_TAIL;
		case ENTITY:
			return head ? ComponentType.ENTITY_HEAD : ComponentType.ENTITY_TAIL;
		case QUEUE:
			return head ? ComponentType.QUEUE_HEAD : ComponentType.QUEUE_TAIL;
		case DATABASE:
			return head ? ComponentType.DATABASE_HEAD : ComponentType.DATABASE_TAIL;
		case COLLECTIONS:
			return head ? ComponentType.COLLECTIONS_HEAD : ComponentType.COLLECTIONS_TAIL;
		}
		throw new IllegalArgumentException();
	}

	static void participants(List<String> out, String setupName) throws Exception {
		final String[][] variants = { { "simple", null, null }, { "multi", "<<foo>>", null },
				{ "creole", null, "#FFAAAA" }, { "simple", "<<bar>> <<foo>>", "#AAFFAA" } };
		for (ParticipantType t : PTYPES)
			for (boolean head : new boolean[] { true, false })
				for (int v = 0; v < variants.length; v++) {
					final String[] variant = variants[v];
					final String caseName = setupName + "/participant/" + headType(t, head) + "/" + v;
					comp(out, caseName, () -> {
						final SkinParam sp = setup(setupName);
						final Participant p = new Participant(null, t, "A", display(variant[0]),
								EnumSet.noneOf(EntityPortion.class), 1, sp.getCurrentStyleBuilder(), "part1");
						if (variant[1] != null)
							p.setStereotype(Stereotype.build(variant[1]), false);
						if (variant[2] != null)
							p.setSpecificColorTOBEREMOVED(ColorType.BACK, HColorSet.instance().getColor(variant[2]));
						return ROSE.createComponentParticipant(p, headType(t, head), null, sp, p.getDisplay(false));
					}, AREA_WIDE);
				}
	}

	static List<ArrowConfiguration> arrowConfigs() {
		final ArrowConfiguration n = ArrowConfiguration.withDirectionNormal();
		final ArrowConfiguration b = ArrowConfiguration.withDirectionBoth();
		final List<ArrowConfiguration> result = new ArrayList<>();
		result.add(n);
		result.add(n.withBody(ArrowBody.DOTTED));
		result.add(n.withBody(ArrowBody.HIDDEN));
		result.add(n.reverse());
		result.add(b);
		result.add(b.withBody(ArrowBody.DOTTED));
		result.add(n.withHead2(ArrowHead.ASYNC));
		result.add(n.withHead1(ArrowHead.ASYNC).withHead2(ArrowHead.ASYNC));
		result.add(b.withHead1(ArrowHead.ASYNC));
		result.add(n.withPart(ArrowPart.TOP_PART));
		result.add(n.withPart(ArrowPart.BOTTOM_PART));
		result.add(b.withPart(ArrowPart.TOP_PART));
		result.add(n.withHead2(ArrowHead.ASYNC).withPart(ArrowPart.BOTTOM_PART));
		result.add(n.withHead2(ArrowHead.CROSSX));
		result.add(n.withHead1(ArrowHead.CROSSX));
		result.add(n.withDecoration2(ArrowDecoration.CIRCLE));
		result.add(n.withDecoration1(ArrowDecoration.CIRCLE));
		result.add(n.withDecoration1(ArrowDecoration.CIRCLE).withDecoration2(ArrowDecoration.CIRCLE));
		result.add(b.withDecoration1(ArrowDecoration.CIRCLE).withDecoration2(ArrowDecoration.CIRCLE));
		result.add(n.withHead1(ArrowHead.ASYNC).withDecoration1(ArrowDecoration.CIRCLE));
		result.add(n.withHead2(ArrowHead.CROSSX).withDecoration2(ArrowDecoration.CIRCLE));
		result.add(n.withHead2(ArrowHead.NONE).withDecoration2(ArrowDecoration.CIRCLE));
		result.add(n.withHead1(ArrowHead.CROSSX).withDecoration1(ArrowDecoration.CIRCLE));
		result.add(n.reverseDefine());
		result.add(n.reverseDefine().withBody(ArrowBody.DOTTED).reverse());
		result.add(n.withInclination(3));
		result.add(n.withInclination(3).reverse());
		result.add(b.withInclination(4).withHead1(ArrowHead.ASYNC).withHead2(ArrowHead.ASYNC));
		result.add(n.withColor(HColors.RED).withThickness(3));
		result.add(n.self());
		result.add(n.self().reverseDefine());
		result.add(ArrowConfiguration.withDirectionSelf(false).withBody(ArrowBody.DOTTED));
		result.add(ArrowConfiguration.withDirectionSelf(true));
		result.add(n.self().withHead2(ArrowHead.ASYNC));
		result.add(n.self().withHead1(ArrowHead.ASYNC).withHead2(ArrowHead.ASYNC).withPart(ArrowPart.TOP_PART));
		result.add(n.self().withHead1(ArrowHead.ASYNC).withPart(ArrowPart.BOTTOM_PART));
		result.add(n.self().withHead2(ArrowHead.CROSSX));
		result.add(n.self().withHead1(ArrowHead.CROSSX));
		result.add(n.self().withHead1(ArrowHead.NORMAL));
		result.add(n.self().withPart(ArrowPart.TOP_PART));
		result.add(n.self().withPart(ArrowPart.BOTTOM_PART).reverseDefine());
		result.add(n.self().withDecoration1(ArrowDecoration.CIRCLE).withDecoration2(ArrowDecoration.CIRCLE));
		result.add(n.self().withDecoration1(ArrowDecoration.CIRCLE).withHead1(ArrowHead.NORMAL).reverseDefine());
		result.add(n.self().withDecoration2(ArrowDecoration.CIRCLE).reverseDefine());
		result.add(n.self().withHead1(ArrowHead.CROSSX).withHead2(ArrowHead.CROSSX).reverseDefine());
		result.add(n.self().withHead2(ArrowHead.ASYNC).reverseDefine());
		result.add(n.self().withBody(ArrowBody.HIDDEN));
		result.add(n.withPart(ArrowPart.TOP_PART).reverse());
		result.add(n.withPart(ArrowPart.BOTTOM_PART).reverse());
		result.add(n.self().withDecoration1(ArrowDecoration.CIRCLE).withHead1(ArrowHead.ASYNC));
		result.add(n.self().withDecoration1(ArrowDecoration.CIRCLE).reverseDefine());
		return result;
	}

	static void arrows(List<String> out, String setupName) throws Exception {
		final List<ArrowConfiguration> configs = arrowConfigs();
		for (int i = 0; i < configs.size(); i++) {
			final ArrowConfiguration config = configs.get(i);
			final String[] labels = i < 2 ? new String[] { "hello", "empty", "message", "creole" }
					: new String[] { "hello" };
			for (String label : labels) {
				final String caseName = setupName + "/arrow/" + i + "/" + label;
				comp(out, caseName, () -> {
					final SkinParam sp = setup(setupName);
					return ROSE.createComponentArrow(new Style[] { seq(sp, SName.arrow) }, config, sp, display(label));
				}, config.isSelfArrow() ? AREA_SELF : AREA_ARROW);
			}
			if (configs.get(i).isSelfArrow() && i < 32)
				for (String label : new String[] { "empty", "message" }) {
					final String caseName = setupName + "/arrow/" + i + "/" + label;
					comp(out, caseName, () -> {
						final SkinParam sp = setup(setupName);
						return ROSE.createComponentArrow(new Style[] { seq(sp, SName.arrow) }, config, sp,
								display(label));
					}, AREA_SELF);
				}
		}
	}

	static void notes(List<String> out, String setupName) throws Exception {
		final ComponentType[] types = { ComponentType.NOTE, ComponentType.NOTE_HEXAGONAL, ComponentType.NOTE_BOX };
		final String[] positions = { "null", "LEFT", "OVER_SEVERAL" };
		final String[] texts = { "note", "notesep", "notecreole", "notelong" };
		for (ComponentType t : types)
			for (String pos : positions)
				for (String text : texts)
					for (boolean colored : new boolean[] { false, true }) {
						if (colored && !text.equals("note"))
							continue;
						final String caseName = setupName + "/note/" + t + "/" + pos + "/" + text
								+ (colored ? "/colored" : "");
						comp(out, caseName, () -> {
							final SkinParam sp = setup(setupName);
							final Style st;
							if (t == ComponentType.NOTE_HEXAGONAL)
								st = seq(sp, SName.note, SName.hnote);
							else if (t == ComponentType.NOTE_BOX)
								st = seq(sp, SName.note, SName.rnote);
							else
								st = seq(sp, SName.note);
							final Colors colors = colored
									? Colors.empty().add(ColorType.BACK, HColorSet.instance().getColor("#AAFFAA"))
											.add(ColorType.LINE, HColors.BLUE)
									: Colors.empty();
							if (pos.equals("null"))
								return ROSE.createComponentNote(new Style[] { st }, t, sp, display(text), colors);
							return ROSE.createComponentNote(new Style[] { st }, t, sp, display(text), colors,
									NotePosition.valueOf(pos));
						}, AREA_NOTE);
					}
	}

	static void others(List<String> out, String setupName) throws Exception {
		// Grouping headers and else separators.
		final String[][] headers = { { "alt", "x > 5" }, { "loop" }, { "group", "multi\\nline comment" },
				{ "critical", "" } };
		for (boolean teoz : new boolean[] { false, true }) {
			final ComponentType ht = teoz ? ComponentType.GROUPING_HEADER_TEOZ : ComponentType.GROUPING_HEADER_LEGACY;
			for (int i = 0; i < headers.length; i++) {
				final String[] h = headers[i];
				comp(out, setupName + "/" + ht + "/" + i, () -> {
					final SkinParam sp = setup(setupName);
					return ROSE.createComponent(new Style[] { seq(sp, SName.group), seq(sp, SName.groupHeader) }, ht,
							null, sp, Display.create(h));
				}, AREA_WIDE);
			}
			final ComponentType et = teoz ? ComponentType.GROUPING_ELSE_TEOZ : ComponentType.GROUPING_ELSE_LEGACY;
			for (String comment : new String[] { "x <= 5", "", "<b>bold</b> case" })
				comp(out, setupName + "/" + et + "/" + comment, () -> {
					final SkinParam sp = setup(setupName);
					return ROSE.createComponent(new Style[] { seq(sp, SName.group) }, et, null, sp,
							Display.create(comment));
				}, AREA_WIDE);
		}
		comp(out, setupName + "/GROUPING_SPACE", () -> ROSE.createComponent(null, ComponentType.GROUPING_SPACE, null,
				setup(setupName), null), AREA_WIDE);

		// Activation boxes.
		for (ComponentType t : new ComponentType[] { ComponentType.ACTIVATION_BOX_CLOSE_CLOSE,
				ComponentType.ACTIVATION_BOX_CLOSE_OPEN, ComponentType.ACTIVATION_BOX_OPEN_CLOSE,
				ComponentType.ACTIVATION_BOX_OPEN_OPEN })
			for (boolean withDisplay : new boolean[] { false, true })
				comp(out, setupName + "/" + t + "/" + withDisplay, () -> {
					final SkinParam sp = setup(setupName);
					return ROSE.createComponent(new Style[] { seq(sp, SName.activationBox) }, t, null, sp,
							withDisplay ? Display.create("Alice") : null);
				}, new double[][] { { 0, 0, 0, 0, 0, 0 }, { 0, 40, 0, 0, 0, 0 }, { 15, 33, 0, 0, 0, 0 } });

		// Lifelines, delays, destroy, page break.
		for (boolean withDisplay : new boolean[] { false, true })
			comp(out, setupName + "/PARTICIPANT_LINE/" + withDisplay, () -> {
				final SkinParam sp = setup(setupName);
				return ROSE.createComponent(new Style[] { seq(sp, SName.lifeLine) }, ComponentType.PARTICIPANT_LINE,
						null, sp, withDisplay ? Display.create("Alice") : null);
			}, new double[][] { { 0, 0, 0, 0, 0, 0 }, { 20, 30, 0, 0, 0, 0 }, { 9, -20, 0, 0, 0, 0 } });
		comp(out, setupName + "/DELAY_LINE", () -> {
			final SkinParam sp = setup(setupName);
			return ROSE.createComponent(new Style[] { seq(sp, SName.lifeLine, SName.delay) }, ComponentType.DELAY_LINE,
					null, sp, null);
		}, AREA_LINE);
		for (String text : new String[] { "5 minutes later", "two\\nlines" })
			comp(out, setupName + "/DELAY_TEXT/" + text, () -> {
				final SkinParam sp = setup(setupName);
				return ROSE.createComponent(new Style[] { seq(sp, SName.lifeLine, SName.delay) },
						ComponentType.DELAY_TEXT, null, sp, Display.getWithNewlines(Pragma.createEmpty(), text));
			}, AREA_WIDE);
		comp(out, setupName + "/DESTROY", () -> {
			final SkinParam sp = setup(setupName);
			return ROSE.createComponent(
					new Style[] {
							ComponentType.DESTROY.getStyleSignature().getMergedStyle(sp.getCurrentStyleBuilder()) },
					ComponentType.DESTROY, null, sp, null);
		}, AREA_WIDE);
		comp(out, setupName + "/NEWPAGE", () -> {
			final SkinParam sp = setup(setupName);
			return ROSE.createComponentNewPage(new Style[] { seq(sp, SName.newpage) }, sp);
		}, AREA_WIDE);

		// Dividers, references, boxes.
		for (String text : new String[] { "", "Initialization", "two\\nlines" })
			comp(out, setupName + "/DIVIDER/" + text, () -> {
				final SkinParam sp = setup(setupName);
				return ROSE.createComponent(new Style[] { seq(sp, SName.separator) }, ComponentType.DIVIDER, null, sp,
						Display.getWithNewlines(Pragma.createEmpty(), text));
			}, AREA_WIDE);
		final String[][] refs = { { "ref", "Init" }, { "ref over A, B", "first line", "<b>second</b> line" },
				{ "ref" } };
		for (int i = 0; i < refs.length; i++) {
			final String[] r = refs[i];
			comp(out, setupName + "/REFERENCE/" + i, () -> {
				final SkinParam sp = setup(setupName);
				return ROSE.createComponent(new Style[] { seq(sp, SName.reference), seq(sp, SName.referenceHeader) },
						ComponentType.REFERENCE, null, sp, Display.create(r));
			}, AREA_WIDE);
		}
		for (String text : new String[] { "Box", "Multi\\nline box", "" })
			comp(out, setupName + "/ENGLOBER/" + text, () -> {
				final SkinParam sp = setup(setupName);
				return ROSE.createComponent(new Style[] { seq(sp, SName.box) }, ComponentType.ENGLOBER, null, sp,
						Display.getWithNewlines(Pragma.createEmpty(), text));
			}, AREA_WIDE);
	}

	// The message alignments of the arrow style.
	static void alignments(List<String> out) throws Exception {
		final ArrowConfiguration n = ArrowConfiguration.withDirectionNormal();
		final List<ArrowConfiguration> configs = Arrays.asList(n, n.reverse(), ArrowConfiguration.withDirectionBoth(),
				n.reverseDefine(), n.reverse().reverseDefine());
		for (String align : new String[] { "first", "direction", "reversedirection", "center", "right", "left" })
			for (int i = 0; i < configs.size(); i++) {
				final ArrowConfiguration config = configs.get(i);
				comp(out, "align/" + align + "/" + i, () -> {
					final SkinParam sp = newSkinParam();
					sp.setParam("SequenceMessageAlignment", align);
					return ROSE.createComponentArrow(new Style[] { seq(sp, SName.arrow) }, config, sp,
							display("hello"));
				}, AREA_ARROW);
			}
	}

	// Error cases and the static helpers.
	static void misc(List<String> out) throws Exception {
		final List<String> log = new ArrayList<>();
		final SkinParam sp = newSkinParam();
		final Style[] arrow = new Style[] { seq(sp, SName.arrow) };
		for (ComponentType t : ComponentType.values()) {
			log.add("nostyle " + t);
			step(log, () -> log.add(name(ROSE.createComponent(null, t, null, sp, Display.create("x")))));
			log.add("styles " + t);
			step(log, () -> log.add(name(
					ROSE.createComponent(new Style[] { seq(sp, SName.group), seq(sp, SName.groupHeader) }, t,
							ArrowConfiguration.withDirectionNormal(), sp, Display.create("x", "y")))));
			log.add("note " + t);
			step(log, () -> log.add(name(ROSE.createComponentNote(arrow, t, sp, Display.create("x"),
					Colors.empty()))));
		}
		for (ComponentType t : ComponentType.values()) {
			log.add("participant " + t);
			step(log, () -> {
				final Participant p = new Participant(null, ParticipantType.PARTICIPANT, "A", Display.create("A"),
						EnumSet.noneOf(EntityPortion.class), 1, sp.getCurrentStyleBuilder(), "part1");
				log.add(name(ROSE.createComponentParticipant(p, t, null, sp, Display.create("A"))));
			});
		}
		step(log, () -> log.add("participant line " + name(ROSE.createComponent(arrow,
				ComponentType.PARTICIPANT_LINE, null, sp, null))));
		final SkinParam sp2 = newSkinParam();
		sp2.setParam("sequenceLifeLineBorderThickness", "2.5");
		for (SkinParam p : Arrays.asList(sp, sp2))
			for (LineParam lp : new LineParam[] { LineParam.sequenceLifeLineBorder, LineParam.sequenceArrow }) {
				log.add("stroke " + lp);
				step(log, () -> log.add(String.valueOf(Rose.getStroke(p, lp, 1.5))));
			}
		step(log, () -> log.add("color " + hc(ROSE.getHtmlColor(sp, ColorParam.background))));
		step(log, () -> log.add("color " + hc(ROSE.getHtmlColor(sp, ColorParam.sequenceLifeLineBorder))));
		step(log, () -> log.add("color " + hc(ROSE.getHtmlColor(sp2, null, ColorParam.noteBackground, ColorParam.arrowHead))));
		step(log, () -> log.add("color " + hc(ROSE.getHtmlColor(sp2, (Stereotype) null))));
		for (String pos : new String[] { "left", "Over", "over_several", "bottom", null, "middle" }) {
			log.add("noteposition " + pos);
			step(log, () -> log.add(String.valueOf(NotePosition.defaultLeft(pos))));
		}
		for (String l : log)
			out.add("misc\t" + l);
	}

	// ------------------------------------------------------------------
	// Symbols

	static final String[] SYMBOL_NAMES = { "ACTION", "ACTOR_AWESOME", "ACTOR_HOLLOW", "ACTOR_STICKMAN",
			"ACTOR_STICKMAN_BUSINESS", "AGENT", "ARCHIMATE", "ARTIFACT", "BOUNDARY", "CARD", "CLOUD", "COLLECTIONS",
			"COMPONENT_RECTANGLE", "COMPONENT1", "COMPONENT2", "CONTROL", "DATABASE", "ENTITY_DOMAIN", "FILE",
			"FOLDER", "FRAME", "GROUP", "HEXAGON", "INTERFACE", "LABEL", "NODE", "PACKAGE", "PARTITION", "PERSON",
			"PROCESS", "QUEUE", "RECTANGLE", "STACK", "STORAGE", "USECASE", "USECASE_BUSINESS" };

	static net.sourceforge.plantuml.decoration.symbol.USymbol symbolOf(String n) throws Exception {
		return (net.sourceforge.plantuml.decoration.symbol.USymbol) USymbols.class.getField(n).get(null);
	}

	static void symbols(Path out) throws Exception {
		final List<String> lines = new ArrayList<>();
		final SkinParam sp = newSkinParam();
		final FontConfiguration fc = FontConfiguration.create(UFontFactory.sansSerif(13), HColors.BLACK,
				HColors.BLUE, UStroke.simple());
		final FontConfiguration fcs = FontConfiguration.create(UFontFactory.sansSerif(10), HColors.RED, HColors.BLUE,
				UStroke.simple());
		final List<Fashion> fashions = Arrays.asList(new Fashion(HColors.MY_YELLOW, HColors.MY_RED),
				new Fashion(HColors.WHITE, HColors.BLUE).withShadow(3).withStroke(UStroke.withThickness(2)),
				new Fashion(HColors.MY_YELLOW, HColors.MY_RED).withCorner(10, 0),
				new Fashion(HColors.WHITE, HColors.BLUE).withShadow(3).withCorner(0, 8));
		final String[][] texts = { { "label" }, { "first", "--", "second", "==", "third" },
				{ "a", "--title--", "b", "..", "c" }, { "first", "----", "second" } };
		final Style noteStyle = seq(sp, SName.note);
		final String[] stereos = { "", "<<stereo>>" };
		final double[][] magnetic = { { 5, 5 }, { 60, 3 }, { 60, -2 }, { 30, -1 }, { 120, 10 }, { 20, 0 } };
		for (String n : SYMBOL_NAMES) {
			final net.sourceforge.plantuml.decoration.symbol.USymbol symbol = symbolOf(n);
			lines.add(n + "\tsnames " + Arrays.toString(symbol.getSNames()) + " supp="
					+ symbol.suppHeightBecauseOfShape() + "," + symbol.suppWidthBecauseOfShape());
			for (int ti = 0; ti < texts.length; ti++)
				for (int si = 0; si < stereos.length; si++)
					for (int fi = 0; fi < fashions.size(); fi++)
						for (StringBounder b : Arrays.asList(FAKE, SVG)) {
							final String bn = b == FAKE ? "fake" : "svg";
							final TextBlock label = ti == 1 || ti == 2
									? BodyFactory.create3(Display.create(texts[ti]), sp, HorizontalAlignment.LEFT, fc,
											LineBreakStrategy.NONE, noteStyle)
									: Display.create(texts[ti]).create(fc, HorizontalAlignment.LEFT, sp);
							final TextBlock stereo = stereos[si].isEmpty() ? TextBlockUtils.empty(0, 0)
									: Display.create(stereos[si]).create(fcs, HorizontalAlignment.CENTER, sp);
							final TextBlock name = Display.create("pkg").create(fcs, HorizontalAlignment.LEFT, sp);
							final Fashion fashion = fashions.get(fi);
							final boolean more = fi == 0 && ti < 2;
							final boolean withMagnetic = fi == 0 && ti == 0 && si == 1;
							for (String kind : new String[] { "small", "big", "smallr", "bigr", "bigl", "smalln" }) {
								if (kind.length() > 5 && !more)
									continue;
								if (kind.equals("bigl") && (ti != 0 || fi != 0))
									continue;
								final List<String> log = new ArrayList<>();
								step(log, () -> {
									final TextBlock tb;
									if (kind.equals("small"))
										tb = symbol.asSmall(name, label, stereo, fashion, HorizontalAlignment.CENTER);
									else if (kind.equals("smallr"))
										tb = symbol.asSmall(name, label, stereo, fashion, HorizontalAlignment.RIGHT);
									else if (kind.equals("smalln"))
										tb = symbol.asSmall(null, label, stereo, fashion, HorizontalAlignment.LEFT);
									else if (kind.equals("big"))
										tb = symbol.asBig(label, HorizontalAlignment.LEFT, stereo, 150, 90, fashion,
												HorizontalAlignment.CENTER);
									else if (kind.equals("bigr"))
										tb = symbol.asBig(label, HorizontalAlignment.RIGHT, stereo, 150, 90, fashion,
												HorizontalAlignment.RIGHT);
									else
										tb = symbol.asBig(label, HorizontalAlignment.CENTER, stereo, 220, 140,
												fashion, HorizontalAlignment.CENTER);
									final XDimension2D dim = tb.calculateDimension(b);
									log.add("dim " + d(dim.getWidth()) + "x" + d(dim.getHeight()));
									tb.drawU(new Rec(log, b).apply(new UTranslate(3, 4)));
									if (withMagnetic)
										for (double[] pt : magnetic) {
											final UTranslate force = tb.getMagneticBorder().getForceAt(b,
													new XPoint2D(pt[0], pt[1]));
											log.add("mb " + d(pt[0]) + "," + d(pt[1]) + " -> " + d(force.getDx()) + ","
													+ d(force.getDy()));
										}
								});
								for (String l : log)
									lines.add(n + "/" + kind + "/" + ti + "/" + si + "/" + fi + "\t" + bn + "\t" + l);
							}
						}
		}
		writeGzip(out, "symbol.golden.gz", lines);
	}

	// ------------------------------------------------------------------
	// svek shapes moved to internal/klimt

	static void svek(Path out) throws Exception {
		final List<String> lines = new ArrayList<>();
		final SkinParam sp = newSkinParam();
		final FontConfiguration fc = FontConfiguration.create(UFontFactory.sansSerif(13), HColors.BLACK,
				HColors.BLUE, UStroke.simple());
		final List<Fashion> fashions = Arrays.asList(new Fashion(HColors.MY_YELLOW, HColors.MY_RED),
				new Fashion(HColors.WHITE, HColors.BLUE).withShadow(3).withStroke(UStroke.withThickness(2)));
		final net.sourceforge.plantuml.utils.Direction[] dirs = { null, net.sourceforge.plantuml.utils.Direction.LEFT,
				net.sourceforge.plantuml.utils.Direction.RIGHT, net.sourceforge.plantuml.utils.Direction.UP,
				net.sourceforge.plantuml.utils.Direction.DOWN };
		final double[][] points = { { 10, 12, -25, 8 }, { 200, -40, 90, 120 }, { -30, 300, 40, -20 } };
		for (StringBounder b : Arrays.asList(FAKE, SVG)) {
			final String bn = b == FAKE ? "fake" : "svg";
			final TextBlock text = Display.create("note", "second <b>line</b>").create(fc, HorizontalAlignment.LEFT,
					sp);
			for (double shadow : new double[] { 0, 3 })
				for (double rc : new double[] { 0, 8 })
					for (boolean withStroke : new boolean[] { false, true })
						for (boolean withLink : new boolean[] { false, true })
							for (int di = 0; di < dirs.length; di++)
								for (int pi = 0; pi < points.length; pi++) {
									if ((!withLink || di == 0) && pi > 0)
										continue;
									final net.sourceforge.plantuml.utils.Direction dir = dirs[di];
									final double[] pt = points[pi];
									final List<String> log = new ArrayList<>();
									step(log, () -> {
										final net.sourceforge.plantuml.svek.image.Opale o = new net.sourceforge.plantuml.svek.image.Opale(
												shadow, HColors.MY_RED, HColors.MY_YELLOW, text, withLink,
												withStroke ? UStroke.withThickness(2) : null);
										o.setRoundCorner(rc);
										if (dir != null)
											o.setOpale(dir, new XPoint2D(pt[0], pt[1]), new XPoint2D(pt[2], pt[3]));
										final XDimension2D dim = o.calculateDimension(b);
										log.add("dim " + d(dim.getWidth()) + "x" + d(dim.getHeight()) + " m="
												+ o.getMarginX1());
										o.drawU(new Rec(log, b));
									});
									for (String l : log)
										lines.add("opale/" + shadow + "/" + rc + "/" + withStroke + "/" + withLink
												+ "/" + dir + "/" + pi + "\t" + bn + "\t" + l);
								}
		}
		for (double rc : new double[] { 0, 6, 10 }) {
			lines.add("opalecorner/" + rc + "\t"
					+ describe(net.sourceforge.plantuml.svek.image.Opale.getCorner(55.5, rc)));
			lines.add("opalepolygon/" + rc + "\t"
					+ describe(net.sourceforge.plantuml.svek.image.Opale.getPolygonNormal(55.5, 33.25, rc)));
		}
		for (int fi = 0; fi < fashions.size(); fi++)
			for (String kind : new String[] { "control", "boundary", "entity" }) {
				final Fashion fashion = fashions.get(fi);
				final List<String> log = new ArrayList<>();
				step(log, () -> {
					final TextBlock tb = kind.equals("control") ? new net.sourceforge.plantuml.svek.Control(fashion)
							: kind.equals("boundary") ? new net.sourceforge.plantuml.svek.Boundary(fashion)
									: new net.sourceforge.plantuml.svek.EntityDomain(fashion);
					final XDimension2D dim = tb.calculateDimension(FAKE);
					log.add("dim " + d(dim.getWidth()) + "x" + d(dim.getHeight()));
					tb.drawU(new Rec(log, FAKE).apply(new UTranslate(2, 3)));
				});
				for (String l : log)
					lines.add(kind + "/" + fi + "\t" + l);
			}
		write(out, "svek.golden", lines);
	}
}
