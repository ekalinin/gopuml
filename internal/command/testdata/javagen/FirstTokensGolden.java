// SPDX-License-Identifier: MIT
// Provenance: original
//
// FirstTokensGolden writes the golden file of the first-token tests
// (internal/command/regex and internal/engine): for each factory of the
// diagram types this implementation supports, its commands in registration
// order, with the class, the starting pattern (when the command has one),
// the first tokens the command declares (mandatoryFirstTokensFast) and those
// FirstTokens reads from the pattern (mandatoryFirstTokensSlow). From the
// repository root:
//
//   mkdir -p ~/scratch-firsttokens && javac --release 21 -cp .cache/upstream/plantuml-mit-1.2026.9.jar \
//     -d ~/scratch-firsttokens internal/command/testdata/javagen/FirstTokensGolden.java
//   docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" -v "$PWD:/repo" \
//     -v "$HOME/scratch-firsttokens:/classes" plantuml-go-reference \
//     java -Djava.awt.headless=true -cp /opt/plantuml.jar:/classes FirstTokensGolden \
//     /repo/internal/command/testdata/first_tokens.golden
//
// Each line is: factory, index, class, pattern (or "-"), fast tokens, slow
// tokens and the tokens FirstTokens.from reads from the pattern ("-" without
// pattern), separated by tabulations. Tokens are sorted and joined with
// spaces, "ANYTHING" stands for null, and "<empty>" for the empty token.

import java.io.IOException;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.TreeSet;

import net.sourceforge.plantuml.command.Command;
import net.sourceforge.plantuml.command.PSystemCommandFactory;
import net.sourceforge.plantuml.regex.FirstTokens;
import net.sourceforge.plantuml.regex.IRegex;
import net.sourceforge.plantuml.regex.Pattern2;

public class FirstTokensGolden {

	static final String[] FACTORIES = { "net.sourceforge.plantuml.sequencediagram.SequenceDiagramFactory",
			"net.sourceforge.plantuml.classdiagram.ClassDiagramFactory",
			"net.sourceforge.plantuml.activitydiagram.ActivityDiagramFactory",
			"net.sourceforge.plantuml.descdiagram.DescriptionDiagramFactory",
			"net.sourceforge.plantuml.statediagram.StateDiagramFactory",
			"net.sourceforge.plantuml.activitydiagram3.ActivityDiagramFactory3",
			"net.sourceforge.plantuml.mindmap.MindMapDiagramFactory",
			"net.sourceforge.plantuml.packetdiag.PacketDiagramFactory",
			"net.sourceforge.plantuml.gantt.GanttDiagramFactory" };

	public static void main(String[] args) throws Exception {
		final List<String> out = new ArrayList<>();
		for (String name : FACTORIES) {
			final Class<?> type = Class.forName(name);
			final PSystemCommandFactory factory = (PSystemCommandFactory) type.getDeclaredConstructor()
					.newInstance();
			final List<Command> cmds = new ArrayList<>();
			final Method init = findMethod(type, "initCommandsList");
			init.setAccessible(true);
			init.invoke(factory, cmds);
			for (int i = 0; i < cmds.size(); i++) {
				final Command cmd = cmds.get(i);
				String slow;
				try {
					slow = tokens(cmd.mandatoryFirstTokensSlow());
				} catch (RuntimeException e) {
					slow = "ERR " + e.getClass().getSimpleName();
				}
				final String pattern = pattern(cmd);
				final String read = pattern.equals("-") ? "-" : tokens(FirstTokens.from(pattern));
				out.add(type.getSimpleName() + "\t" + i + "\t" + cmd.getClass().getName() + "\t" + escape(pattern)
						+ "\t" + tokens(cmd.mandatoryFirstTokensFast()) + "\t" + slow + "\t" + read);
			}
		}
		Files.write(Paths.get(args[0]), out, StandardCharsets.UTF_8);
	}

	static Method findMethod(Class<?> type, String name) {
		for (Class<?> c = type; c != null; c = c.getSuperclass())
			for (Method m : c.getDeclaredMethods())
				if (m.getName().equals(name))
					return m;
		throw new IllegalArgumentException(name);
	}

	// The starting pattern of a command: its "pattern" or "starting" field, or
	// that of the command it decorates.
	static String pattern(Object cmd) throws IllegalAccessException {
		for (Class<?> c = cmd.getClass(); c != null; c = c.getSuperclass())
			for (Field f : c.getDeclaredFields()) {
				final String n = f.getName();
				if (n.equals("pattern") == false && n.equals("starting") == false && n.equals("cmd") == false)
					continue;
				f.setAccessible(true);
				final Object v = f.get(cmd);
				if (v instanceof IRegex)
					return ((IRegex) v).getPatternAsString();
				if (v instanceof Pattern2)
					return ((Pattern2) v).pattern();
				if (v instanceof Command)
					return pattern(v);
			}
		return "-";
	}

	static String tokens(Collection<String> tokens) {
		if (tokens == null)
			return "ANYTHING";
		final List<String> result = new ArrayList<>();
		for (String t : new TreeSet<>(tokens))
			result.add(t.isEmpty() ? "<empty>" : escape(t));
		return String.join(" ", result);
	}

	static String escape(String s) {
		final StringBuilder sb = new StringBuilder();
		for (char c : s.toCharArray())
			if (c == '\t' || c == '\n' || c == '\r' || c == '\\' || c > 126 || c < 32)
				sb.append(String.format("\\u%04X", (int) c));
			else
				sb.append(c);
		return sb.toString();
	}
}
