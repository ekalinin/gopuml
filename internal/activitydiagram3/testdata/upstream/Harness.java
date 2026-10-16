// SPDX-License-Identifier: MIT
// Provenance: original
//
// Harness parses activity diagram sources with the reference PlantUML jar
// (v1.2026.8, MIT variant) and dumps the instruction tree that upstream's
// ActivityDiagramFactory3 builds, or its error, in the text form that
// dump_test.go writes for the Go port. The Go tests compare both dumps.
//
// Regenerate the golden files in the reference image with:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/activitydiagram3
//
// Usage: java -cp plantuml.jar:<classes> Harness <out.txt> <case files...>
//
// A case file holds cases; a line "=== name" starts a case. A source
// without an @start line is wrapped in @startuml/@enduml. Backslash-u
// escapes (four hex digits) are decoded when the case name ends with
// " (escaped)". The generated cases ("combos/") only dump the result, the
// errors, the warnings and the instruction tree. For the cases whose name
// starts with "vega/" or "allconf/", the whole
// PSystemBuilder runs first: a source that upstream does not turn into an
// activity diagram is only reported as "not activity3".

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

import net.sourceforge.plantuml.BlockUml;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.abel.DisplayPositioned;
import net.sourceforge.plantuml.activitydiagram3.ActivityDiagram3;
import net.sourceforge.plantuml.activitydiagram3.ActivityDiagramFactory3;
import net.sourceforge.plantuml.activitydiagram3.Branch;
import net.sourceforge.plantuml.activitydiagram3.Instruction;
import net.sourceforge.plantuml.activitydiagram3.InstructionBreak;
import net.sourceforge.plantuml.activitydiagram3.InstructionEnd;
import net.sourceforge.plantuml.activitydiagram3.InstructionFork;
import net.sourceforge.plantuml.activitydiagram3.InstructionGoto;
import net.sourceforge.plantuml.activitydiagram3.InstructionGroup;
import net.sourceforge.plantuml.activitydiagram3.InstructionIf;
import net.sourceforge.plantuml.activitydiagram3.InstructionLabel;
import net.sourceforge.plantuml.activitydiagram3.InstructionList;
import net.sourceforge.plantuml.activitydiagram3.InstructionRepeat;
import net.sourceforge.plantuml.activitydiagram3.InstructionSimple;
import net.sourceforge.plantuml.activitydiagram3.InstructionSplit;
import net.sourceforge.plantuml.activitydiagram3.InstructionSpot;
import net.sourceforge.plantuml.activitydiagram3.InstructionStart;
import net.sourceforge.plantuml.activitydiagram3.InstructionStop;
import net.sourceforge.plantuml.activitydiagram3.InstructionSwitch;
import net.sourceforge.plantuml.activitydiagram3.InstructionWhile;
import net.sourceforge.plantuml.activitydiagram3.LinkRendering;
import net.sourceforge.plantuml.activitydiagram3.PositionedNote;
import net.sourceforge.plantuml.activitydiagram3.ftile.BoxStyle;
import net.sourceforge.plantuml.activitydiagram3.ftile.Swimlane;
import net.sourceforge.plantuml.activitydiagram3.ftile.Swimlanes;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.core.UmlSource;
import net.sourceforge.plantuml.decoration.HtmlColorAndStyle;
import net.sourceforge.plantuml.decoration.Rainbow;
import net.sourceforge.plantuml.decoration.symbol.USymbol;
import net.sourceforge.plantuml.decoration.symbol.USymbols;
import net.sourceforge.plantuml.error.PSystemError;
import net.sourceforge.plantuml.klimt.color.Colors;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.color.HColorSet;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.PreprocessingArtifact;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.PragmaKey;
import net.sourceforge.plantuml.stereo.Stereogroup;
import net.sourceforge.plantuml.stereo.Stereotype;
import net.sourceforge.plantuml.style.PName;
import net.sourceforge.plantuml.style.Style;
import net.sourceforge.plantuml.text.StringLocated;
import net.sourceforge.plantuml.url.Url;
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
			if (name.startsWith("vega/") || name.startsWith("allconf/")) {
				final Diagram full = block.getDiagram();
				if (full instanceof ActivityDiagram3 == false) {
					out.add("not activity3");
					return;
				}
			}
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
			final Diagram d = new ActivityDiagramFactory3().createSystem(PathSystem.fetch(), umlSource, null, pa);
			if (isCompact(name)) {
				final List<String> full = new ArrayList<>();
				dump(full, d);
				for (String line : full)
					if (keepCompact(line))
						out.add(line);
			} else {
				dump(out, d);
			}
		} catch (Throwable t) {
			out.add("crash " + t);
		}
	}

	// isCompact tells whether only the result, the errors, the warnings and
	// the instruction tree of a case are dumped (the generated cases).
	static boolean isCompact(String name) {
		return name.startsWith("combos/");
	}

	static boolean keepCompact(String line) {
		return line.startsWith("result") || line.startsWith("error") || line.startsWith("warning")
				|| line.startsWith("ins ") || line.startsWith("branch ") || line.startsWith("current ");
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
		final ActivityDiagram3 ad = (ActivityDiagram3) d;
		out.add("result ActivityDiagram3");
		out.add("description " + ad.getDescription());
		out.add("pages " + ad.getNbImages());
		out.add("title " + dp(ad.getTitle()));
		out.add("caption " + dp(ad.getCaption()));
		out.add("legend " + dp(ad.getLegend()));
		out.add("header " + dp(ad.getHeader()));
		out.add("footer " + dp(ad.getFooter()));
		out.add("mainframe " + dp(ad.getMainFrame()));
		out.add("skinparamused " + ad.isSkinParamUsed());
		final Pragma pragma = ad.getPragma();
		for (PragmaKey k : PragmaKey.values())
			if (pragma.isDefine(k))
				out.add("pragma " + k + "=" + pragma.getValue(k));
		for (Warning w : pragma.getWarnings())
			out.add("warning " + String.join("\\n", w.getMessage()));
		final Map<String, String> sp = new TreeMap<>(ad.getSkinParam().values());
		for (Map.Entry<String, String> ent : sp.entrySet())
			out.add("skinparam " + ent.getKey() + "=" + ent.getValue());
		out.add("hasurl " + ad.hasUrl());
		out.add("strategy " + field(ad, "swimlaneStrategy"));
		final Swimlanes swimlanes = (Swimlanes) field(ad, "swimlanes");
		@SuppressWarnings("unchecked")
		final List<Swimlane> raw = (List<Swimlane>) field(swimlanes, "swimlanesRaw");
		for (Swimlane s : raw)
			out.add("swimlane " + s.getName() + " order=" + field(s, "order") + " display=" + display(s.getDisplay())
					+ " colors=" + s.getColors());
		out.add("currentswimlane " + lane(swimlanes.getCurrentSwimlane()));
		out.add("nextlink " + lr(swimlanes.nextLinkRenderer()));
		final Map<Object, Integer> ids = new IdentityHashMap<>();
		dumpIns(out, ids, (Instruction) field(swimlanes, "root"), 0, "root");
		final Integer current = ids.get(swimlanes.getCurrent());
		out.add("current " + (current == null ? "?" : current));
	}

	static String color(HColor c) {
		return c == null ? "null" : c.toString();
	}

	static String colors(Colors c) {
		return c == null ? "null" : c.toString();
	}

	static String display(Display d) {
		if (d == null)
			return "null";
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

	static String stereogroup(Stereogroup g) {
		if (g == null)
			return "nil";
		if (g.isEmpty())
			return "null";
		return g.getLabels().toString();
	}

	static String lane(Swimlane s) {
		return s == null ? "null" : s.getName();
	}

	static String lanes(Set<Swimlane> set) {
		final TreeSet<String> names = new TreeSet<>();
		for (Swimlane s : set)
			names.add(lane(s));
		return names.toString();
	}

	static String rainbow(Rainbow r) {
		if (r == null)
			return "null";
		final StringBuilder sb = new StringBuilder("[");
		boolean first = true;
		for (HtmlColorAndStyle c : r.getColors()) {
			if (first == false)
				sb.append(";");
			first = false;
			sb.append(color(c.getArrowColor())).append("/").append(color(c.getArrowHeadColor())).append("/")
					.append(c.getStyle());
		}
		return sb.append("]/").append(r.getColorArrowSeparationSpace()).toString();
	}

	static String lr(LinkRendering l) {
		if (l == null)
			return "null";
		return "{" + display(l.getDisplay()) + " " + rainbow(l.getRainbow()) + "}";
	}

	static String boxStyle(BoxStyle b) {
		return b == null ? "nil" : String.valueOf(b.name());
	}

	static String note(PositionedNote n) {
		if (n == null)
			return "null";
		return "{" + display(n.getDisplay()) + " " + n.getNotePosition() + " " + n.getType() + " lane="
				+ lane(n.getSwimlaneNote()) + " colors=" + colors(n.getColors()) + " stereo="
				+ stereo(n.getStereotype()) + "}";
	}

	static String notes(Collection<PositionedNote> notes) {
		final StringBuilder sb = new StringBuilder("[");
		for (PositionedNote n : notes)
			sb.append(note(n));
		return sb.append("]").toString();
	}

	static String symbol(USymbol s) {
		if (s == USymbols.PARTITION)
			return "PARTITION";
		if (s == USymbols.PACKAGE)
			return "PACKAGE";
		if (s == USymbols.RECTANGLE)
			return "RECTANGLE";
		if (s == USymbols.CARD)
			return "CARD";
		if (s == USymbols.GROUP)
			return "GROUP";
		return String.valueOf(s);
	}

	static String id(Map<Object, Integer> ids, Object o) {
		if (o == null)
			return "null";
		final Integer id = ids.get(o);
		return id == null ? "?" : "" + id;
	}

	interface Getter {
		String get();
	}

	// safe returns the value, or the class of the exception it throws.
	static String safe(Getter g) {
		try {
			return g.get();
		} catch (Throwable t) {
			return "!" + t.getClass().getSimpleName();
		}
	}

	@SuppressWarnings("unchecked")
	static void dumpIns(List<String> out, Map<Object, Integer> ids, Instruction ins, int depth, String role) {
		final int id = ids.size();
		ids.put(ins, id);
		final StringBuilder sb = new StringBuilder("ins " + id + " " + depth + " " + role + " "
				+ ins.getClass().getSimpleName());
		sb.append(" lanes=").append(safe(() -> lanes(ins.getSwimlanes())));
		sb.append(" break=").append(safe(() -> "" + ins.containsBreak()));
		sb.append(" inlink=").append(safe(() -> lr(ins.getInLinkRendering())));
		final List<Runnable> children = new ArrayList<>();
		if (ins instanceof InstructionList) {
			final InstructionList l = (InstructionList) ins;
			sb.append(" default=").append(lane((Swimlane) field(l, "defaultSwimlane")));
			sb.append(" outlink=").append(lr((LinkRendering) field(l, "outlinkRendering")));
			sb.append(" empty=").append(l.isEmpty());
			sb.append(" onlystop=").append(l.isOnlySingleStopOrSpot());
			sb.append(" notes=").append(notes(l.getPositionedNotes()));
			for (Instruction child : (List<Instruction>) field(l, "all"))
				children.add(() -> dumpIns(out, ids, child, depth + 1, "item"));
		} else if (ins instanceof InstructionSimple) {
			final InstructionSimple s = (InstructionSimple) ins;
			sb.append(" in=").append(lane(s.getSwimlaneIn())).append(" out=").append(lane(s.getSwimlaneOut()));
			sb.append(" label=").append(display((Display) field(s, "label")));
			sb.append(" colors=").append(colors((Colors) field(s, "colors")));
			sb.append(" box=").append(boxStyle((BoxStyle) field(s, "boxStyle")));
			sb.append(" url=").append(url((Url) field(s, "url")));
			sb.append(" stereo=").append(stereo((Stereotype) field(s, "stereotype")));
			sb.append(" killed=").append(s.isKilled());
			sb.append(" notes=").append(notes(s.getPositionedNotes()));
		} else if (ins instanceof InstructionSpot) {
			final InstructionSpot s = (InstructionSpot) ins;
			sb.append(" in=").append(lane(s.getSwimlaneIn())).append(" out=").append(lane(s.getSwimlaneOut()));
			sb.append(" spot=").append(field(s, "spot"));
			sb.append(" color=").append(color((HColor) field(s, "color")));
			sb.append(" killed=").append(field(s, "killed"));
			sb.append(" notes=").append(notes(s.getPositionedNotes()));
		} else if (ins instanceof InstructionGoto || ins instanceof InstructionLabel) {
			sb.append(" in=").append(lane(ins.getSwimlaneIn())).append(" out=").append(lane(ins.getSwimlaneOut()));
			sb.append(" name=").append(field(ins, "name"));
			sb.append(" notes=").append(notes((Collection<PositionedNote>) field(ins, "notes")));
		} else if (ins instanceof InstructionStart || ins instanceof InstructionStop
				|| ins instanceof InstructionEnd) {
			sb.append(" in=").append(lane(ins.getSwimlaneIn())).append(" out=").append(lane(ins.getSwimlaneOut()));
			sb.append(" colors=").append(colors((Colors) field(ins, "colors")));
			sb.append(" notes=").append(notes((Collection<PositionedNote>) field(ins, "notes")));
		} else if (ins instanceof InstructionBreak) {
			sb.append(" in=").append(lane(ins.getSwimlaneIn())).append(" out=").append(lane(ins.getSwimlaneOut()));
			sb.append(" notes=").append(notes((Collection<PositionedNote>) field(ins, "notes")));
		} else if (ins instanceof InstructionIf) {
			final InstructionIf i = (InstructionIf) ins;
			sb.append(" parent=").append(id(ids, i.getParent()));
			sb.append(" lane=").append(lane((Swimlane) field(i, "swimlane")));
			sb.append(" url=").append(url((Url) field(i, "url")));
			sb.append(" stereo=").append(stereo((Stereotype) field(i, "stereotype")));
			sb.append(" outcolor=").append(lr((LinkRendering) field(i, "outColor")));
			sb.append(" endif=").append(field(i, "endifCalled"));
			final List<Branch> thens = (List<Branch>) field(i, "thens");
			final Branch elseBranch = (Branch) field(i, "elseBranch");
			final Branch current = (Branch) field(i, "current");
			sb.append(" current=").append(current == elseBranch ? "else" : "" + thens.indexOf(current));
			sb.append(" notes=").append(notes(i.getPositionedNotes()));
			for (int k = 0; k < thens.size(); k++) {
				final Branch b = thens.get(k);
				final String r = "then" + k;
				children.add(() -> dumpBranch(out, ids, b, depth + 1, r));
			}
			if (elseBranch != null)
				children.add(() -> dumpBranch(out, ids, elseBranch, depth + 1, "else"));
		} else if (ins instanceof InstructionSwitch) {
			final InstructionSwitch s = (InstructionSwitch) ins;
			sb.append(" parent=").append(id(ids, s.getParent()));
			sb.append(" lane=").append(lane((Swimlane) field(s, "swimlane")));
			sb.append(" test=").append(display((Display) field(s, "labelTest")));
			sb.append(" colors=").append(colors((Colors) field(s, "colors")));
			sb.append(" endcolors=").append(colors((Colors) field(s, "endColors")));
			sb.append(" after=").append(lr((LinkRendering) field(s, "afterEndwhile")));
			final List<Branch> switches = (List<Branch>) field(s, "switches");
			sb.append(" current=").append(switches.indexOf(field(s, "current")));
			sb.append(" notes=").append(notes(s.getPositionedNotes()));
			for (int k = 0; k < switches.size(); k++) {
				final Branch b = switches.get(k);
				final String r = "case" + k;
				children.add(() -> dumpBranch(out, ids, b, depth + 1, r));
			}
		} else if (ins instanceof InstructionWhile) {
			final InstructionWhile w = (InstructionWhile) ins;
			sb.append(" parent=").append(id(ids, w.getParent()));
			sb.append(" lane=").append(lane((Swimlane) field(w, "swimlane")));
			sb.append(" test=").append(display((Display) field(w, "test")));
			sb.append(" yes=").append(display((Display) field(w, "yes")));
			sb.append(" color=").append(color((HColor) field(w, "color")));
			sb.append(" killed=").append(field(w, "killed"));
			sb.append(" testcalled=").append(field(w, "testCalled"));
			sb.append(" outcolor=").append(lr((LinkRendering) field(w, "outColor")));
			sb.append(" box=").append(boxStyle((BoxStyle) field(w, "boxStyle")));
			sb.append(" backward=").append(display((Display) field(w, "backward")));
			sb.append(" stereo=").append(stereo((Stereotype) field(w, "stereotype")));
			sb.append(" in1=").append(lr((LinkRendering) field(w, "incoming1")));
			sb.append(" in2=").append(lr((LinkRendering) field(w, "incoming2")));
			sb.append(" backwardcalled=").append(field(w, "backwardCalled"));
			sb.append(" notes=").append(notes(w.getPositionedNotes()));
			final Instruction special = (Instruction) field(w, "specialOut");
			if (special != null)
				children.add(() -> dumpIns(out, ids, special, depth + 1, "special"));
			children.add(() -> dumpIns(out, ids, (Instruction) field(w, "repeatList"), depth + 1, "loop"));
		} else if (ins instanceof InstructionRepeat) {
			final InstructionRepeat r = (InstructionRepeat) ins;
			sb.append(" parent=").append(id(ids, r.getParent()));
			sb.append(" lane=").append(lane((Swimlane) field(r, "swimlane")));
			sb.append(" laneout=").append(lane((Swimlane) field(r, "swimlaneOut")));
			sb.append(" lanebackward=").append(lane((Swimlane) field(r, "swimlaneBackward")));
			sb.append(" start=").append(display((Display) field(r, "startLabel")));
			sb.append(" boxin=").append(boxStyle((BoxStyle) field(r, "boxStyleIn")));
			sb.append(" sgloop=").append(stereogroup((Stereogroup) field(r, "stereogroupLoop")));
			sb.append(" test=").append(display((Display) field(r, "test")));
			sb.append(" yes=").append(display((Display) field(r, "yes")));
			sb.append(" out=").append(display((Display) field(r, "out")));
			sb.append(" sg2=").append(stereogroup((Stereogroup) field(r, "stereotype2")));
			sb.append(" testcalled=").append(field(r, "testCalled"));
			sb.append(" killed=").append(field(r, "killed"));
			sb.append(" endlink=").append(lr((LinkRendering) field(r, "endRepeatLinkRendering")));
			sb.append(" backward=").append(display((Display) field(r, "backward")));
			sb.append(" hasbackward=").append(r.hasBackward());
			sb.append(" box=").append(boxStyle((BoxStyle) field(r, "boxStyle")));
			sb.append(" stereoback=").append(stereo((Stereotype) field(r, "stereotypeBack")));
			sb.append(" in1=").append(lr((LinkRendering) field(r, "incoming1")));
			sb.append(" in2=").append(lr((LinkRendering) field(r, "incoming2")));
			sb.append(" backnotes=").append(notes((List<PositionedNote>) field(r, "backwardNotes")));
			children.add(() -> dumpIns(out, ids, (Instruction) field(r, "repeatList"), depth + 1, "loop"));
		} else if (ins instanceof InstructionFork) {
			final InstructionFork f = (InstructionFork) ins;
			sb.append(" parent=").append(id(ids, f.getParent()));
			sb.append(" in=").append(lane(f.getSwimlaneIn())).append(" out=").append(lane(f.getSwimlaneOut()));
			sb.append(" style=").append(field(f, "style"));
			sb.append(" label=").append(field(f, "label"));
			sb.append(" finished=").append(field(f, "finished"));
			sb.append(" colors=").append(colors((Colors) field(f, "colors")));
			sb.append(" notes=").append(notes(f.getPositionedNotes()));
			final List<InstructionList> forks = (List<InstructionList>) field(f, "forks");
			for (int k = 0; k < forks.size(); k++) {
				final InstructionList l = forks.get(k);
				final String r = "fork" + k;
				children.add(() -> dumpIns(out, ids, l, depth + 1, r));
			}
		} else if (ins instanceof InstructionSplit) {
			final InstructionSplit s = (InstructionSplit) ins;
			sb.append(" parent=").append(id(ids, s.getParent()));
			sb.append(" in=").append(lane((Swimlane) field(s, "swimlaneIn")));
			sb.append(" out=").append(lane(s.getSwimlaneOut()));
			sb.append(" colors=").append(colors((Colors) field(s, "colors")));
			final List<InstructionList> splits = (List<InstructionList>) field(s, "splits");
			for (int k = 0; k < splits.size(); k++) {
				final InstructionList l = splits.get(k);
				final String r = "split" + k;
				children.add(() -> dumpIns(out, ids, l, depth + 1, r));
			}
		} else if (ins instanceof InstructionGroup) {
			final InstructionGroup g = (InstructionGroup) ins;
			sb.append(" parent=").append(id(ids, g.getParent()));
			sb.append(" title=").append(display((Display) field(g, "title")));
			sb.append(" type=").append(symbol((USymbol) field(g, "type")));
			sb.append(" back=").append(color((HColor) field(g, "backColor")));
			final Style style = (Style) field(g, "style");
			sb.append(" styleline=").append(color(style.value(PName.LineColor).asColor(HColorSet.instance())));
			sb.append(" styleback=").append(color(style.value(PName.BackGroundColor).asColor(HColorSet.instance())));
			sb.append(" note=").append(note((PositionedNote) field(g, "note")));
			children.add(() -> dumpIns(out, ids, (Instruction) field(g, "list"), depth + 1, "body"));
		}
		out.add(sb.toString());
		for (Runnable r : children)
			r.run();
	}

	static void dumpBranch(List<String> out, Map<Object, Integer> ids, Branch b, int depth, String role) {
		final StringBuilder sb = new StringBuilder("branch " + depth + " " + role);
		sb.append(" test=").append(display(b.getLabelTest()));
		sb.append(" positive=").append(lr((LinkRendering) field(b, "labelPositive")));
		sb.append(" inlink=").append(lr((LinkRendering) field(b, "inlinkRendering")));
		sb.append(" inlabel=").append(lr((LinkRendering) field(b, "inlabel")));
		sb.append(" special=").append(lr(b.getSpecial()));
		sb.append(" color=").append(color((HColor) field(b, "color")));
		sb.append(" specialcolors=").append(colors((Colors) field(b, "specialColors")));
		sb.append(" getcolor=").append(color(b.getColor()));
		out.add(sb.toString());
		dumpIns(out, ids, (Instruction) field(b, "list"), depth + 1, "body");
	}
}
