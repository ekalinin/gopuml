// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/GeneralImageBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
)

// createEntityImageBlock is GeneralImageBuilder.createEntityImageBlock: the
// image of a leaf, by its type. The images of the entities that only the
// diagram types not ported yet use (Chen diagrams...) panic with "not
// ported yet".
func createEntityImageBlock(leaf *abel.Entity, isHideEmptyDescriptionForState bool, portionShower PortionShower,
	bibliotekon *Bibliotekon, graphvizVersion dot.GraphvizVersion, links []*abel.Link) klimt.IEntityImage {
	if leaf.IsRemoved() {
		throwIllegalState("")
	}
	leafType := leaf.LeafType()
	switch {
	case leafType.IsLikeClass():
		entityImageClass := newEntityImageClass(leaf, portionShower)
		if bibliotekon != nil {
			if n := bibliotekon.neighborhoods[leaf]; n != nil {
				return &EntityImageProtected{orig: entityImageClass, border: 20, neighborhood: n,
					bibliotekon: bibliotekon}
			}
		}
		return entityImageClass
	case leafType == abel.LeafTypeNote:
		return newEntityImageNote(leaf)
	case leafType == abel.LeafTypeActivity:
		return newEntityImageActivity(leaf, bibliotekon)
	case leafType == abel.LeafTypePortin || leafType == abel.LeafTypePortout:
		return newEntityImagePort(leaf, bibliotekon.cluster(leaf.ParentContainer()), bibliotekon)
	case leafType == abel.LeafTypeState:
		if leaf.EntityPosition() != abel.EntityPositionNormal {
			return newEntityImageStateBorder(leaf, bibliotekon.cluster(leaf.ParentContainer()), bibliotekon)
		}
		if isHideEmptyDescriptionForState && len(leaf.Bodier().RawBody()) == 0 {
			return newEntityImageStateEmptyDescription(leaf)
		}
		if st := leaf.Stereotype(); st != nil {
			if label, ok := st.Label(klimt.GuillemetDoubleComparator()); ok && label == "<<sdlreceive>>" {
				return newEntityImageState2(leaf)
			}
		}
		return newEntityImageState(leaf)
	case leafType == abel.LeafTypeCircleStart:
		return newEntityImageCircleStart(leaf)
	case leafType == abel.LeafTypeCircleEnd:
		return newEntityImageCircleEnd(leaf)
	case leafType == abel.LeafTypeBranch || leafType == abel.LeafTypeStateChoice:
		return newEntityImageBranch(leaf)
	case leafType == abel.LeafTypeLollipopFull || leafType == abel.LeafTypeLollipopHalf:
		return newEntityImageLollipopInterface(leaf)
	case leafType == abel.LeafTypeCircle, leafType == abel.LeafTypeDescription, leafType == abel.LeafTypeUsecase,
		leafType == abel.LeafTypeUsecaseBusiness:
		return newEntityImageDescription(leaf, portionShower, links, bibliotekon)
	case leafType == abel.LeafTypeObject:
		return newEntityImageObject(leaf, portionShower)
	case leafType == abel.LeafTypeMap:
		return newEntityImageMap(leaf, portionShower)
	case leafType == abel.LeafTypeJSON:
		return newEntityImageJson(leaf, portionShower)
	case leafType == abel.LeafTypeSynchroBar || leafType == abel.LeafTypeStateForkJoin:
		return newEntityImageSynchroBar(leaf)
	case leafType == abel.LeafTypeArcCircle:
		notPorted("EntityImageArcCircle")
	case leafType == abel.LeafTypePointForAssociation:
		return newEntityImageAssociationPoint(leaf)
	case leafType == abel.LeafTypeNull && leaf.IsGroup():
		notPorted("EntityImageGroup")
	case leafType == abel.LeafTypeEmptyPackage:
		if leaf.USymbol() != abel.USymbolNull {
			return newEntityImageDescription(leaf, portionShower, links, bibliotekon)
		}
		return newEntityImageEmptyPackage(leaf, portionShower)
	case leafType == abel.LeafTypeAssociation:
		return newEntityImageAssociation(leaf)
	case leafType == abel.LeafTypePseudoState:
		return newEntityImagePseudoState(leaf, "H")
	case leafType == abel.LeafTypeStateTransitionLabel:
		return newEntityImageTransitionLabel(leaf)
	case leafType == abel.LeafTypeDeepHistory:
		return newEntityImagePseudoState(leaf, "H*")
	case leafType == abel.LeafTypeTips:
		return newEntityImageTips(leaf, bibliotekon)
	case leafType == abel.LeafTypeChenEntity, leafType == abel.LeafTypeChenRelationship,
		leafType == abel.LeafTypeChenAttribute, leafType == abel.LeafTypeChenCircle:
		notPorted("EntityImageChen")
	case leafType == abel.LeafTypeDomain || leafType == abel.LeafTypeRequirement:
		return newEntityImageProblemFrames(leaf)
	default:
		throwUnsupported(leafType.String())
	}
	return nil
}

// forcedStroke is GeneralImageBuilder.getForcedStroke.
func forcedStroke(stereotype *klimt.Stereotype, skinParam klimt.ISkinParam) klimt.UStroke {
	if stroke := skinParam.Thickness(klimt.LineParamPackageBorder, stereotype); stroke != nil {
		return *stroke
	}
	return klimt.UStrokeWithThickness(1.5)
}

// newEntityImageProblemFrames is the end of upstream's createEntityImageBlockInternal:
// the image of a domain chosen by its stereotype, or of a requirement.
func newEntityImageProblemFrames(leaf *abel.Entity) klimt.IEntityImage {
	stereotype := leaf.Stereotype()
	isDomain := leaf.LeafType() == abel.LeafTypeDomain
	switch {
	case isDomain && stereotype != nil && stereotype.IsMachineOrSpecification():
		return newEntityImageMachine(leaf)
	case isDomain && stereotype != nil && stereotype.IsDesignedOrSolved():
		return newEntityImageDesignedDomain(leaf)
	case leaf.LeafType() == abel.LeafTypeRequirement:
		return newEntityImageRequirement(leaf)
	case isDomain && stereotype != nil && stereotype.IsLexicalOrGiven():
		return newEntityImageDomain(leaf, 'X')
	case isDomain && stereotype != nil && stereotype.IsCausal():
		return newEntityImageDomain(leaf, 'C')
	case isDomain && stereotype != nil && stereotype.IsBiddableOrUncertain():
		return newEntityImageDomain(leaf, 'B')
	}
	return newEntityImageDomain(leaf, 'P')
}
