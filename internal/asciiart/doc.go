// SPDX-License-Identifier: MIT
// Provenance: original

// Package asciiart is upstream's asciiart package, for the ASCII (TXT) and
// Unicode (UTXT) text art. Upstream 1.2026.9 draws sequence diagrams in
// text through internal/asciiverse; only the text measurer
// (TextStringBounder) of this package is used for them, so only it is
// ported. The legacy text components (TextSkin, ComponentText*, the
// character areas) and Wcwidth are not used by the sequence text output.
package asciiart
