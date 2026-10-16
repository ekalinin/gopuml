# internal/klimt handoff

State of the `internal/klimt` package after the drawing model and style system step (tasks 6.1, 6.5 partly, 7.1, 7.2). It lists what is ported, what is a stub, and what the next steps need to know.

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
- `klimt/creole`: Display, CreoleMode, Stencil, Parser (static helpers).
- Moved small types, each with a comment naming its upstream package: svek Ports, PortGeometry, WithPorts, Margins, ShapeType, IEntityImage, PackageStyle, ConditionStyle, ConditionEndStyle, GuideLine; abel Hideable; dot DotSplines; tikz TikzFontDistortion; utils Direction; geom MagneticBorder/MagneticBorderNone; awt XAffineTransform; text Guillemet; jaws constants; warning Warning, WarningHandler, JawsWarning; preproc ConfigurationStore, OptionKey; net.atmp SpecialText and PixelImage; activitydiagram3.ftile CenteredText; compress CompressionMode and UShapeIgnorableForCompression; StringUtils and SignatureUtils helpers.
- Skins: `skin/*.skin` (plantuml, rose, debug, sonyxperiadev, strictuml) are embedded.

## Stubs and missing parts

Every stub panics with `*color.RuntimeError` whose message starts with "not ported yet:" and names the upstream class, or returns an explicit error.

### Creole (next step)

- `creole_stubs.go` holds placeholders: `Sheet`, `SheetBlock1`, `SheetBlock2`, and `newCreoleParser` (upstream's `legacy.CreoleParser`). `SheetBuilder` is the real interface (`CreateSheet(*Display) *Sheet`). The creole step should delete this file and provide:
  - `Sheet`, `NewSheetBlock1(sheet, maxWidth LineBreakStrategy, padding ClockwiseTopRightBottomLeft, marginX1, marginX2) *SheetBlock1` (a TextBlock and a Stencil), `NewSheetBlock2(block, stencil Stencil, defaultStroke UStroke) *SheetBlock2` with `EnlargeMe(left, right) TextBlock` (used by TextBlockBordered), and `newCreoleParser(fc, alignment, skin ISkinSimple, mode, stereo) SheetBuilder`.
  - AtomText must implement the marker method `isAtomText()` (interface `atomTextMarker` in `drawing_abstractugraphic.go`): graphics matching "SPECIALTXT" skip its bounds.
- `Display.Create*`, `Style.CreateTextBlockBordered` and `TextBlockTitle` therefore panic until creole lands. Text blocks that do not need creole work: TextBlockRaw, SingleLine, TileText, GraphicStrings in NO_CREOLE mode (error images).
- `Display` alignment: `HorizontalAlignmentNull` (-1) stands for upstream's null alignment and can reach `Create0` from styles without HorizontalAlignment; the creole code must handle it as upstream handles null.
- Display elements are `CharSequence` (`String() string`): `Text` for plain lines, `*Stereotype`, and sequence message numbers through the `MessageNumber` interface (marker method `MessageNumberMarker()`); embedded diagrams are not handled yet.

### Sprites

- `Sprite` and `SpriteContainer` are the real interfaces; no implementation is ported (monochrome, color, SVG, image sprites, `sprite` command).
- `SpriteImageFromInternal`: upstream's jar sprites (139 archimate sprites) are not bundled. A name that upstream would find panics (names listed in `sprite_internalnames.go`); any other name gives nil, as upstream.

### Other gaps

- `TextBlockInEllipse` needs `svek.image.Footprint` and `ContainingEllipse`: construction panics.
- `UbrexUrlBuilder` and `StereotypePattern.umandatory/uoptional` need the UBrex engine (`com.plantuml.ubrex`), not ported: `NewUbrexURLBuilder` panics, the UBrex leaves are absent.
- `PortableImageARGB.Scale` (image scaling with Java2D interpolation) panics: raster driver step. `PixelImage` is otherwise ported, so `UImage` recoloring works.
- `klimt.compress` (UGraphicCompressOnXorY, slots, expansions) is not ported, only CompressionMode and UShapeIgnorableForCompression.
- `klimt.geom` placement strategies, `ULayoutGroup` and `HorizontalAlignment.draw/asPlacementStrategy` are not ported (they need UGraphic/TextBlock, so they belong here when needed).
- Methods returning diagram-level types are not on the klimt types; the packages that define those types should provide functions:
  - `Stereogroup.getBoxStyle` (activitydiagram3.ftile.BoxStyle) and `getLeafType` (abel.LeafType): compute from `Stereogroup.Labels()`.
  - `PackageStyle.toUSymbol`, `ComponentStyle.toUSymbol`, `ActorStyle.toUSymbol/getTextBlock` (decoration.symbol, skin actors).
  - `Direction.getWBSDirection` (WBS diagram).
  - `Display.getWithNewlines(Quark<Entity>)` (abel), `Display.asciiDimension/asciiDraw` (asciiverse).
- `CommandStyleImport`: local files are read through the render context file system (`ISkinParam.PathSystem()`), then the embedded skins; remote http/https style sheets return the error "Cannot read: <path> (remote style sheets are not supported yet)".
- Left for later steps: `PngTitler` (PNG step); `klimt.font.UnusedSpace`, which centers a glyph by rasterizing it, needed by the SVG driver for UCenteredCharacter (it needs glyph outlines); the sequence and component classes of `skin` (Component, AbstractComponent, AbstractTextualComponent, Area, ArrowConfiguration and the other Arrow* types, ActorStickMan, ActorAwesome, ActorHollow, CircleInterface, ComponentType, Context2D, SimpleContext2D).
- Not ported on purpose: logging (`Log.info` of used styles), `Check.isJunit` (process-wide test flag), `UAntiAliasing.apply`, the java.awt conversion of `AffineTransformType`, `UText.createTextLayout`, `DotPath.draw(Graphics2D)`, the TeaVM branches.

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

## Upstream behavior kept or fixed

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

- `golden_test.go` and `golden2_test.go` compare against `testdata/golden/*.golden`, written by the upstream jar with `testdata/javagen/KlimtGolden.java` and `KlimtGolden2.java` (instructions in the Java header). They cover style resolution for the five skins, a user style sheet and skinparams, FromSkinparamToStyle, SkinParam getters, Display splitting and operations, stereotypes, links, skinparam key cleaning, text block dimensions and LimitFinder bounds with a fixed StringBounder, recorded drawing operations, Java hash orders, font configurations, strokes and values, DotPath and geometry, and style parser edge cases.
- `klimt_test.go` covers the rest: graphics composition, drivers, stubs, commands, caches and concurrency.
