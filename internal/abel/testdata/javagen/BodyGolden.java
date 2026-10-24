// SPDX-License-Identifier: MIT
// Provenance: original
//
// BodyGolden dumps the golden values of the class body tests
// (internal/abel/body_golden_test.go) from the upstream jar: each case of
// internal/abel/testdata/body_inputs.txt is parsed as a class diagram, and
// for each leaf the body of upstream's Bodier.getBody (with the style and
// the shown portions that the svek entity images pass) is measured, its
// ports and member inner positions listed and its drawing recorded, with
// the harness's fixed StringBounder and with the SVG StringBounder of the
// reference fonts. The link constraints are drawn too. The second bounder
// needs the font set of the reference image, which internal/klimt/font
// reproduces, so the file is generated there. From the repository root:
//
//   mkdir -p ~/scratch-body && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-body internal/abel/testdata/javagen/BodyGolden.java
//   docker run --rm -v "$PWD:/repo" -v "$HOME/scratch-body:/classes" plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/classes BodyGolden /repo/internal/abel/testdata/body_inputs.txt \
//     /repo/internal/abel/testdata/golden/body.golden
//
// The drawing helpers (Rec, describe, fcString) are those of
// internal/klimt/rose/testdata/javagen/RoseGolden.java.

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

import net.atmp.CucaDiagram;
import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.abel.Entity;
import net.sourceforge.plantuml.abel.EntityPortion;
import net.sourceforge.plantuml.abel.LeafType;
import net.sourceforge.plantuml.abel.Link;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.cucadiagram.LinkConstraint;
import net.sourceforge.plantuml.cucadiagram.TextBlockCucaJSon;
import net.sourceforge.plantuml.cucadiagram.TextBlockMap;
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
import net.sourceforge.plantuml.klimt.color.HColors;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontStyle;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.geom.USegment;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.geom.XRectangle2D;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UEmpty;
import net.sourceforge.plantuml.klimt.shape.UHorizontalLine;
import net.sourceforge.plantuml.klimt.shape.UImage;
import net.sourceforge.plantuml.klimt.shape.UImageSvg;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.style.ISkinParam;
import net.sourceforge.plantuml.style.SName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.style.StyleQuery;
import net.sourceforge.plantuml.svek.WithPorts;
import net.sourceforge.plantuml.url.Url;

public class BodyGolden {

	public static void main(String[] args) throws Exception {
		final List<String> input = Files.readAllLines(Paths.get(args[0]), StandardCharsets.UTF_8);
		final List<String> out = new ArrayList<>();
		String name = null;
		StringBuilder body = new StringBuilder();
		for (String line : input) {
			if (line.startsWith("=== ")) {
				if (name != null)
					runCase(out, name, body.toString());
				name = line.substring(4);
				body = new StringBuilder();
			} else {
				body.append(line).append("\n");
			}
		}
		if (name != null)
			runCase(out, name, body.toString());
		final Path p = Paths.get(args[1]);
		Files.createDirectories(p.getParent());
		Files.write(p, (String.join("\n", out) + "\n").getBytes(StandardCharsets.UTF_8));
	}

	static String rect(XRectangle2D r) {
		if (r == null)
			return "null";
		return d(r.getMinX()) + "," + d(r.getMinY()) + "," + d(r.getWidth()) + "," + d(r.getHeight());
	}

	static void runCase(List<String> out, String name, String source) throws Exception {
		out.add("=== " + name);
		final SourceStringReader reader = new SourceStringReader("@startuml\n" + source + "@enduml\n");
		final Diagram d = reader.getBlocks().get(0).getDiagram();
		if (d instanceof CucaDiagram == false) {
			out.add("not a CucaDiagram: " + d.getClass().getSimpleName());
			return;
		}
		final CucaDiagram cd = (CucaDiagram) d;
		final ISkinParam sp = cd.getSkinParam();
		for (Entity leaf : cd.leafs()) {
			final LeafType t = leaf.getLeafType();
			final Style style;
			final boolean sm, sf;
			FontConfiguration fc = null;
			if (t.isLikeClass()) {
				style = leaf.getCurrentStyleBuilder()
						.getMergedStyle(StyleQuery.of3(SName.root, SName.element, SName.classDiagram, SName.class_)
								.withStereotype(leaf.getStereotype()).withStereostyles(leaf.getStereostyles()));
				sm = cd.showPortion(EntityPortion.METHOD, leaf);
				sf = cd.showPortion(EntityPortion.FIELD, leaf);
			} else if (t == LeafType.OBJECT) {
				style = sp.getCurrentStyleBuilder()
						.getMergedStyle(StyleQuery.of3(SName.root, SName.element, SName.objectDiagram, SName.object)
								.withStereotype(leaf.getStereotype()));
				sm = false;
				sf = cd.showPortion(EntityPortion.FIELD, leaf);
			} else if (t == LeafType.MAP || t == LeafType.JSON) {
				style = sp.getCurrentStyleBuilder().getMergedStyle(StyleQuery
						.of3(SName.root, SName.element, SName.objectDiagram, t == LeafType.MAP ? SName.map : SName.json)
						.withStereotype(leaf.getStereotype()));
				fc = style.getFontConfiguration(sp.getIHtmlColorSet());
				sm = false;
				sf = false;
			} else {
				style = leaf.getCurrentStyleBuilder()
						.getMergedStyle(StyleQuery.of3(SName.root, SName.element, SName.classDiagram, SName.class_)
								.withStereotype(leaf.getStereotype()));
				sm = true;
				sf = true;
			}
			out.add(leaf.getName() + " " + t + " sm=" + sm + " sf=" + sf);
			final FontConfiguration fc2 = fc;
			for (int bi = 0; bi < 2; bi++) {
				final StringBounder b = bi == 0 ? FAKE : SVG;
				final List<String> log = new ArrayList<>();
				step(log, () -> {
					final TextBlock tb = leaf.getBodier().getBody(sp, sm, sf, leaf.getStereotype(), style, fc2);
					if (tb == null) {
						log.add("null");
						return;
					}
					final XDimension2D dim = tb.calculateDimension(b);
					log.add("dim " + d(dim.getWidth()) + "x" + d(dim.getHeight()));
					if (tb instanceof TextBlockMap)
						((TextBlockMap) tb).setTotalWidth(dim.getWidth() + 20);
					if (tb instanceof TextBlockCucaJSon)
						((TextBlockCucaJSon) tb).setTotalWidth(dim.getWidth() + 20);
					if (tb instanceof WithPorts)
						log.add("ports " + ((WithPorts) tb).getPorts(b));
					if (t.isLikeClass() || t == LeafType.OBJECT) {
						final List<String> members = new ArrayList<>();
						for (CharSequence cs : leaf.getBodier().getRawBody())
							members.add(cs.toString());
						members.add("zzz");
						for (String m : members)
							step(log, () -> log.add("inner '" + m + "' " + rect(tb.getInnerPosition(m, b))));
					}
					tb.drawU(new Rec(log, b).apply(new UTranslate(3, 4)));
				});
				for (String l : log)
					out.add(name + "\t" + (bi == 0 ? "fake" : "svg") + "\t" + l);
			}
		}
		LinkConstraint done = null;
		for (Link link : cd.getLinks()) {
			final LinkConstraint c = link.getLinkConstraint();
			if (c == null || c == done)
				continue;
			done = c;
			for (int bi = 0; bi < 2; bi++) {
				final StringBounder b = bi == 0 ? FAKE : SVG;
				final List<String> log = new ArrayList<>();
				step(log, () -> {
					final List<Link> links = cd.getLinks();
					Link l1 = null, l2 = null;
					for (Link l : links)
						if (l.getLinkConstraint() == c) {
							if (l1 == null)
								l1 = l;
							else
								l2 = l;
						}
					c.setPosition(l1, new XPoint2D(10, 20));
					c.setPosition(l2, new XPoint2D(60, 90));
					c.drawMe(new Rec(log, b), sp);
				});
				for (String l : log)
					out.add(name + "\tconstraint " + (bi == 0 ? "fake" : "svg") + "\t" + l);
			}
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

}
