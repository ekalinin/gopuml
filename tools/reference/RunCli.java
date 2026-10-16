// SPDX-License-Identifier: MIT
// Provenance: original

import net.sourceforge.plantuml.Run;
import net.sourceforge.plantuml.error.PSystemError;

/**
 * Runs the PlantUML command line with the time-based decorations of error
 * images disabled. Upstream adds sponsor messages to error images during some
 * minutes of each hour, which would make references depend on the clock.
 *
 * Usage: java -cp plantuml.jar:tools RunCli [plantuml arguments]
 */
public class RunCli {
	public static void main(String[] args) throws Exception {
		PSystemError.disableTimeBasedErrorDecorations();
		Run.main(args);
	}
}
