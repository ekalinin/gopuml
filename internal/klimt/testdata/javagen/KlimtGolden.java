// SPDX-License-Identifier: MIT
// Provenance: original
//
// KlimtGolden dumps the golden values of the internal/klimt tests from the
// upstream jar. From the repository root:
//
//   mkdir -p ~/scratch-klimt && javac -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-klimt internal/klimt/testdata/javagen/KlimtGolden.java
//   java -Djava.awt.headless=true -XX:+UnlockDiagnosticVMOptions -XX:-InlineMathNatives \
//     -cp .cache/upstream/plantuml-mit-1.2026.9.jar:$HOME/scratch-klimt \
//     KlimtGolden internal/klimt/testdata/golden
//
// -XX:-InlineMathNatives makes Math.sin, Math.cos and the other elementary
// functions compute as StrictMath (fdlibm), which the Go port follows
// (design D17); HotSpot's intrinsics differ in the last bit for some inputs.
//
// It reads the *.txt and *.css inputs of the golden directory and writes the
// *.golden files next to them. The Go tests (golden_test.go) compute the same
// lines and compare.

import java.io.IOException;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.TreeSet;

import net.sourceforge.plantuml.FileFormat;
import net.sourceforge.plantuml.core.DiagramType;
import net.sourceforge.plantuml.klimt.Fashion;
import net.sourceforge.plantuml.klimt.UClip;
import net.sourceforge.plantuml.klimt.UPath;
import net.sourceforge.plantuml.klimt.UStroke;
import net.sourceforge.plantuml.klimt.UTranslate;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.color.HColors;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.klimt.drawing.LimitFinder;
import net.sourceforge.plantuml.klimt.drawing.TextLimitFinder;
import net.sourceforge.plantuml.klimt.drawing.UGraphic;
import net.sourceforge.plantuml.klimt.font.FontConfiguration;
import net.sourceforge.plantuml.klimt.font.FontParam;
import net.sourceforge.plantuml.klimt.font.StringBounder;
import net.sourceforge.plantuml.klimt.font.UFont;
import net.sourceforge.plantuml.klimt.font.UFontFace;
import net.sourceforge.plantuml.klimt.font.UFontFactory;
import net.sourceforge.plantuml.klimt.geom.HorizontalAlignment;
import net.sourceforge.plantuml.klimt.geom.MinMax;
import net.sourceforge.plantuml.klimt.geom.VerticalAlignment;
import net.sourceforge.plantuml.klimt.geom.XCubicCurve2D;
import net.sourceforge.plantuml.klimt.geom.XDimension2D;
import net.sourceforge.plantuml.klimt.shape.BigFrame;
import net.sourceforge.plantuml.klimt.shape.CircledCharacter;
import net.sourceforge.plantuml.klimt.shape.DotPath;
import net.sourceforge.plantuml.klimt.shape.GraphicStrings;
import net.sourceforge.plantuml.klimt.shape.TextBlock;
import net.sourceforge.plantuml.klimt.shape.TextBlockArrow;
import net.sourceforge.plantuml.klimt.shape.TextBlockRaw;
import net.sourceforge.plantuml.klimt.shape.TextBlockUtils;
import net.sourceforge.plantuml.klimt.shape.UDrawable;
import net.sourceforge.plantuml.klimt.shape.UEllipse;
import net.sourceforge.plantuml.klimt.shape.UEmpty;
import net.sourceforge.plantuml.klimt.shape.ULine;
import net.sourceforge.plantuml.klimt.shape.UPolygon;
import net.sourceforge.plantuml.klimt.shape.URectangle;
import net.sourceforge.plantuml.klimt.shape.UText;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.ConfigurationStore;
import net.sourceforge.plantuml.skin.AlignmentParam;
import net.sourceforge.plantuml.skin.ArrowDirection;
import net.sourceforge.plantuml.skin.ColorParam;
import net.sourceforge.plantuml.skin.CornerParam;
import net.sourceforge.plantuml.skin.LineParam;
import net.sourceforge.plantuml.skin.Padder;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.SkinParam;
import net.sourceforge.plantuml.skin.VisibilityModifier;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.style.ClockwiseTopRightBottomLeft;
import net.sourceforge.plantuml.style.FromSkinparamToStyle;
import net.sourceforge.plantuml.style.LevelConstraint;
import net.sourceforge.plantuml.style.PName;
import net.sourceforge.plantuml.style.SName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.style.StyleAtom;
import net.sourceforge.plantuml.style.StyleBuilder;
import net.sourceforge.plantuml.style.StyleLoader;
import net.sourceforge.plantuml.style.StyleQuery;
import net.sourceforge.plantuml.text.Guillemet;
import net.sourceforge.plantuml.url.Url;
import net.sourceforge.plantuml.url.UrlBuilder;
import net.sourceforge.plantuml.url.UrlMode;
import net.sourceforge.plantuml.utils.BlocLines;
import net.sourceforge.plantuml.utils.Direction;

public class KlimtGolden {

	static Path dir;

	public static void main(String[] args) throws Exception {
		dir = Paths.get(args[0]);
		styles();
		fromSkinparam();
		skinParam();
		display();
		stereotypes();
		urls();
		cleanForKey();
		textBlocks();
	}

	static List<String> read(String name) throws IOException {
		return Files.readAllLines(dir.resolve(name), StandardCharsets.UTF_8);
	}

	static void write(String name, List<String> lines) throws IOException {
		Files.write(dir.resolve(name), (String.join("\n", lines) + "\n").getBytes(StandardCharsets.UTF_8));
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

	// ------------------------------------------------------------------
	// Styles

	static StyleQuery signature(String spec) {
		StyleQuery sig = StyleQuery.empty();
		for (String tok : spec.trim().split(" +")) {
			if (tok.isEmpty())
				continue;
			if (tok.startsWith("@"))
				sig = sig.withStereotype(tok.substring(1));
			else if (tok.startsWith("#"))
				sig = sig.addLevel(Integer.parseInt(tok.substring(1)));
			else if (tok.equals("*"))
				sig = sig.addStar();
			else
				sig = sig.add(SName.valueOf(tok));
		}
		return sig;
	}

	// The atoms and the level constraint of a query are package-private.
	static Object invoke(Object target, String method) {
		try {
			final Method m = target.getClass().getDeclaredMethod(method);
			m.setAccessible(true);
			return m.invoke(target);
		} catch (ReflectiveOperationException e) {
			throw new IllegalStateException(e);
		}
	}

	static String canonical(StyleQuery sig) {
		final List<String> names = new ArrayList<>();
		final TreeSet<String> st = new TreeSet<>();
		for (StyleAtom atom : (Iterable<StyleAtom>) invoke(sig, "getAtoms"))
			if (atom.isName())
				names.add(atom.getSName().name());
			else
				st.add(atom.getStereotype());
		final LevelConstraint level = (LevelConstraint) invoke(sig, "getLevelConstraint");
		return String.join(".", names) + st + (level.hasLevel() ? "#" + level.getLevel() : "")
				+ (level.isStar() ? "*" : "");
	}

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

	static String derived(Style style, HColorSet set) {
		if (style == null)
			return "null";
		final StringBuilder sb = new StringBuilder();
		sb.append("font=" + item(() -> uf(style.getUFont())));
		sb.append(" stroke=" + item(() -> style.getStroke()));
		sb.append(" padding=" + item(() -> style.getPadding()));
		sb.append(" margin=" + item(() -> style.getMargin()));
		sb.append(" align=" + item(() -> style.getHorizontalAlignment()));
		sb.append(" shadow=" + item(() -> style.getShadowing()));
		sb.append(" wrap=" + item(() -> style.wrapWidth()));
		sb.append(" back=" + item(() -> style.value(PName.BackGroundColor).asColor(set).asString()));
		sb.append(" line=" + item(() -> style.value(PName.LineColor).asColor(set).asString()));
		sb.append(" fontcolor=" + item(() -> style.value(PName.FontColor).asColor(set).asString()));
		sb.append(" round=" + item(() -> style.value(PName.RoundCorner).asDouble()));
		sb.append(" fontsizeint=" + item(() -> style.value(PName.FontSize).asInt()));
		sb.append(" fashion=" + item(() -> {
			final Fashion f = style.getSymbolContext(set);
			return f.getStroke() + "/" + f.getDeltaShadow() + "/" + f.getRoundCorner() + "/" + f.getDiagonalCorner();
		}));
		sb.append(" fc=" + item(() -> {
			final FontConfiguration fc = style.getFontConfiguration(set);
			return uf(fc.getFont()) + "/" + fc.getColor().asString() + "/" + fc.getUnderlineStroke();
		}));
		return sb.toString();
	}

	static void dumpStyles(String name, StyleBuilder builder) throws IOException {
		final HColorSet set = HColorSet.instance();
		final List<String> out = new ArrayList<>();
		final List<String> out2 = new ArrayList<>();
		for (String spec : read("signatures.txt")) {
			final StyleQuery sig = signature(spec);
			final Style style = builder.getMergedStyle(sig);
			out.add(spec + "\t" + canonical(sig) + "\t" + styleValues(style));
			out2.add(spec + "\t" + derived(style, set));
		}
		write(name + ".golden", out);
		write(name + "_derived.golden", out2);
	}

	static void styles() throws Exception {
		final StyleBuilder base = StyleLoader.loadSkin("plantuml.skin");
		dumpStyles("style_default", base);

		final StyleBuilder user = StyleLoader.loadSkin("plantuml.skin");
		final BlocLines lines = BlocLines.load(Files.newInputStream(dir.resolve("user_style.css")), null);
		final Collection<Style> parsed = StyleLoader.parseStyleText(lines.subExtract(1, 1), user);
		final List<String> out = new ArrayList<>();
		for (Style s : parsed)
			out.add(canonical(s.getQuery()) + "\t" + styleValues(s));
		write("style_user_parsed.golden", out);
		dumpStyles("style_user", user.muteStyle(parsed));

		for (String skin : new String[] { "rose.skin", "debug.skin", "sonyxperiadev.skin" })
			dumpStyles("style_" + skin.replace(".skin", ""), StyleLoader.loadSkin(skin));

		final List<String> single = new ArrayList<>();
		for (String s : new String[] { "BackGroundColor:red;", "FontSize: 12; LineColor: blue;",
				"BackgroundColor:#FF0000;LineThickness:2;", "FontColor : #ABC ; FontStyle : italic ;" }) {
			// As Stereogroup.getInnerColors parses a bare property list.
			final Style st = StyleLoader
					.parseStyleText(BlocLines.singleString("inlineStereogroupColors {" + s + "}"),
							new net.sourceforge.plantuml.style.AutomaticCounterBasic())
					.iterator().next();
			single.add(esc(s) + "\t" + styleValues(st));
		}
		write("style_singleline.golden", single);
	}

	// ------------------------------------------------------------------
	// FromSkinparamToStyle

	static void fromSkinparam() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String line : read("skinparams.txt")) {
			final String[] kv = line.split("\t", 2);
			final SkinParam sp = newSkinParam();
			final Method clean = SkinParam.class.getDeclaredMethod("cleanForKeySlow", String.class);
			clean.setAccessible(true);
			@SuppressWarnings("unchecked")
			final List<String> keys = (List<String>) clean.invoke(sp, kv[0]);
			for (String key : keys) {
				final FromSkinparamToStyle conv = new FromSkinparamToStyle(key);
				conv.convertNow(kv[1], new net.sourceforge.plantuml.style.AutomaticCounterBasic());
				final StringBuilder sb = new StringBuilder();
				for (Style s : conv.getStyles()) {
					if (sb.length() > 0)
						sb.append(" | ");
					sb.append(canonical(s.getQuery()) + " " + styleValues(s));
				}
				out.add(esc(line) + "\t" + key + "\t" + sb);
			}
		}
		write("fromskinparam.golden", out);
	}

	static SkinParam newSkinParam() {
		return SkinParam.create(new HashMap<String, String>(), PathSystem.fetch(), DiagramType.SEQUENCE,
				Pragma.createEmpty(), ConfigurationStore.createEmpty());
	}

	static String uf(UFont f) {
		return f.getFamily(null, net.sourceforge.plantuml.klimt.font.UFontContext.SVG) + "/" + f.getFontFace() + "/"
				+ f.getSize();
	}

	static String hc(HColor c) {
		return c == null ? "null" : c.asString();
	}

	static void skinParam() throws Exception {
		final SkinParam sp = newSkinParam();
		final List<String> out = new ArrayList<>();
		for (String line : read("skinparams.txt")) {
			final String[] kv = line.split("\t", 2);
			sp.setParam(kv[0], kv[1]);
		}
		final StyleBuilder builder = sp.getCurrentStyleBuilder();
		dumpStyles("style_skinparam", builder);
		final Stereotype foo = Stereotype.build("<<foo>>");
		out.add("value participantbackgroundcolor\t" + sp.getValue("participantbackgroundcolor"));
		out.add("value missing\t" + sp.getValue("missing"));
		out.add("hyperlink\t" + hc(sp.getHyperlinkColor()));
		out.add("background\t" + hc(sp.getBackgroundColor()));
		out.add("html activityBackground\t" + hc(sp.getHtmlColor(ColorParam.activityBackground, null, false)));
		out.add("html noteBackground foo\t" + hc(sp.getHtmlColor(ColorParam.noteBackground, foo, false)));
		out.add("font note\t" + uf(sp.getFont(null, false, FontParam.NOTE)));
		out.add("font note foo\t" + uf(sp.getFont(foo, false, FontParam.NOTE)));
		out.add("font package\t" + uf(sp.getFont(null, true, FontParam.PACKAGE)));
		out.add("font circled\t" + uf(sp.getFont(null, false, FontParam.CIRCLED_CHARACTER)));
		out.add("fontcolor note\t" + hc(sp.getFontHtmlColor(null, FontParam.NOTE)));
		out.add("fontcolor footer\t" + hc(sp.getFontHtmlColor(null, FontParam.FOOTER)));
		out.add("thickness arrow\t" + sp.getThickness(LineParam.arrow, null));
		out.add("thickness legendBorder\t" + sp.getThickness(LineParam.legendBorder, null));
		out.add("thickness noteBorder\t" + sp.getThickness(LineParam.noteBorder, null));
		out.add("roundcorner default\t" + sp.getRoundCorner(CornerParam.DEFAULT, null));
		out.add("roundcorner titleBorder\t" + sp.getRoundCorner(CornerParam.titleBorder, null));
		out.add("diagonal rectangle\t" + sp.getDiagonalCorner(CornerParam.rectangle, null));
		out.add("shadowing\t" + sp.shadowing(null));
		out.add("shadowingForNote\t" + sp.shadowingForNote(null));
		out.add("strictuml\t" + sp.strictUmlStyle());
		out.add("underline\t" + sp.useUnderlineForHyperlink());
		out.add("circledradius\t" + sp.getCircledCharacterRadius());
		out.add("align seqmsg\t"
				+ sp.getHorizontalAlignment(AlignmentParam.sequenceMessageAlignment, ArrowDirection.LEFT_TO_RIGHT_NORMAL,
						false, null));
		out.add("align seqmsgtext\t" + sp.getHorizontalAlignment(AlignmentParam.sequenceMessageTextAlignment,
				ArrowDirection.RIGHT_TO_LEFT_REVERSE, false, null));
		out.add("align note\t" + sp.getHorizontalAlignment(AlignmentParam.noteTextAlignment,
				ArrowDirection.LEFT_TO_RIGHT_NORMAL, false, null));
		out.add("padding\t" + sp.getPadding());
		out.add("dpi\t" + sp.getDpi());
		out.add("tabsize\t" + sp.getTabSize());
		out.add("guillemet\t" + sp.guillemet().manageGuillemet("<<x>>"));
		out.add("forceUnderline\t" + sp.forceSequenceParticipantUnderlined());
		out.add("padder\t" + (sp.sequenceDiagramPadder() == Padder.NONE));
		out.add("splitparam\t" + sp.getSplitParam().isSet());
		out.add("componentstyle\t" + sp.componentStyle());
		out.add("packagestyle\t" + sp.packageStyle());
		out.add("conditionstyle\t" + sp.getConditionStyle());
		out.add("maxmessagesize\t" + sp.maxMessageSize());
		out.add("lengthadjust\t" + sp.getlengthAdjust());
		out.add("missingroot\t" + StyleLoader.getMissingRootProperties(builder));
		write("skinparam.golden", out);
	}

	// ------------------------------------------------------------------
	// Display, stereotypes, urls, keys

	static void display() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String line : read("display_inputs.txt")) {
			final Display d = Display.getWithNewlines(Pragma.createEmpty(), line);
			final List<String> parts = new ArrayList<>();
			for (CharSequence cs : d)
				parts.add(esc(cs.toString()));
			out.add(esc(line) + "\t" + d.size() + "\t" + String.join("|", parts) + "\t"
					+ d.getNaturalHorizontalAlignment() + "\t" + esc(String.valueOf(Display.getWithNewlines3(line))) + "\t"
					+ d.hasSeveralGuideLines() + "\t" + d.hasUrl() + "\t" + d.isWhite());
		}
		write("display.golden", out);
	}

	static void stereotypes() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String line : read("stereotypes.txt")) {
			final Stereotype simple = Stereotype.build(line);
			String complex;
			try {
				final Stereotype c = Stereotype.build(line, 6, UFontFactory.monospace(17), HColorSet.instance());
				complex = c.toString() + "|" + (int) c.getCharacter() + "|" + hc(c.getHtmlColor()) + "|"
						+ c.getStyleNames() + "|" + c.getLabels(Guillemet.GUILLEMET);
			} catch (Exception e) {
				complex = "ERR " + e.getClass().getSimpleName();
			}
			out.add(esc(line) + "\t" + simple + "\t" + simple.getLabel(Guillemet.GUILLEMET) + "\t"
					+ simple.getLabel(Guillemet.DOUBLE_COMPARATOR) + "\t" + simple.getLabels(Guillemet.NONE) + "\t"
					+ simple.getMultipleLabels() + "\t" + simple.getStyleNames() + "\t" + simple.getPackageStyle()
					+ "\t" + simple.isWithOOSymbol() + "\t" + simple.isSpotted() + "\t" + complex + "\t"
					+ canonical(StyleQuery.of3(SName.root, SName.element).withStereotype(simple)));
		}
		write("stereotypes.golden", out);
	}

	static String url(Url u) {
		if (u == null)
			return "null";
		return u.getUrl() + "|" + u.getTooltip() + "|" + u.getLabel();
	}

	static void urls() throws Exception {
		final List<String> out = new ArrayList<>();
		for (String line : read("urls.txt"))
			out.add(line + "\t" + url(new UrlBuilder(null, UrlMode.ANYWHERE).getUrl(line)) + "\t"
					+ url(new UrlBuilder(null, UrlMode.STRICT).getUrl(line)) + "\t"
					+ url(new UrlBuilder("http://top/", UrlMode.ANYWHERE).getUrl(line)));
		write("urls.golden", out);
	}

	static void cleanForKey() throws Exception {
		final Method clean = SkinParam.class.getDeclaredMethod("cleanForKeySlow", String.class);
		clean.setAccessible(true);
		final List<String> out = new ArrayList<>();
		final SkinParam sp = newSkinParam();
		for (String line : read("cleanforkey_inputs.txt"))
			out.add(line + "\t" + clean.invoke(sp, line));
		write("cleanforkey.golden", out);
	}

	// ------------------------------------------------------------------
	// Text blocks and LimitFinder

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

	static String mm(MinMax m) {
		return m.getMinX() + "," + m.getMinY() + "," + m.getMaxX() + "," + m.getMaxY();
	}

	static String block(TextBlock tb) {
		final XDimension2D d = tb.calculateDimension(BOUNDER);
		return d.getWidth() + "x" + d.getHeight() + "\t" + mm(TextBlockUtils.getMinMax(tb, BOUNDER, false)) + "\t"
				+ mm(TextBlockUtils.getMinMax(tb, BOUNDER, true));
	}

	static void textBlocks() throws Exception {
		final List<String> out = new ArrayList<>();
		final TextBlock t1 = raw(14, "Hello", "World!", "");
		out.add("raw\t" + block(t1));
		out.add("error\t" + block(GraphicStrings.createForError(Arrays.asList("Syntax error", "line 2"), true)));
		out.add("mergeTB\t" + block(TextBlockUtils.mergeTB(t1, TextBlockUtils.empty(30, 40), HorizontalAlignment.CENTER)));
		out.add("mergeTBright\t"
				+ block(TextBlockUtils.mergeTB(raw(10, "a"), raw(12, "bbbbbb"), HorizontalAlignment.RIGHT)));
		out.add("mergeLR\t" + block(TextBlockUtils.withMargin(
				TextBlockUtils.mergeLR(raw(13, "A"), raw(11, "BBB\tC"), VerticalAlignment.BOTTOM), 5, 7)));
		out.add("mergeLRcenter\t"
				+ block(TextBlockUtils.mergeLR(raw(20, "big"), raw(8, "s"), VerticalAlignment.CENTER)));
		out.add("bordered\t" + block(TextBlockUtils.bordered(raw(12, "note"), UStroke.withThickness(1.5), HColors.RED,
				HColors.WHITE, 5, ClockwiseTopRightBottomLeft.topRightBottomLeft(2, 3, 4, 5), "id")));
		out.add("minwidth\t" + block(TextBlockUtils.withMinWidth(raw(12, "x"), 100, HorizontalAlignment.RIGHT)));
		out.add("minwidthcenter\t"
				+ block(TextBlockUtils.withMinWidth(raw(12, "xyz"), 50, HorizontalAlignment.CENTER)));
		out.add("margin4\t" + block(TextBlockUtils.withMargin(raw(12, "m"), 1, 2, 3, 4)));
		out.add("bigframe\t" + block(new BigFrame(raw(12, "frame"), raw(12, "line one", "line two"),
				ClockwiseTopRightBottomLeft.same(3), new Fashion(HColors.WHITE, HColors.BLACK).withShadow(2))));
		out.add("bigframeempty\t" + block(new BigFrame(TextBlockUtils.empty(0, 0), raw(12, "content"),
				ClockwiseTopRightBottomLeft.none(), new Fashion(null, HColors.BLACK))));
		out.add("shapes\t" + block(new ShapesBlock()));
		out.add("padder\t" + block(Padder.NONE.withMargin(ClockwiseTopRightBottomLeft.same(2))
				.withPadding(ClockwiseTopRightBottomLeft.margin1margin2(3, 4)).withBorderColor(HColors.RED)
				.apply(raw(12, "padded"))));
		out.add("visibility\t" + block(VisibilityModifier.PROTECTED_METHOD.getUBlock(10, HColors.BLACK, HColors.RED,
				true)));
		out.add("visibilityfield\t"
				+ block(VisibilityModifier.PACKAGE_PRIVATE_FIELD.getUBlock(9, HColors.BLACK, null, false)));
		out.add("circled\t" + block(new CircledCharacter('C', 9, UFontFactory.monospace(17), HColors.GREEN,
				HColors.BLACK, HColors.BLACK)));
		out.add("arrow\t" + block(new TextBlockArrow(Direction.RIGHT, fc(13))));
		out.add("arrowup\t" + block(new TextBlockArrow(Direction.UP, fc(9))));
		out.add("backcolor\t" + block(TextBlockUtils.mergeTB(TextBlockUtils.addBackcolor(raw(12, "colored"), HColors.RED),
				raw(12, "plain text"), HorizontalAlignment.LEFT)));
		final TextLimitFinder tlf = TextLimitFinder.create(BOUNDER, false);
		t1.drawU(tlf);
		out.add("textlimitfinder\t" + tlf.getMinX() + "," + tlf.getMinY() + "," + tlf.getMaxX() + "," + tlf.getMaxY());
		write("textblock.golden", out);
	}

	static class ShapesBlock implements TextBlock {
		@Override
		public void drawU(UGraphic ug) {
			final URectangle r = URectangle.build(10, 20);
			r.setDeltaShadow(3);
			ug.apply(new UTranslate(5, 5)).draw(r);
			ug.apply(new UTranslate(-3, 2)).draw(UEllipse.build(8, 8));
			final UPolygon p = new UPolygon();
			p.addPoint(0, 0);
			p.addPoint(15, 4);
			p.addPoint(3, 9);
			ug.apply(new UTranslate(40, 1)).draw(p);
			ug.apply(new UTranslate(7, 30)).draw(new ULine(12, -5));
			final UPath path = UPath.none();
			path.moveTo(0, 0);
			path.lineTo(10, 3);
			path.arcTo(5, 5, 0, 0, 1, 20, 8);
			path.cubicTo(25, 10, 30, 12, 31, -4);
			ug.apply(new UTranslate(50, 20)).draw(path);
			final DotPath dp = new DotPath().addCurve(new net.sourceforge.plantuml.klimt.geom.XPoint2D(0, 0),
					new net.sourceforge.plantuml.klimt.geom.XPoint2D(5, -8),
					new net.sourceforge.plantuml.klimt.geom.XPoint2D(9, 12),
					new net.sourceforge.plantuml.klimt.geom.XPoint2D(14, 3));
			ug.apply(new UTranslate(60, 40)).draw(dp);
			ug.apply(new UTranslate(2, 50)).draw(UText.build("text", fc(10)));
			ug.apply(new UTranslate(80, 60)).draw(new UEmpty(4, 6));
			final UGraphic clipped = ug.apply(new UClip(0, 0, 70, 70));
			clipped.apply(new UTranslate(65, 65)).draw(URectangle.build(20, 20));
			clipped.apply(new UTranslate(10, 10)).draw(URectangle.build(5, 5));
		}

		@Override
		public XDimension2D calculateDimension(StringBounder stringBounder) {
			return new XDimension2D(1, 1);
		}
	}
}
