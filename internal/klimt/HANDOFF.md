# internal/klimt handoff

State of the `internal/klimt` package after the drawing model and style system step (tasks 6.1, 6.5 partly, 7.1, 7.2), the creole step (task 7.4) and the skin components step. It lists what is ported, what is a stub, and what the next steps need to know.

## Ported

One Go package for upstream's cyclic packages (design D2); file names carry the origin package as prefix.

- `klimt` root: UShape/UChange (any-typed, see below), UParam, UParamNull, UStroke, UTranslate, UClip, UPattern, UHidden, UGroup/UGroupType, UPath, UMotif, UAntiAliasing, Shadowable/AbstractShadowable, Fashion, SvgAttributes, CopyForegroundColorToBackgroundColor, LineBreakStrategy, ClipContainer, AffineTransformType, Arrows (with ArrowsRegular/ArrowsTriangle from activitydiagram3.ftile), UShapeSized, MutableImage, PortableImage (interface without the java.awt accessors).
- `klimt/shape`: every shape and text block of the directory except the java.awt and svek.image dependent parts noted below.
- `klimt/drawing` core: UGraphic, UDriver, AbstractCommonUGraphic, AbstractUGraphic, AbstractUGraphicHorizontalLine, UGraphicDelegator, UGraphicFilter, UGraphicNo, UGraphicNull, UGraphicStencil, UGraphicDispatchDrawable, LimitFinder, TextLimitFinder.
- `url`: Url (as `URL`), UrlBuilder, UrlMode, CMapData.
- `stereo`: Stereotype, StereotypeDecoration, Stereogroup, Stereostyles, Stereotag, StereotypePattern (regex part).
- `style` and `style/parser`: SName, PName, Style, StyleBuilder, StyleStorage, StyleKey, StyleSignature/StyleSignatureBasic/StyleSignatures, Value/ValueImpl/ValueColor/ValueNull, DarkString, StyleLoader, FromSkinparamToStyle, ISkinParam, ISkinSimple, ClockwiseTopRightBottomLeft, LengthAdjust, MergeStrategy, StyleScheme, StringTrie, AutomaticCounter, Styleable/WithStyle, NoStyleAvailableException, the three style commands, the CSS-like parser (StyleParser, Context, CssVariables, tokens).
- `skin`: Pragma, PragmaKey, ColorParam, LineParam, AlignmentParam, CornerParam, PaddingParam, SplitParam, ComponentStyle, ActorStyle, ArrowDirection, Padder, VisibilityModifier, SkinParam, SkinParamDelegator and its five subclasses, SkinParamUtils (with Rose.getHtmlColor inlined).
- `klimt/font`: FontConfiguration, FontParam (`font_` prefix).
- `klimt/creole`: Display, CreoleMode, Stencil, Parser (static helpers), and the creole step (see "Creole" below): Sheet, SheetBuilder, SheetBlock1, SheetBlock2, Fission, Sea, Position, Neutron, Stripe, StripeStyle, CreoleContext, CreoleHorizontalLine; `creole/atom` (every atom but AtomEmoji, AtomOpenIconic and AtomMath); `creole/command` (every registered command, Splitter's patterns, AddStyle); `creole/legacy` (CreoleParser, CreoleStripeSimpleParser, StripeSimple, StripeTable, StripeTree, StripeCode, StripeLatex, StripeRaw, AtomText, AtomTextUtils, CommandCreoleBuilder).
- `com.plantuml.ubrex` (UBrex engine, `ubrex*.go`, unexported): the creole style and superscript commands use it. The `builder` subpackage is not ported (unused).
- `skin` components: Component, AbstractComponent, AbstractTextualComponent, ArrowComponent, Area, Context2D, SimpleContext2D, ComponentType, ArrowConfiguration, ArrowBody, ArrowDecoration, ArrowDressing, ArrowHead, ArrowPart, ActorStickMan, ActorAwesome, ActorHollow, CircleInterface, and `ActorStyle.TextBlock` (upstream's getTextBlock).
- `cucadiagram`: BodyFactory.create3, BodyEnhanced2 and BodyEnhancedAbstract (`cucadiagram_bodyenhanced.go`), which AbstractTextualComponent uses for enhanced displays (notes).
- Moved small helpers: utils CharHidder, salt Skeleton2 (AtomTree), EmbeddedDiagram.getEmbeddedType and the line skipping of createAndSkip, utils Base64Coder.decode (with its error messages).
- Moved small types, each with a comment naming its upstream package: svek Ports, PortGeometry, WithPorts, Margins, ShapeType, IEntityImage, PackageStyle, ConditionStyle, ConditionEndStyle, GuideLine; abel Hideable; dot DotSplines; tikz TikzFontDistortion; utils Direction; geom MagneticBorder/MagneticBorderNone; awt XAffineTransform; text Guillemet; jaws constants; warning Warning, WarningHandler, JawsWarning; preproc ConfigurationStore, OptionKey; net.atmp SpecialText and PixelImage; activitydiagram3.ftile CenteredText; compress CompressionMode and UShapeIgnorableForCompression; StringUtils and SignatureUtils helpers.
- Skins: `skin/*.skin` (plantuml, rose, debug, sonyxperiadev, strictuml) are embedded.

## Stubs and missing parts

Every stub panics with `*color.RuntimeError` whose message starts with "not ported yet:" and names the upstream class, or returns an explicit error.

### Creole

Ported and checked against the reference jar (see "Tests and golden files"). What is left:

- Stubs that panic with "not ported yet": `<math>` and `<latex>` (inline and `<latex>` blocks; AtomMath, ScientificEquationSafe), `<qrcode:...>` (AtomImg.createQrcode, FlashCodeFactory), `<img>` of a file, a URL or the standard library (`data:image/png;spm`): only data URLs (`data:image/png;base64,` and `data:image/svg+xml;base64,`) and the preprocessor's MD5 replacement (`data:image/png;md5,`, through `SpriteContainer.FromMd5`) are decoded. Raster data is decoded with Go's image/png, image/jpeg and image/gif (Java's ImageIO also reads BMP, WBMP and TIFF); decoding errors give "ERROR javax.imageio.IIOException: <Go message>" where Java's message differs.
- Embedded diagrams (`{{` ... `}}`): the lines are collected as upstream (nested `{{`, `@start<type>`/`@end<type>`), then the ISkinSimple of the text must implement `EmbeddedDiagramMaker` (the engine builds and exports the diagram); otherwise it panics with "not ported yet: EmbeddedDiagram". Nothing implements it yet; SkinParam needs a way to receive the maker from the engine.
- Icons, emoji and sprites, per the styling spec ("render as their literal text until the follow-up change"): the commands are registered as upstream, and `StripeSimple.AddOpenIcon`, `AddEmoji` and `AddSprite` add the matched markup as a text atom (a separate text run, so `<&check> done` draws "<&check>" and " done"). In link labels, `createAtomTextForURL` does the same. A known sprite (none can be defined yet) gives AtomSprite as upstream. The follow-up must restore upstream: OpenIconic and emoji drawings, nothing for an unknown sprite or icon, "¿name?" in red for an unknown emoji, an emoji without color dropped from a link label (and UnsupportedOperationException with a color).
- Not ported because upstream does not use them: Splitter's tokenizer (htmlTag, purgeAllTag), HtmlCommand, PlainText, TextLink, Img, UCreole; the `@startcreole` diagram (PSystemCreole, PSystemCreoleFactory) is a diagram type, not part of this step.
- `SheetBlock1.Neutrons`, `SheetBlock2.Neutrons` and `StripeLatex.Neutrons` panic as upstream (UnsupportedOperationException).
- API notes: `NewSheetBlock1` (with margins), `NewSheetBlock1Padding`, `NewSheetBlock1Same` and `NewSheetBlock1Style` are upstream's four constructors; `SheetBlock2.EnlargeMe` returns `*SheetBlock2` (upstream's type; it was `TextBlock` in the stub); `NewCreoleParser` exposes the parser for ISkinSimple implementations outside the package. `Atom` is exported (TextBlock + StartingAltitude + Neutrons); atoms embed `AtomBase` (TextBlockDefaults and TextBlockMemoized). AtomText implements `isAtomText()`.
- Line positions are UTF-16 units as upstream (`StripeSimple` works on `[]uint16`, commands return lengths in units); the regular expression commands match the rest of the line, as Java's `Matcher.region`.

### Sprites

- `Sprite` and `SpriteContainer` are the real interfaces; no implementation is ported (monochrome, color, SVG, image sprites, `sprite` command).
- `SpriteImageFromInternal`: upstream's jar sprites (139 archimate sprites) are not bundled. A name that upstream would find panics (names listed in `sprite_internalnames.go`); any other name gives nil, as upstream.

### Other gaps

- `TextBlockInEllipse` needs `svek.image.Footprint` and `ContainingEllipse`: construction panics.
- `UbrexUrlBuilder` and `StereotypePattern.umandatory/uoptional` are not ported: `NewUbrexURLBuilder` panics, the UBrex leaves are absent. The UBrex engine itself is now ported (`ubrex*.go`, unexported: `ubrexBuild(expression).match(utf16, pos)`); the command leaves built on it (com.plantuml.ubrex.builder, UBrexPart in the command framework) are not.
- `PortableImageARGB.Scale` (image scaling with Java2D interpolation) panics: raster driver step. `PixelImage` is otherwise ported, so `UImage` recoloring works.
- `klimt.compress` (UGraphicCompressOnXorY, slots, expansions) is not ported, only CompressionMode and UShapeIgnorableForCompression.
- `klimt.geom` placement strategies, `ULayoutGroup` and `HorizontalAlignment.draw/asPlacementStrategy` are not ported (they need UGraphic/TextBlock, so they belong here when needed).
- Methods returning diagram-level types are not on the klimt types; the packages that define those types should provide functions:
  - `Stereogroup.getBoxStyle` (activitydiagram3.ftile.BoxStyle) and `getLeafType` (abel.LeafType): compute from `Stereogroup.Labels()`.
  - `PackageStyle.toUSymbol`, `ComponentStyle.toUSymbol`, `ActorStyle.toUSymbol` (decoration.symbol).
  - `Direction.getWBSDirection` (WBS diagram).
  - `Display.getWithNewlines(Quark<Entity>)` (abel), `Display.asciiDimension/asciiDraw` (asciiverse).
- `CommandStyleImport`: local files are read through the render context file system (`ISkinParam.PathSystem()`), then the embedded skins; remote http/https style sheets return the error "Cannot read: <path> (remote style sheets are not supported yet)".
- Left for later steps: `PngTitler` (PNG step); `klimt.font.UnusedSpace`, which centers a glyph by rasterizing it, needed by the SVG driver for UCenteredCharacter (it needs glyph outlines).
- `PortableImageARGB.Scale` panics (above), so `<img:...{scale=x}>` with a raster image panics when the image is measured or drawn (the golden case `imgpngscale` is skipped for this).

- Not ported on purpose: logging (`Log.info` of used styles), `Check.isJunit` (process-wide test flag), `UAntiAliasing.apply`, the java.awt conversion of `AffineTransformType`, `UText.createTextLayout`, `DotPath.draw(Graphics2D)`, the TeaVM branches.

### Skin components (for the Rose components step)

- A component embeds `AbstractComponent` (or `AbstractTextualComponent`) and calls `InitAbstractComponent(self, style, skinParam)` (or `InitAbstractTextualComponent(self, style, stereo, maxMessageSize, padding, skinParam, display, enhanced)` / `InitAbstractTextualComponentLabel(self, style, maxMessageSize, padding, skinParam, label, ok)`) with itself, so that `DrawU(ug, area, context)` and `PreferredDimension` reach its methods. `AbstractComponentHooks` lists them: `DrawInternalU`, `PreferredWidth` and `PreferredHeight` are required; `DrawBackgroundInternalU`, `PaddingX` and `PaddingY` default to the base (no background, 0) unless the component defines its own. Upstream's protected getters are exported methods without "get" (`Style`, `SkinParam`, `SymbolContext`, `ColorLine`, `TextBlock`, `TextWidth`, `TextHeight`, `OldPaddingX1`...). `UsedStyles` and `StyleSignature` panic as upstream.
- `ArrowConfiguration` is an immutable pointer type; `ApplyStroke(ug, style)` is upstream's current applyStroke, `ApplyStrokeLegacy(ug)` the deprecated one. `ArrowDressing` is a value.
- `ComponentType.StyleSignature` inlines `sequencediagram.LifeEventType.DESTROY.getStyleSignature()`.
- What `skin/rose` imports from outside skin, klimt, style and creole is not in this package: `decoration.symbol.USymbols`, `svek.image.Opale`, `svek.EntityDomain`, `svek.Control`, `svek.Boundary` (text blocks of the participant heads and notes), and `sequencediagram.Participant` and `NotePosition`.

## Design notes for the next steps

- `UShape` and `UChange` are `any`: `color.HColor` (a foreground color change) lives in the color package and cannot implement a klimt marker. `Bg(color)` is upstream's `HColor.bg()`, `ChangeBack(ug)` is `HColors.changeBack`.
- Output drivers (svg, g2d) embed `AbstractUGraphic[O]`, implement `CommonUGraphic` (`CopyUGraphic` must call `SetSelf` on the copy and `CopyFrom`), register drivers with `RegisterDriver[S]`/`IgnoreShape[S]`, and may implement the hooks `ManageHiddenAutomatically() bool`, `BeforeDraw()`, `AfterDraw()`, `DrawComment(*UComment)`. As upstream, drivers are per graphic: register them in the constructor and again in `CopyUGraphic`, so that drivers built on the graphic (upstream passes it as the ClipContainer) read the copy's clip.
- Graphics derived by `Apply` share their bounds and output object, as upstream.
- Java nullability: `*UStroke` where a stroke may be null, `*font.UFont` for null fonts, `(value, ok)` for null strings and enums, `HorizontalAlignmentNull` for null alignments, nil for colors and text blocks.
- `ISkinSimple.PathSystem()` returns the `*core.RenderContext`, which stands for upstream's `nio.PathSystem`.
- `SkinParam.Value` caches results, missing ones included, and `SetParam` does not clear the cache: a value read before it is set stays missing, as upstream.
- No mutable package state (design D3): the parsed embedded skins are parsed once (`sync.Once` per skin) and cloned for each `StyleLoaderLoadSkin`; style values and styles are immutable once parsed (upstream's `ValueImpl` number caches are dropped, the results are the same). `StyleBuilder` guards its merged-style cache with a mutex. Read-only tables (SName names, skinparam conversions, sprite names) are built at initialization.
- Java orders: EnumMap/EnumSet as enum-ordered arrays and bitsets; LinkedHashMap as `orderedMap`; HashMap/HashSet orders that reach output are emulated by `javaHashMapOrder`/`javaHashOrder` (Ports, `SkinParam.TransformStringForSizeHack`, `DotPath.Sample`), which replay HashMap's resizes, including the one on a crowded bucket.
- `StyleSignatureBasic` is a comparable value; its stereotypes are kept sorted, so its String differs from upstream's HashSet order (only the String).
- `CMapString` takes a slice: upstream's Set order decides between links of equal area, here the caller's order does.
- For the SVG driver, creole draws: `UText` with the FontConfiguration of the run (font, color, styles UNDERLINE/STRIKE/WAVE/BACKCOLOR with the extended color, the `Space()` of superscripts and subscripts already applied to the position, and the SVG attributes of `<text ...>` in `FontConfiguration.Attributes()`); `StartURL`/`CloseURL` around link text and images; `UHorizontalLine` for rules, which reach the driver only through `UGraphicStencil` (SheetBlock2 wraps its graphic in one); `UImage` (raster data URLs, possibly scaled) and `UImageSvg`; rectangles, ellipses and lines for bullets, tables and trees. On a graphic matching "SPECIALTXT", AtomText draws itself (`ug.Draw(atomText)`), and LimitFinder skips it.

## Upstream behavior kept or fixed

- Creole, kept: `<font color=unknown>` panics with `*color.NoSuchColorRuntimeError` (upstream throws NoSuchColorRuntimeException at drawing time), while `<color:unknown>` is ignored; `&#nnn;` and `<U+hhhh>` are replaced again when line breaking splits a text (AtomText.withText reapplies them, so `&#38;#65;` breaks into "A"); an empty legacy tag (`<b></b>x`) falls to the end-of-line command; the vertical padding of SheetBlock1 is added to its width too.
- Creole, fixed per design D16 (the golden cases are listed in `creoleDifferences` and checked by `TestCreoleDifferences`): a list item without text and a maximum width (`* ` with maxMessageSize) gives the bullet alone where upstream throws NullPointerException in Fission; image data that ImageIO cannot read gives "(Cannot decode: <source>)" (upstream's null check fails, NullPointerException later); `&#nnn;` above U+10FFFF is left as text, as an overflow is (upstream throws IllegalArgumentException); a stereotype without labels (`<<O-O>>`) adds no line (upstream throws NullPointerException).
- Kept:
  - Values set only for the dark mode have no regular text: `AsColor` panics (IllegalArgumentException), and strokes, paddings and margins panic (NullPointerException), as upstream.
  - `AsDoubleDefaultTo` never returns its default; the parser keeps the dark scheme after an `@media` block; `participant.bar` is the stereotype "participantbar"; Colors.applyStyle always adds four colors (ValueNull is never null).
  - The styles of a style sheet without stereotype share the values of their parser context until parsing ends, so declarations written after the "}" of the never-popped root or top-level `depth(n)` context still land in them.
  - PixelImage.muteColor overflows the alpha; UPolygon.affine does not update the bounds; LimitFinder rejects UHorizontalLine and UPattern; DotPaths built from another share its curves (MoveStartPoint/MoveEndPoint change them in both); the null Display throws where upstream uses its null list; `SkinParamDelegator.getFont` forces `inPackageTitle` to false.
- Fixed per design D16: `SkinParamDelegator.shadowingForNote` calls itself forever upstream (StackOverflowError); here it delegates.
- Not emulated:
  - HashMap tree bins (a 9th colliding element in a table of 64 slots or more move their tree root first) in the hash order emulation.
  - A FontName set only for the dark mode gives the font of an empty name; upstream builds a font of a null name that fails when measured.
  - Upstream's process-wide StyleLoader cache also caches local skin files by name (up to 30) and finds skins by absolute path and include path; here local skins are read from the render context file system on each load, relative to its current directory.
  - The creole regular expressions are compiled with RE2's `(?i)`, which folds a few non-ASCII letters (the Kelvin sign, the long s) where Java's CASE_INSENSITIVE folds ASCII only (see the regex note below).
  - `javaUpper` only adds the "ß" expansion to Go's simple mappings, `javaParseInt` only takes ASCII digits, and lone surrogates (from `<U+D800>` or a spot character outside the BMP) become U+FFFD: all callers pass ASCII, digits or BMP text.
  - `PortableImageARGB.ToPngDataURL` encodes with Go's image/png: valid, but not byte-identical to Java's ImageIO (color type, filters and compression differ). The raster step decides whether byte parity matters.

## Suggested additions to leaf packages (not done here)

- `color.Colors`: a `WithShadowing(bool)` method, so that upstream's `Colors.applyStereotypeForNote` (which sets the private shadowing field) can be ported; the rest of the style-dependent Colors methods are klimt functions (`ColorsApplyStyle`, `ColorsColorOf`, `ColorsSpecificLineStroke`, `ColorsApplyStroke`, `ColorsMuteStroke`, `ColorsMute`, `ColorsApplyStereotype`).
- `Guillemet`, `ConfigurationStore` and `OptionKey` have no klimt dependency and could move to `internal/text` and the preprocessor; the preprocessor may also import klimt for them.
- `internal/command/regex`: RE2's "." also matches '\r', U+0085, U+2028 and U+2029, and its `(?i)` folds with Unicode rules (the Kelvin sign matches "k"), where Java's "." excludes those line terminators and `CASE_INSENSITIVE` folds ASCII only. The klimt patterns rewrite "." with `javaDots`; the regex package (and the other packages using it) could do the same in `Transform` and restrict the folding.

## Numeric notes

- Trigonometry uses `internal/jmath` (fdlibm, Java's StrictMath), per design D17. HotSpot's `Math.sin/cos` intrinsics differ from StrictMath in the last bit for some inputs (for example `Math.cos(2.9)` on aarch64 JDK 25), so the golden files are generated with `-XX:-InlineMathNatives`, which makes `Math` compute as `StrictMath`.
- Products that feed additions are wrapped in `float64(...)`; untyped constant expressions that Java rounds step by step use variables (for example `pi * 4 / 5`).

## Tests and golden files

- `creole_golden_test.go`, `skin_component_test.go` and `ubrex_test.go` compare against `creole.golden`, `creole_skin.golden` and `ubrex.golden`, written by `testdata/javagen/CreoleGolden.java` in the reference image (instructions in its header): it measures with the harness's fixed StringBounder and with the SVG StringBounder of the reference fonts, which `internal/klimt/font` reproduces. `creole_inputs.txt` has 157 cases (every creole syntax element, nested styles, lists, tables, trees, code blocks, links, escapes, Unicode, line breaking with maximum widths, alignments, creole modes, stereotypes, bordered blocks, error cases); each gives the dimension and the recorded drawing for both bounders. The style-bordered note (`Style.CreateTextBlockBordered`) and `TextBlockTitle` are also recorded. `creole_skin.golden` covers ArrowConfiguration operations and strokes, Area, the four actor styles with three fashions, CircleInterface, the style signatures of ComponentType, and a textual component (labels, displays, enhanced displays with separators, stereotypes, background and content drawing). `ubrex_inputs.txt` checks the UBrex engine directly (the creole expressions and every construct).
- `golden_test.go` and `golden2_test.go` compare against `testdata/golden/*.golden`, written by the upstream jar with `testdata/javagen/KlimtGolden.java` and `KlimtGolden2.java` (instructions in the Java header). They cover style resolution for the five skins, a user style sheet and skinparams, FromSkinparamToStyle, SkinParam getters, Display splitting and operations, stereotypes, links, skinparam key cleaning, text block dimensions and LimitFinder bounds with a fixed StringBounder, recorded drawing operations, Java hash orders, font configurations, strokes and values, DotPath and geometry, and style parser edge cases.
- `klimt_test.go` covers the rest: graphics composition, drivers, stubs, commands, caches and concurrency.
