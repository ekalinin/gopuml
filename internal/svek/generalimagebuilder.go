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
// diagram types not ported yet use (states, activities, descriptions...)
// panic with "not ported yet".
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
		notPorted("EntityImageActivity")
	case leafType == abel.LeafTypePortin || leafType == abel.LeafTypePortout:
		notPorted("EntityImagePort")
	case leafType == abel.LeafTypeState:
		notPorted("EntityImageState")
	case leafType == abel.LeafTypeCircleStart:
		notPorted("EntityImageCircleStart")
	case leafType == abel.LeafTypeCircleEnd:
		notPorted("EntityImageCircleEnd")
	case leafType == abel.LeafTypeBranch || leafType == abel.LeafTypeStateChoice:
		notPorted("EntityImageBranch")
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
		notPorted("EntityImageSynchroBar")
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
		notPorted("EntityImagePseudoState")
	case leafType == abel.LeafTypeStateTransitionLabel:
		notPorted("EntityImageTransitionLabel")
	case leafType == abel.LeafTypeDeepHistory:
		notPorted("EntityImageDeepHistory")
	case leafType == abel.LeafTypeTips:
		return newEntityImageTips(leaf, bibliotekon)
	case leafType == abel.LeafTypeChenEntity, leafType == abel.LeafTypeChenRelationship,
		leafType == abel.LeafTypeChenAttribute, leafType == abel.LeafTypeChenCircle:
		notPorted("EntityImageChen")
	case leafType == abel.LeafTypeDomain, leafType == abel.LeafTypeRequirement:
		notPorted("EntityImageDomain")
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
