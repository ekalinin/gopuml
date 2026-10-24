// SPDX-License-Identifier: MIT
// Provenance: original
//
// Harness parses sequence diagram sources with the reference PlantUML jar
// (v1.2026.9, MIT variant) and dumps the model that upstream's
// SequenceDiagramFactory builds, or its error, in the text form that
// dump_test.go writes for the Go port. The Go tests compare both dumps.
//
// Regenerate the golden files in the reference image with:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/sequencediagram
//
// Usage: java -cp plantuml.jar:<classes> Harness <out.txt> <case files...>
//
// A case file holds cases; a line "=== name" starts a case. A source
// without an @start line is wrapped in @startuml/@enduml. Backslash-u
// escapes (four hex digits) are decoded when the case name ends with
// " (escaped)".

import java.io.PrintStream;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

import net.sourceforge.plantuml.BlockUml;
import net.sourceforge.plantuml.SourceStringReader;
import net.sourceforge.plantuml.abel.DisplayPositioned;
import net.sourceforge.plantuml.abel.EntityPortion;
import net.sourceforge.plantuml.core.Diagram;
import net.sourceforge.plantuml.core.UmlSource;
import net.sourceforge.plantuml.error.PSystemError;
import net.sourceforge.plantuml.klimt.Fashion;
import net.sourceforge.plantuml.klimt.color.HColor;
import net.sourceforge.plantuml.klimt.creole.Display;
import net.sourceforge.plantuml.klimt.sprite.Sprite;
import net.sourceforge.plantuml.nio.PathSystem;
import net.sourceforge.plantuml.preproc.PreprocessingArtifact;
import net.sourceforge.plantuml.sequencediagram.AbstractMessage;
import net.sourceforge.plantuml.sequencediagram.AutoNumber;
import net.sourceforge.plantuml.sequencediagram.Delay;
import net.sourceforge.plantuml.sequencediagram.Divider;
import net.sourceforge.plantuml.sequencediagram.Event;
import net.sourceforge.plantuml.sequencediagram.Grouping;
import net.sourceforge.plantuml.sequencediagram.GroupingLeaf;
import net.sourceforge.plantuml.sequencediagram.GroupingStart;
import net.sourceforge.plantuml.sequencediagram.HSpace;
import net.sourceforge.plantuml.sequencediagram.LifeEvent;
import net.sourceforge.plantuml.sequencediagram.LinkAnchor;
import net.sourceforge.plantuml.sequencediagram.Message;
import net.sourceforge.plantuml.sequencediagram.MessageExo;
import net.sourceforge.plantuml.sequencediagram.MessageNumber;
import net.sourceforge.plantuml.sequencediagram.Newpage;
import net.sourceforge.plantuml.sequencediagram.Note;
import net.sourceforge.plantuml.sequencediagram.Notes;
import net.sourceforge.plantuml.sequencediagram.Participant;
import net.sourceforge.plantuml.sequencediagram.ParticipantEnglober;
import net.sourceforge.plantuml.sequencediagram.Reference;
import net.sourceforge.plantuml.sequencediagram.SequenceDiagram;
import net.sourceforge.plantuml.sequencediagram.SequenceDiagramFactory;
import net.sourceforge.plantuml.skin.ArrowConfiguration;
import net.sourceforge.plantuml.skin.Pragma;
import net.sourceforge.plantuml.skin.PragmaKey;
import net.sourceforge.plantuml.stereo.Stereotype;
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
			if (c == '\\' && i + 5 < s.length() + 0 && s.charAt(i + 1) == 'u') {
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

	static final long[] NUMBERS = { 0, 1, 7, 12, 15, 25, 99, 123, 1234, 99999, 123456789, -5, 2147483647 };

	// dumpFormats formats numbers with each line of source as a
	// DecimalFormat pattern, as autonumber does.
	static void dumpFormats(List<String> out, String source) {
		for (String pattern : source.split("\n", -1)) {
			if (pattern.isEmpty())
				continue;
			final StringBuilder sb = new StringBuilder("format [" + pattern + "]");
			try {
				final java.text.DecimalFormat df = new java.text.DecimalFormat(pattern);
				for (long n : NUMBERS)
					sb.append(" | ").append(df.format(n));
			} catch (IllegalArgumentException e) {
				sb.append(" error ").append(e.getMessage());
			}
			out.add(sb.toString());
		}
	}

	static void dumpCase(List<String> out, String name, String source) {
		out.add("=== " + name);
		if (name.contains("decimalformat/")) {
			dumpFormats(out, source);
			return;
		}
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
			final Diagram d = new SequenceDiagramFactory().createSystem(PathSystem.fetch(), umlSource, null, pa);
			if (isCompact(name)) {
				final List<String> full = new ArrayList<>();
				dump(full, d);
				for (String line : full)
					if (line.startsWith("result") || line.startsWith("error") || line.startsWith("event"))
						out.add(line);
			} else {
				dump(out, d);
			}
		} catch (Throwable t) {
			out.add("crash " + t);
		}
	}

	// isCompact tells whether only the result, the error and the events of a
	// case are dumped (the generated arrow cases).
	static boolean isCompact(String name) {
		return name.startsWith("arrowcombos/") || name.startsWith("exocombos/");
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
		final SequenceDiagram sd = (SequenceDiagram) d;
		out.add("result SequenceDiagram");
		out.add("description " + sd.getDescription());
		out.add("pages " + sd.getNbImages());
		out.add("title " + dp(sd.getTitle()));
		out.add("caption " + dp(sd.getCaption()));
		out.add("legend " + dp(sd.getLegend()));
		out.add("header " + dp(sd.getHeader()));
		out.add("footer " + dp(sd.getFooter()));
		out.add("mainframe " + dp(sd.getMainFrame()));
		for (int i = 1; i < sd.getNbImages(); i++)
			out.add("pagetitle " + i + " " + dp(sd.getTitle(i)));
		out.add("scale " + scale(sd.getScale()));
		out.add("splitpages " + sd.getSplitPagesHorizontal() + " " + sd.getSplitPagesVertical());
		out.add("namespaceseparator " + sd.getNamespaceSeparator());
		out.add("skinparamused " + sd.isSkinParamUsed());
		out.add("smetana " + sd.isUseSmetana() + " elk " + sd.isUseElk());
		out.add("footbox " + sd.isShowFootbox());
		out.add("hideunlinked " + sd.isHideUnlinkedData());
		out.add("autoactivate " + sd.isAutoactivate());
		out.add("hidden " + field(sd, "hiddenPortions"));
		final AutoNumber an = sd.getAutoNumber();
		out.add("autonumber running=" + field(an, "running") + " current=" + field(an, "current") + " increment="
				+ field(an, "increment") + " last=" + field(an, "last"));
		final Pragma pragma = sd.getPragma();
		for (PragmaKey k : PragmaKey.values())
			if (pragma.isDefine(k))
				out.add("pragma " + k + "=" + pragma.getValue(k));
		for (Warning w : pragma.getWarnings())
			out.add("warning " + String.join("\\n", w.getMessage()));
		final Map<String, String> sp = new TreeMap<>(sd.getSkinParam().values());
		for (Map.Entry<String, String> ent : sp.entrySet())
			out.add("skinparam " + ent.getKey() + "=" + ent.getValue());
		final Collection<String> sprites = new java.util.TreeSet<>(
				((net.sourceforge.plantuml.skin.SkinParam) sd.getSkinParam()).getAllSpriteNames());
		for (String s : sprites)
			out.add("sprite " + s + " " + sprite(sd.getSkinParam().getSprite(s)));
		out.add("defaultskin " + ((net.sourceforge.plantuml.skin.SkinParam) sd.getSkinParam()).getDefaultSkin());
		for (Participant p : sd.participants())
			out.add("participant " + participant(sd, p));
		final List<Event> events = sd.events();
		for (int i = 0; i < events.size(); i++)
			out.add("event " + i + " " + event(events, events.get(i)));
		for (LinkAnchor a : sd.getLinkAnchors())
			out.add("linkanchor " + a.getAnchor1() + " " + a.getAnchor2() + " " + a.getMessage());
	}

	static String sprite(Sprite s) {
		if (s == null)
			return "null";
		final String cls = s.getClass().getSimpleName();
		if (cls.equals("SpriteMonochrome")) {
			final int[][] gray = (int[][]) field(s, "gray");
			final StringBuilder sb = new StringBuilder();
			for (int[] row : gray)
				for (int v : row)
					sb.append(Integer.toHexString(v));
			return cls + " " + field(s, "width") + "x" + field(s, "height") + "/" + field(s, "grayLevel") + " " + sb;
		}
		if (cls.equals("SpriteColor")) {
			final int[][] color = (int[][]) field(s, "color");
			final StringBuilder sb = new StringBuilder();
			for (int[] row : color)
				for (int v : row)
					sb.append(Integer.toHexString(v)).append(',');
			return cls + " " + field(s, "width") + "x" + field(s, "height") + " " + sb;
		}
		if (cls.equals("SpriteImage"))
			return cls;
		return "svg";
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

	static String fashion(Fashion f) {
		if (f == null)
			return "null";
		return color(f.getBackColor()) + "/" + color(f.getForeColor());
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
			else if (cs instanceof MessageNumber)
				sb.append("N:");
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

	static String participant(SequenceDiagram sd, Participant p) {
		final StringBuilder sb = new StringBuilder();
		sb.append(p.getCode()).append(" uid=").append(p.getUid()).append(" type=").append(p.getType());
		sb.append(" order=").append(p.getOrder());
		sb.append(" display=").append(display(p.getDisplay(false)));
		sb.append(" underlined=").append(display(p.getDisplay(true)));
		sb.append(" stereo=").append(stereo(p.getStereotype()));
		sb.append(" stereotop=").append(field(p, "stereotypePositionTop"));
		sb.append(" colors=").append(p.getColors());
		sb.append(" url=").append(url(p.getUrl()));
		sb.append(" initiallife=").append(p.getInitialLife());
		@SuppressWarnings("unchecked")
		final List<Fashion> lives = (List<Fashion>) field(p, "liveBackcolors");
		sb.append(" lives=[");
		for (Fashion f : lives)
			sb.append(fashion(f)).append(";");
		sb.append("]");
		final ParticipantEnglober e = sd.getEnglober(p);
		sb.append(" box=");
		if (e == null)
			sb.append("null");
		else
			for (ParticipantEnglober g : e.getGenealogy())
				sb.append("<").append(display(g.getTitle())).append(" ").append(color(g.getBoxColor())).append(" ")
						.append(stereo(g.getStereotype())).append(">");
		sb.append(" at=").append(p.getLocation() == null ? "null" : "" + p.getLocation().getPosition());
		return sb.toString();
	}

	static String arrow(ArrowConfiguration a) {
		return a.name() + " parts=" + a.getDressing1().getPart() + "/" + a.getDressing2().getPart() + " thickness="
				+ field(a, "thickness") + " reverse=" + a.isReverseDefine() + " inclination=" + field(a, "inclination");
	}

	static String note(Note n) {
		final StringBuilder sb = new StringBuilder("Note");
		sb.append(" p=").append(n.getParticipant() == null ? "null" : n.getParticipant().getCode());
		sb.append(" p2=").append(n.getParticipant2() == null ? "null" : n.getParticipant2().getCode());
		sb.append(" pos=").append(n.getPosition()).append(" style=").append(n.getNoteStyle());
		sb.append(" display=").append(display(n.getDisplay()));
		sb.append(" colors=").append(n.getColors());
		sb.append(" shadowing=").append(n.getColors().getShadowing());
		sb.append(" url=").append(url(n.getUrl()));
		sb.append(" stereo=").append(stereo((Stereotype) field(n, "stereotype")));
		sb.append(" parallel=").append(n.isParallel());
		return sb.toString();
	}

	static String notes(List<Note> notes) {
		final StringBuilder sb = new StringBuilder("[");
		for (Note n : notes)
			sb.append("{").append(note(n)).append("}");
		return sb.append("]").toString();
	}

	static String event(List<Event> events, Event e) {
		final StringBuilder sb = new StringBuilder(e.getClass().getSimpleName());
		if (e instanceof AbstractMessage) {
			final AbstractMessage m = (AbstractMessage) e;
			if (m instanceof Message) {
				final Message mm = (Message) m;
				sb.append(" p1=").append(mm.getParticipant1().getCode()).append(" p2=")
						.append(mm.getParticipant2().getCode());
				sb.append(" multicast=[");
				for (Participant p : mm.getMulticast())
					sb.append(p.getCode()).append(";");
				sb.append("]");
			} else {
				final MessageExo mx = (MessageExo) m;
				sb.append(" p=").append(mx.getParticipant().getCode()).append(" type=").append(mx.getType())
						.append(" short=").append(mx.isShortArrow());
			}
			sb.append(" label=").append(display(m.getLabel()));
			sb.append(" numbered=").append(display(m.getLabelNumbered()));
			sb.append(" number=").append(m.getMessageNumber());
			sb.append(" arrow=").append(arrow(m.getArrowConfiguration()));
			sb.append(" parallel=").append(m.isParallel());
			final Object brother = field(m, "parallelBrother");
			sb.append(" brother=").append(brother == null ? -1 : events.indexOf(brother));
			sb.append(" life=").append(field(m, "lifeEventsType"));
			sb.append(" first=").append(field(m, "firstIsActivate"));
			sb.append(" notes=").append(notes(m.getNoteOnMessages()));
			sb.append(" url=").append(url((Url) field(m, "url")));
			sb.append(" hasurl=").append(m.hasUrl());
			sb.append(" stereo=").append(stereo((Stereotype) field(m, "stereotype")));
			sb.append(" anchors=").append(m.getAnchor()).append("/").append(m.getPart1Anchor()).append("/")
					.append(m.getPart2Anchor());
			sb.append(" at=").append(m.getLineLocation() == null ? "null" : "" + m.getLineLocation().getPosition());
		} else if (e instanceof LifeEvent) {
			final LifeEvent l = (LifeEvent) e;
			sb.append(" p=").append(l.getParticipant().getCode()).append(" type=").append(l.getType());
			sb.append(" colors=").append(fashion(l.getSpecificColors()));
			sb.append(" message=").append(l.getMessage() == null ? -1 : events.indexOf(l.getMessage()));
		} else if (e instanceof Note) {
			sb.append(" ").append(note((Note) e));
		} else if (e instanceof Notes) {
			sb.append(" ").append(notes(((Notes) e).asList()));
		} else if (e instanceof Grouping) {
			final Grouping g = (Grouping) e;
			sb.append(" type=").append(g.getType()).append(" title=").append(g.getTitle()).append(" comment=")
					.append(g.getComment());
			sb.append(" general=").append(color(g.getBackColorGeneral())).append(" element=")
					.append(color(g.getBackColorElement()));
			sb.append(" level=").append(g.getLevel()).append(" parallel=").append(g.isParallel());
			if (g instanceof GroupingLeaf) {
				final GroupingLeaf gl = (GroupingLeaf) g;
				sb.append(" start=").append(events.indexOf(gl.getGroupingStart()));
				sb.append(" notes=").append(notes(gl.getNoteOnMessages()));
			} else {
				final Object parent = field(g, "parent");
				sb.append(" parent=").append(parent == null ? -1 : events.indexOf(parent));
			}
		} else if (e instanceof Divider) {
			sb.append(" text=").append(display(((Divider) e).getText()));
		} else if (e instanceof Delay) {
			sb.append(" text=").append(display(((Delay) e).getText()));
		} else if (e instanceof HSpace) {
			sb.append(" pixel=").append(((HSpace) e).getPixel());
		} else if (e instanceof Reference) {
			final Reference r = (Reference) e;
			sb.append(" participants=[");
			for (Participant p : r.getParticipant())
				sb.append(p.getCode()).append(";");
			sb.append("] strings=").append(display(r.getStrings()));
			sb.append(" url=").append(url(r.getUrl()));
			sb.append(" general=").append(color(r.getBackColorGeneral())).append(" element=")
					.append(color(r.getBackColorElement()));
			sb.append(" notes=").append(notes(r.getNoteOnMessages()));
		} else if (e instanceof Newpage) {
		}
		return sb.toString();
	}
}
