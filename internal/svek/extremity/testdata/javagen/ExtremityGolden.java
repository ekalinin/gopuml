// SPDX-License-Identifier: MIT
// Provenance: original
//
// ExtremityGolden dumps the golden values of internal/svek/extremity
// (golden_test.go) from the upstream jar: for every LinkDecor (the legacy
// and the complete factories) and every LinkMiddleDecor, the decorations
// built from several Graphviz arrow-head triangles, angles and sides, their
// somePoint, decoration length and qualifier move, and the drawing
// operations recorded on graphics of several states. It is in the package
// of the extremities to reach their package-private classes. From the
// repository root:
//
//   mkdir -p ~/scratch-extremity && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.8.jar \
//     -d ~/scratch-extremity internal/svek/extremity/testdata/javagen/ExtremityGolden.java
//   docker run --rm -v "$PWD:/repo" -v "$HOME/scratch-extremity:/classes" plantuml-go-reference \
//     java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp /opt/plantuml.jar:/classes net.sourceforge.plantuml.svek.extremity.ExtremityGolden \
//     /repo/internal/svek/extremity/testdata/golden
//
// It writes extremity.golden.gz.

package net.sourceforge.plantuml.svek.extremity;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

import net.sourceforge.plantuml.decoration.LinkDecor;
import net.sourceforge.plantuml.decoration.LinkMiddleDecor;
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
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.geom.Side;
import net.sourceforge.plantuml.klimt.geom.USegment;
import net.sourceforge.plantuml.klimt.geom.XPoint2D;
import net.sourceforge.plantuml.klimt.shape.UDrawable;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.url.Url;

public class ExtremityGolden {

	public static void main(String[] args) throws Exception {
		final List<String> lines = new ArrayList<>();
		decors(lines);
		middles(lines);
		direct(lines);
		writeGzip(Paths.get(args[0]), "extremity.golden.gz", lines);
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

	static String pt(XPoint2D p) {
		return p == null ? "null" : p.getX() + ":" + p.getY();
	}

	// ------------------------------------------------------------------
	// Cases

	static final HColor[] BACKS = { HColors.WHITE, HColors.MY_YELLOW, null };

	// Graphviz arrow-head triangles: base corner, tip, base corner.
	static final double[][] TRIPLES = { //
			{ 96, 120, 100, 130, 104, 120 }, // down
			{ 104, 50, 100, 40, 96, 50 }, // up
			{ 50, 96, 60, 100, 50, 104 }, // right
			{ 50, 104, 40, 100, 50, 96 }, // left
			{ 10.5, 20.25, 17.3, 31.7, 19.1, 18.4 }, // oblique
			{ 5, 5, 5, 5, 5, 5 }, // degenerate
			{ 0, 0, 3.5, 7.25, 0.0005, 10 }, // nearly vertical
			{ 200.123, 12.5, 190.75, 14.125, 199.5, 21.875 }, // other oblique
			{ 30, 30, 31, 31, 30.5, 30 }, // tiny
	};

	static final double[] ANGLES = { 0, Math.PI / 2, Math.PI, -Math.PI / 4, 1.234, 3 * Math.PI / 2,
			Math.PI / 2 + 0.0005, 2 * Math.PI - 0.0001, -2.5, 0.75 };

	static final double[][] POINTS = { { 0, 0 }, { 33.5, 47.25 } };

	static final Side[] SIDES = { null, Side.NORTH, Side.EAST, Side.SOUTH, Side.WEST };

	static Side[] sides(LinkDecor d) {
		if (d == LinkDecor.CROWFOOT)
			return SIDES;
		return new Side[] { null, Side.SOUTH };
	}

	static String side(Side s) {
		return s == null ? "null" : s.name();
	}

	static String name(Object o) {
		return o == null ? "null" : o.getClass().getSimpleName();
	}

	static void decors(List<String> lines) {
		for (LinkDecor d : LinkDecor.values())
			for (HColor back : BACKS) {
				final ExtremityFactory legacy = d.getExtremityFactoryLegacy(back);
				final ExtremityFactory complete = d.getExtremityFactoryComplete(back);
				lines.add("decor " + d + " back=" + hc(back) + " legacy=" + name(legacy) + " complete="
						+ name(complete));
				if (legacy != null)
					factory(lines, d + "/" + hc(back) + "/legacy", legacy, sides(d));
				if (complete != null && d == LinkDecor.EXTENDS)
					factory(lines, d + "/" + hc(back) + "/complete", complete, sides(d));
			}
		// The factory that SvekEdge uses for EXTENDS inside clusters.
		for (HColor back : BACKS)
			factory(lines, "extends/" + hc(back), new ExtremityFactoryExtends(back), SIDES);
	}

	static void factory(List<String> lines, String prefix, ExtremityFactory f, Side[] sides) {
		for (Side s : sides) {
			for (int i = 0; i < TRIPLES.length; i++) {
				final double[] t = TRIPLES[i];
				final XPoint2D p0 = new XPoint2D(t[0], t[1]);
				final XPoint2D p1 = new XPoint2D(t[2], t[3]);
				final XPoint2D p2 = new XPoint2D(t[4], t[5]);
				record(lines, prefix + "/tbr" + i + "/" + side(s), () -> f.createTBRDrawableLegacy(p0, p1, p2, s));
			}
			for (int j = 0; j < POINTS.length; j++)
				for (int i = 0; i < ANGLES.length; i++) {
					final XPoint2D p0 = new XPoint2D(POINTS[j][0], POINTS[j][1]);
					final double angle = ANGLES[i];
					record(lines, prefix + "/u" + j + "_" + i + "/" + side(s), () -> f.createUDrawable(p0, angle, s));
				}
		}
	}

	interface Maker {
		UDrawable make();
	}

	// Graphics states: default, colored with a thick stroke, without color.
	static Rec[] states(List<String> log) {
		return new Rec[] { new Rec(log, new UTranslate(0, 0), HColors.BLACK, HColors.none(), UStroke.simple()),
				new Rec(log, new UTranslate(0, 0), HColors.MY_RED, HColors.BLUE, UStroke.withThickness(2.5)),
				new Rec(log, new UTranslate(0, 0), null, HColors.none(), UStroke.simple()) };
	}

	static void record(List<String> lines, String prefix, Maker m) {
		final List<String> log = new ArrayList<>();
		final UDrawable[] holder = new UDrawable[1];
		step(log, () -> {
			holder[0] = m.make();
			log.add("class " + name(holder[0]));
		});
		final UDrawable u = holder[0];
		if (u instanceof Extremity) {
			final Extremity e = (Extremity) u;
			step(log, () -> log.add("somePoint " + pt(e.somePoint())));
			step(log, () -> log.add("length " + e.getDecorationLength()));
			step(log, () -> log.add("kal " + e.getDeltaForKal().getDx() + "," + e.getDeltaForKal().getDy()));
		}
		if (u != null) {
			int k = 0;
			for (Rec r : states(log)) {
				log.add("state " + k++);
				step(log, () -> u.drawU(r.apply(new UTranslate(3, 4))));
				if (u instanceof ExtremityArrow)
					step(log, () -> ((ExtremityArrow) u).drawLineIfTransparent(r));
			}
		}
		for (String l : log)
			lines.add(prefix + "\t" + l);
	}

	static void middles(List<String> lines) {
		final HColor[][] colors = { { HColors.WHITE, HColors.WHITE }, { HColors.MY_YELLOW, HColors.BLUE },
				{ null, null } };
		for (LinkMiddleDecor d : LinkMiddleDecor.values())
			for (HColor[] c : colors) {
				final String prefix = "middle " + d + "/" + hc(c[0]) + "/" + hc(c[1]);
				final List<String> log = new ArrayList<>();
				final MiddleFactory[] holder = new MiddleFactory[1];
				step(log, () -> {
					holder[0] = d.getMiddleFactory(c[0], c[1]);
					log.add("factory " + name(holder[0]));
				});
				for (String l : log)
					lines.add(prefix + "\t" + l);
				if (holder[0] == null)
					continue;
				for (int i = 0; i < ANGLES.length; i++) {
					final double angle = ANGLES[i] * 57;
					record(lines, prefix + "/" + i, () -> holder[0].createUDrawable(angle));
				}
			}
	}

	// Classes that no factory builds by default.
	static void direct(List<String> lines) {
		for (int i = 0; i < ANGLES.length; i++) {
			final double a = ANGLES[i];
			final XPoint2D c = new XPoint2D(12.5, 7.75);
			record(lines, "arrow/" + i, () -> new ExtremityArrow(c, a));
			record(lines, "arrowcenter/" + i, () -> new ExtremityArrow(c, a, new XPoint2D(20, 30)));
			record(lines, "arrowshort/" + i, () -> new ExtremityArrow(c, a, new XPoint2D(13, 8)));
			record(lines, "stateline1/" + i, () -> new ExtremityStateLine1(a, c));
			record(lines, "stateline2/" + i, () -> new ExtremityStateLine2(a, c));
			record(lines, "parenthesis2/" + i, () -> new ExtremityParenthesis2(c, a, new XPoint2D(40.5, 3.25)));
			record(lines, "halfarrow/" + i, () -> new ExtremityHalfArrow(c, a, -1));
		}
		final UPolygon poly = new UPolygon();
		poly.addPoint(1, 2);
		poly.addPoint(5.5, 7.25);
		poly.addPoint(0, 9);
		record(lines, "other", () -> new ExtremityOther(poly));
		record(lines, "otherempty", () -> new ExtremityOther(new UPolygon()));
		record(lines, "triangle/null", () -> new ExtremityTriangle(new XPoint2D(3, 4), 0.5, false, null, 8, 3, 8));
	}

	// ------------------------------------------------------------------
	// Recording graphics

	interface Step {
		void run() throws Exception;
	}

	static void step(List<String> log, Step s) {
		try {
			s.run();
		} catch (Throwable e) {
			// The helpful messages of NullPointerException are not reproduced.
			if (e instanceof NullPointerException)
				log.add("ERR NullPointerException");
			else
				log.add("ERR " + e.getClass().getSimpleName() + " " + e.getMessage());
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
		return shape.getClass().getSimpleName();
	}

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
			return null;
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
			log.add("startUrl");
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
}
