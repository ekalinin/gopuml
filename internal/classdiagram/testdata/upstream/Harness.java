// SPDX-License-Identifier: MIT
// Provenance: original
//
// Harness parses class diagram sources with the reference PlantUML jar
// (v1.2026.9, MIT variant) and dumps the model that upstream's
// ClassDiagramFactory builds, or its error, in the text form that
// dump_test.go writes for the Go port. The Go tests compare both dumps.
//
// Regenerate the golden files in the reference image with:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/classdiagram
//
// Usage: java -cp plantuml.jar:<classes> Harness <out.txt> <case files...>
//
// A case file holds cases; a line "=== name" starts a case. A source
// without an @start line is wrapped in @startuml/@enduml. Backslash-u
// escapes (four hex digits) are decoded when the case name ends with
// " (escaped)".

import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collection;
import java.util.IdentityHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.function.Supplier;

import net.atmp.CucaDiagram;
import net.sourceforge.plantuml.BlockUml;
import net.sourceforge.plantuml.NewpagedDiagram;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.UgDiagram;
import net.sourceforge.plantuml.abel.CucaNote;
import net.sourceforge.plantuml.abel.DisplayPositioned;
import net.sourceforge.plantuml.abel.Entity;
import net.sourceforge.plantuml.abel.EntityGender;
import net.sourceforge.plantuml.abel.EntityPortion;
import net.sourceforge.plantuml.abel.LeafType;
import net.sourceforge.plantuml.abel.Link;
import net.sourceforge.plantuml.abel.LinkArg;
import net.sourceforge.plantuml.abel.Tip;
import net.sourceforge.plantuml.abel.Together;
import net.sourceforge.plantuml.classdiagram.ClassDiagram;
import net.sourceforge.plantuml.classdiagram.ClassDiagramFactory;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.core.UmlSource;
import net.sourceforge.plantuml.cucadiagram.Bodier;
import net.sourceforge.plantuml.cucadiagram.BodierJSon;
import net.sourceforge.plantuml.cucadiagram.BodierLikeClassOrObject;
import net.sourceforge.plantuml.cucadiagram.BodierMap;
import net.sourceforge.plantuml.cucadiagram.HideOrShow;
import net.sourceforge.plantuml.cucadiagram.LinkConstraint;
import net.sourceforge.plantuml.cucadiagram.Member;
import net.sourceforge.plantuml.decoration.symbol.USymbol;
import net.sourceforge.plantuml.decoration.symbol.USymbols;
import net.sourceforge.plantuml.error.PSystemError;
import net.sourceforge.plantuml.klimt.color.Colors;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.plasma.Quark;
import net.sourceforge.plantuml.preproc.PreprocessingArtifact;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.PragmaKey;
import net.sourceforge.plantuml.skin.SkinParam;
import net.sourceforge.plantuml.skin.VisibilityModifier;
import net.sourceforge.plantuml.stereo.Stereotag;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.text.StringLocated;
import net.sourceforge.plantuml.url.Url;
import net.sourceforge.plantuml.utils.Position;
import net.sourceforge.plantuml.warning.Warning;

public class Harness {

	public static void main(String[] args) throws Exception {
		final List<String> out = new ArrayList<>();
		for (int i = 1; i < args.length; i++)
			for (String[] c : readCases(Paths.get(args[i])))
				dumpCase(out, c[0], c[1]);
		Files.write(Paths.get(args[0]), out, StandardCharsets.UTF_8);
	}

	static List<String[]> readCases(Path p) throws Exception {
		final List<String[]> result = new ArrayList<>();
		String name = null;
		StringBuilder sb = null;
		for (String line : Files.readAllLines(p, StandardCharsets.UTF_8)) {
			if (line.startsWith("=== ")) {
				if (name != null)
					result.add(new String[] { name, sb.toString() });
				name = line.substring(4);
				sb = new StringBuilder();
				continue;
			}
			if (name == null)
				continue;
			if (name.endsWith(" (escaped)"))
				line = unescape(line);
			sb.append(line).append("\n");
		}
		if (name != null)
			result.add(new String[] { name, sb.toString() });
		return result;
	}

	static String unescape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (int i = 0; i < s.length(); i++) {
			final char c = s.charAt(i);
			if (c == '\\' && i + 5 < s.length() && s.charAt(i + 1) == 'u') {
				sb.append((char) Integer.parseInt(s.substring(i + 2, i + 6), 16));
				i += 5;
			} else {
				sb.append(c);
			}
		}
		return sb.toString();
	}

	static Object field(Object o, String name) {
		Class<?> c = o.getClass();
		while (c != null) {
			try {
				final Field f = c.getDeclaredField(name);
				f.setAccessible(true);
				return f.get(o);
			} catch (NoSuchFieldException e) {
				c = c.getSuperclass();
			} catch (IllegalAccessException e) {
				throw new RuntimeException(e);
			}
		}
		throw new RuntimeException("no field " + name + " in " + o.getClass());
	}

	static Object staticField(Class<?> c, String name) {
		try {
			final Field f = c.getDeclaredField(name);
			f.setAccessible(true);
			return f.get(null);
		} catch (Exception e) {
			throw new RuntimeException(e);
		}
	}

	// safe returns the text, or the exception. The message of a
	// PatternSyntaxException is left out: the Go port compiles user patterns
	// with another engine, whose messages differ.
	static String safe(Supplier<String> s) {
		try {
			return s.get();
		} catch (java.util.regex.PatternSyntaxException t) {
			return "EXC " + t.getClass().getName();
		} catch (Throwable t) {
			return "EXC " + t;
		}
	}

	static void dumpCase(List<String> out, String name, String source) {
		out.add("=== " + name);
		String src = source;
		if (src.trim().startsWith("@start") == false)
			src = "@startuml\n" + src + "@enduml\n";
		try {
			final SourceStringReader reader = new SourceStringReader(src);
			final List<BlockUml> blocks = reader.getBlocks();
			if (blocks.size() == 0) {
				out.add("no block");
				return;
			}
			final BlockUml block = blocks.get(0);
			if ((Boolean) field(block, "preprocessorError")) {
				out.add("preprocessor error");
				return;
			}
			final List<StringLocated> data = block.getData();
			@SuppressWarnings("unchecked")
			final List<StringLocated> raw = (List<StringLocated>) field(block, "rawSource");
			final PreprocessingArtifact pa = (PreprocessingArtifact) field(block, "preprocessingArtifact");
			final UmlSource umlSource = UmlSource.createWithRaw(data, true, raw);
			umlSource.patchBase64();
			final Diagram d = new ClassDiagramFactory().createSystem(PathSystem.fetch(), umlSource, null, pa);
			if (isCompact(name)) {
				final List<String> full = new ArrayList<>();
				dump(full, d);
				for (String line : full)
					if (line.startsWith("result") || line.startsWith("error") || line.startsWith("link")
							|| (name.startsWith("declcombos/") && (line.startsWith("entity ent")
									|| line.startsWith("body ") || line.startsWith("crash"))))
						out.add(line);
			} else {
				dump(out, d);
			}
		} catch (Throwable t) {
			out.add("crash " + t);
		}
	}

	// isCompact tells whether only the result, the error and the links of a
	// case are dumped (the generated cases), with the entities and bodies
	// for the generated declarations.
	static boolean isCompact(String name) {
		return name.startsWith("linkcombos/") || name.startsWith("declcombos/");
	}

	static void dump(List<String> out, Diagram d) {
		if (d == null) {
			out.add("result null");
			return;
		}
		if (d instanceof PSystemError) {
			final PSystemError e = (PSystemError) d;
			out.add("result error");
			out.add("error " + e.getFirstError().toString());
			out.add("error.message " + e.getFirstError().getError());
			out.add("error.score " + e.score());
			out.add("error.last " + e.getLineLocation().getPosition());
			return;
		}
		if (d instanceof NewpagedDiagram) {
			final NewpagedDiagram n = (NewpagedDiagram) d;
			out.add("result NewpagedDiagram");
			out.add("description " + n.getDescription());
			out.add("pages " + n.getNbImages());
			final List<UgDiagram> pages = n.getDiagrams();
			for (int i = 0; i < pages.size(); i++) {
				out.add("page " + i);
				dumpClass(out, (ClassDiagram) pages.get(i));
			}
			return;
		}
		if (d instanceof ClassDiagram) {
			out.add("result ClassDiagram");
			dumpClass(out, (ClassDiagram) d);
			return;
		}
		out.add("result " + d.getClass().getSimpleName());
	}

	static void dumpClass(List<String> out, ClassDiagram cd) {
		final IdentityHashMap<Together, Integer> togethers = new IdentityHashMap<>();
		final List<Together> togetherOrder = new ArrayList<>();
		out.add("description " + cd.getDescription());
		out.add("pages " + cd.getNbImages());
		out.add("type " + cd.getDiagramType());
		out.add("title " + dp(cd.getTitle()));
		out.add("caption " + dp(cd.getCaption()));
		out.add("legend " + dp(cd.getLegend()));
		out.add("header " + dp(cd.getHeader()));
		out.add("footer " + dp(cd.getFooter()));
		out.add("mainframe " + dp(cd.getMainFrame()));
		out.add("scale " + scale(cd.getScale()));
		out.add("splitpages " + cd.getSplitPagesHorizontal() + " " + cd.getSplitPagesVertical());
		out.add("namespaceseparator " + cd.getNamespaceSeparator());
		out.add("skinparamused " + cd.isSkinParamUsed());
		out.add("smetana " + cd.isUseSmetana() + " elk " + cd.isUseElk());
		final Pragma pragma = cd.getPragma();
		for (PragmaKey k : PragmaKey.values())
			if (pragma.isDefine(k))
				out.add("pragma " + k + "=" + pragma.getValue(k));
		for (Warning w : pragma.getWarnings())
			out.add("warning " + String.join("\\n", w.getMessage()));
		final Map<String, String> sp = new TreeMap<>(cd.getSkinParam().values());
		for (Map.Entry<String, String> ent : sp.entrySet())
			out.add("skinparam " + ent.getKey() + "=" + ent.getValue());
		final Collection<String> sprites = new TreeSet<>(((SkinParam) cd.getSkinParam()).getAllSpriteNames());
		for (String s : sprites)
			out.add("sprite " + s);
		out.add("defaultskin " + ((SkinParam) cd.getSkinParam()).getDefaultSkin());
		out.add("rankdir " + cd.getSkinParam().getRankdir());
		out.add("allowmixing " + cd.isAllowMixing() + " layoutexplicit " + field(cd, "useLayoutExplicit")
				+ " rawlayout " + field(cd, "rawLayout"));
		out.add("visibilitymodifierpresent " + cd.isVisibilityModifierPresent());
		out.add("hidevisibility " + visibilities(cd.getHidesVisibilityModifier()));
		out.add("lastentity " + uid(cd.getLastEntity()));
		final StringBuilder stack = new StringBuilder("stack");
		for (Object bag : (List<?>) field(cd, "stacks")) {
			if (bag instanceof Entity)
				stack.append(" ").append(((Entity) bag).getUid());
			else
				stack.append(" T").append(togetherId(togethers, togetherOrder, (Together) bag));
		}
		out.add(stack.toString());
		for (Object o : (List<?>) field(cd, "hideOrShows")) {
			final EntityGender gender = (EntityGender) field(o, "gender");
			out.add("hideorshow gender=" + gender.getGender() + " portion=" + field(o, "portion") + " show="
					+ field(o, "show"));
		}
		for (Object o : (List<?>) field(cd, "hides2"))
			out.add("hides2 " + o);
		for (Object o : (List<?>) field(cd, "removed"))
			out.add("removed " + o);
		for (Quark<Entity> q : cd.quarks())
			out.add("quark [" + q.getQualifiedName() + "] name=[" + q.getName() + "] parent="
					+ (q.getParent() == null ? "null" : "[" + q.getParent().getQualifiedName() + "]") + " data="
					+ uid(q.getData()) + " children=" + q.countChildren());
		for (Quark<Entity> q : cd.quarks()) {
			final Entity e = q.getData();
			if (e == null)
				continue;
			out.add("entity " + entity(cd, e, togethers, togetherOrder));
			if (e.getBodier() != null)
				out.add("body " + e.getUid() + " " + body(e));
			if (e.isRoot() == false && e.isGroup() == false
					&& (e.getLeafType().isLikeClass() || e.getLeafType() == LeafType.OBJECT)) {
				final StringBuilder sb = new StringBuilder("portions " + e.getUid());
				for (EntityPortion p : EntityPortion.values())
					sb.append(" ").append(p).append("=").append(safe(() -> "" + cd.showPortion(p, e)));
				sb.append(" stereolabels=").append(safe(() -> "" + cd.getVisibleStereotypeLabels(e)));
				out.add(sb.toString());
			}
		}
		for (Link l : cd.getLinks())
			out.add("link " + link(l));
		for (int i = 0; i < togetherOrder.size(); i++) {
			final Together t = togetherOrder.get(i);
			out.add("together T" + i + " parent="
					+ (t.getParent() == null ? "null" : "T" + togetherId(togethers, togetherOrder, t.getParent())));
		}
	}

	static int togetherId(IdentityHashMap<Together, Integer> togethers, List<Together> order, Together t) {
		Integer id = togethers.get(t);
		if (id == null) {
			id = order.size();
			togethers.put(t, id);
			order.add(t);
		}
		return id;
	}

	static String uid(Entity e) {
		return e == null ? "null" : e.getUid();
	}

	static String visibilities(Set<VisibilityModifier> set) {
		final StringBuilder sb = new StringBuilder("[");
		for (VisibilityModifier v : VisibilityModifier.values())
			if (set.contains(v))
				sb.append(v).append(";");
		return sb.append("]").toString();
	}

	static String symbolName(USymbol s) {
		if (s == null)
			return "null";
		@SuppressWarnings("unchecked")
		final Map<String, USymbol> all = (Map<String, USymbol>) staticField(USymbols.class, "all");
		for (Map.Entry<String, USymbol> ent : new TreeMap<>(all).entrySet())
			if (ent.getValue() == s)
				return ent.getKey();
		return "?" + s.getClass().getSimpleName();
	}

	static String entity(CucaDiagram cd, Entity e, IdentityHashMap<Together, Integer> togethers,
			List<Together> togetherOrder) {
		final StringBuilder sb = new StringBuilder();
		sb.append(e.getUid()).append(" quark=[").append(e.getQuark().getQualifiedName()).append("]");
		if (e.isGroup())
			sb.append(" group=").append(e.getGroupType());
		else
			sb.append(" leaf=").append(e.getLeafType());
		sb.append(" display=").append(display(e.getDisplay()));
		sb.append(" mode=").append(Display.isNull(e.getDisplay()) ? "null" : field(e.getDisplay(), "defaultCreoleMode"));
		sb.append(" stereo=").append(stereo(e.getStereotype()));
		sb.append(" styles=").append(field(e.getStereostyles(), "names"));
		final StringBuilder tags = new StringBuilder("[");
		for (Stereotag t : e.stereotags())
			tags.append(t).append(";");
		sb.append(" tags=").append(tags).append("]");
		sb.append(" generic=").append(field(e, "generic"));
		sb.append(" visibility=").append(e.getVisibilityModifier());
		sb.append(" colors=").append(colors(e.getColors()));
		sb.append(" url=").append(url(e.getUrl99()));
		sb.append(" symbol=").append(symbolName(e.getUSymbol()));
		sb.append(" static=").append(e.isStatic()).append(" packed=").append(e.isPacked());
		sb.append(" rawlayout=").append(e.getRawLayout());
		final Together t = e.getTogether();
		sb.append(" together=").append(t == null ? "null" : "T" + togetherId(togethers, togetherOrder, t));
		sb.append(" at=").append(e.getLocation() == null ? "null" : "" + e.getLocation().getPosition());
		sb.append(" legend=").append(e.getLegend() == null ? "null" : dp(e.getLegend()));
		sb.append(" notestop=").append(notes(e.getNotes(Position.TOP)));
		sb.append(" notesbottom=").append(notes(e.getNotes(Position.BOTTOM)));
		final StringBuilder tips = new StringBuilder("[");
		for (Map.Entry<String, Tip> ent : e.getTips().entrySet())
			tips.append("{").append(ent.getKey()).append(" ").append(display(ent.getValue().getDisplay())).append(" ")
					.append(colors(ent.getValue().getColors())).append(" ").append(stereo(ent.getValue().getStereotype()))
					.append("}");
		sb.append(" tips=").append(tips).append("]");
		@SuppressWarnings("unchecked")
		final Collection<String> ports = new TreeSet<>((Collection<String>) field(e, "portShortNames"));
		sb.append(" ports=").append(ports);
		if (e.isRoot() == false) {
			sb.append(" hidden=").append(safe(() -> "" + e.isHidden()));
			sb.append(" removed=").append(safe(() -> "" + e.isRemoved()));
		}
		sb.append(" position=").append(safe(() -> "" + e.getEntityPosition()));
		sb.append(" hasurl=").append(safe(() -> "" + e.hasUrl()));
		return sb.toString();
	}

	static String body(Entity e) {
		final Bodier b = e.getBodier();
		if (b instanceof BodierJSon) {
			final Object json = field(b, "json");
			return "json " + (json == null ? "null" : json.toString());
		}
		final StringBuilder sb = new StringBuilder();
		sb.append(b.getClass().getSimpleName()).append(" raw=[");
		for (CharSequence cs : b.getRawBody())
			sb.append(cs.toString()).append(";");
		sb.append("]");
		if (b instanceof BodierMap) {
			sb.append(" map=").append(field(b, "map"));
		} else if (b instanceof BodierLikeClassOrObject) {
			sb.append(" type=").append(field(b, "type"));
			sb.append(" fields=").append(safe(() -> members(b.getFieldsToDisplay())));
			sb.append(" methods=").append(safe(() -> members(b.getMethodsToDisplay())));
		}
		return sb.toString();
	}

	static String members(Display d) {
		final StringBuilder sb = new StringBuilder("[");
		for (CharSequence cs : d) {
			final Member m = (Member) cs;
			sb.append("{").append(m.getVisibilityModifier()).append("|").append(m.isStatic() ? "S" : "")
					.append(m.isAbstract() ? "A" : "").append("|").append(m.getDisplay(false)).append("|")
					.append(url(m.getUrl())).append("|").append(m.toString()).append("}");
		}
		return sb.append("]").toString();
	}

	static String notes(List<CucaNote> notes) {
		final StringBuilder sb = new StringBuilder("[");
		for (CucaNote n : notes)
			sb.append("{").append(note(n)).append("}");
		return sb.append("]").toString();
	}

	static String note(CucaNote n) {
		if (n == null)
			return "null";
		return display(n.getDisplay()) + " " + n.getPosition() + " " + colors(n.getColors()) + " " + n.getStrategy();
	}

	static String link(Link l) {
		final StringBuilder sb = new StringBuilder();
		sb.append(l.getUid()).append(" e1=").append(l.getEntity1().getUid()).append(" e2=")
				.append(l.getEntity2().getUid());
		sb.append(" type=").append(l.getType()).append(" middle=").append(l.getType().getMiddleDecor());
		final LinkArg a = l.getLinkArg();
		sb.append(" label=").append(display(a.getLabel()));
		sb.append(" q1=").append(a.getQuantifier1()).append(" q2=").append(a.getQuantifier2());
		sb.append(" r1=").append(a.getRole1()).append(" r2=").append(a.getRole2());
		sb.append(" kal1=").append(a.getKal1()).append(" kal2=").append(a.getKal2());
		sb.append(" ld=").append(a.getLabeldistance()).append(" la=").append(a.getLabelangle());
		sb.append(" length=").append(a.getLength()).append(" vis=").append(a.getVisibilityModifier());
		sb.append(" arrow=").append(l.getLinkArrow()).append(" inverted=").append(l.isInverted());
		sb.append(" invis=").append(l.isInvis()).append(" hiddenflag=").append(field(l, "hidden"));
		sb.append(" single=").append(l.isSingle()).append(" nodestyle=").append(l.useNodeStyle());
		sb.append(" constraint=").append(l.isConstraint()).append(" weight=").append(l.getWeight());
		sb.append(" solitary=").append(l.isHorizontalSolitary()).append(" opale=").append(l.isOpale());
		sb.append(" sametail=").append(l.getSametail());
		sb.append(" ports=").append(l.getPortName1()).append("/").append(l.getPortName2());
		sb.append(" url=").append(url(l.getUrl()));
		sb.append(" stereo=").append(stereo(l.getStereotype()));
		sb.append(" colors=").append(colors(l.getColors()));
		final StringBuilder sup = new StringBuilder("[");
		for (Colors c : l.getSupplementaryColors())
			sup.append(colors(c)).append(";");
		sb.append(" supplementary=").append(sup).append("]");
		sb.append(" note=").append(note(l.getNote()));
		final LinkConstraint lc = l.getLinkConstraint();
		if (lc == null)
			sb.append(" linkconstraint=null");
		else
			sb.append(" linkconstraint=").append(((Link) field(lc, "link1")).getUid()).append("/")
					.append(((Link) field(lc, "link2")).getUid()).append(" ")
					.append(display((Display) field(lc, "display")));
		sb.append(" codeline=").append(l.getCodeLine());
		sb.append(" at=").append(l.getLocation() == null ? "null" : "" + l.getLocation().getPosition());
		sb.append(" hidden=").append(safe(() -> "" + l.isHidden()));
		sb.append(" removed=").append(safe(() -> "" + l.isRemoved()));
		return sb.toString();
	}

	static String scale(Object scale) {
		if (scale == null)
			return "null";
		final String cls = scale.getClass().getSimpleName();
		final StringBuilder sb = new StringBuilder(cls);
		for (String f : new String[] { "scale", "maxWidth", "maxHeight" }) {
			try {
				final Object value = field(scale, f);
				sb.append(" ").append(f).append("=").append(value);
			} catch (RuntimeException e) {
			}
		}
		return sb.toString();
	}

	static String color(HColor c) {
		return c == null ? "null" : c.toString();
	}

	// colors writes the colors without the identity hash code that
	// gradients print.
	static String colors(Colors c) {
		if (c == null)
			return "null";
		return "{" + c.toString().replaceAll("(HColorGradient)@[0-9a-f]+", "$1") + " shadowing=" + c.getShadowing()
				+ "}";
	}

	static String display(Display d) {
		if (Display.isNull(d))
			return "NULL";
		final StringBuilder sb = new StringBuilder("[");
		boolean first = true;
		for (CharSequence cs : d) {
			if (first == false)
				sb.append("|");
			first = false;
			if (cs instanceof Stereotype)
				sb.append("S:");
			else if (cs instanceof Member)
				sb.append("M:");
			sb.append(cs.toString().replace("\n", "\\n"));
		}
		sb.append("]");
		if (d.getNaturalHorizontalAlignment() != null)
			sb.append("@").append(d.getNaturalHorizontalAlignment());
		return sb.toString();
	}

	static String dp(DisplayPositioned dp) {
		if (dp == null)
			return "null";
		return display(dp.getDisplay()) + " h=" + dp.getHorizontalAlignment() + " v=" + dp.getVerticalAlignment()
				+ " at=" + (dp.getLineLocation() == null ? "null" : "" + dp.getLineLocation().getPosition());
	}

	static String url(Url u) {
		if (u == null)
			return "null";
		return "{" + u.getUrl() + "|" + u.getTooltip() + "|" + u.getLabel() + "}";
	}

	static String stereo(Stereotype s) {
		return s == null ? "null" : s.toString();
	}
}
