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
	// TODO:
	// - followed_by
	// - derived_from
	// - translation_of

)

var relationLabels = map[string][2]string{
	"has_contributor":    {"Has contributor", "Har bidrag fra"},
	"has_subject":        {"Has subject", "Har som emne"},
	"published_by":       {"Published by", "Utgitt av"},
	"in_series":          {"In series", "I serien"},
	"has_parent":         {"Has parent", "Hører til under"}, //  TODO norwegian label sounds odd
	"has_part":           {"Has part", "Innehodler del"},
	"has_classification": {"Has classification", "Klassifisert som"},
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
