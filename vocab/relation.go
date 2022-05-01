package vocab

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

type Relation string

func (r Relation) String() string {
	return string(r)
}

const (
	RelationInvalid           Relation = "invalid"
	RelationHasContributor    Relation = "has_contributor"
	RelationHasSubject        Relation = "has_subject"
	RelationPublishedBy       Relation = "published_by"
	RelationInSeries          Relation = "in_series"
	RelationHasParent         Relation = "has_parent"
	RelationHasPart           Relation = "has_part"
	RelationHasClassification Relation = "has_classification" // has_dewey?
	RelationSubsidiaryOf      Relation = "subsidiary_of"      // TODO has_parent is enough?
	RelationImprintOf         Relation = "imprint_of"
	// TODO:
	// - followed_by
	// - derived_from
	// - translation_of

)

var relationLabels = map[string][4]string{
	"invalid":            {"Invalid", "Ugyldig", "Invalid", "Ugyldig"},
	"has_contributor":    {"Has contributor", "Har bidrag fra", "Is contributing to", "Har bidrag i"},
	"has_subject":        {"Has subject", "Har som emne", "Is subject of", "Er emne i"},
	"published_by":       {"Published by", "Utgitt av", "Published", "Utgta"},
	"in_series":          {"In series", "I serien", "Has serial entry", "Har seriedel"},      // ?
	"has_parent":         {"Has parent", "Hører til under", "Is parent of", "Er overordnet"}, //  TODO norwegian label sounds odd
	"has_part":           {"Has part", "Inneholder del", "Is part of", "Er del av"},
	"has_classification": {"Has classification", "Klassifisert som", "Is classification of", "Er klassifikasjon for"},
	"subsidiary of":      {"Subsidiary of", "Datterselskap av", "Has subsidiary", "Har datterselskap"},
	"imprint_of":         {"Imprint of", "Imprint under", "Has imprint", "Har imprint"},
}

func ParseRelation(s string) Relation {
	switch s {
	case "has_contributor":
		return RelationHasContributor
	case "has_subject":
		return RelationHasSubject
	case "published_by":
		return RelationPublishedBy
	case "in_series":
		return RelationInSeries
	case "has_parent":
		return RelationHasParent
	case "has_part":
		return RelationHasPart
	case "has_classification":
		return RelationHasClassification
	case "subsidiary_of":
		return RelationSubsidiaryOf
	case "imprint_of":
		return RelationImprintOf
	default:
		return RelationInvalid
	}
}

func (r Relation) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	i := 0
	if match == language.Norwegian {
		i = 1
	}
	return relationLabels[string(r)][i]
}

func (r Relation) InverseLabel(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	i := 2
	if match == language.Norwegian {
		i = 3
	}
	return relationLabels[string(r)][i]
}
