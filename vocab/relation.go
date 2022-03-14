package vocab

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

type Relation string

const (
	RelationInvalid        Relation = "invalid"
	RelationHasContributor Relation = "has_contributor"
	RelationHasSubject     Relation = "has_subject"
	RelationPublishedBy    Relation = "published_by"
	RelationInSeries       Relation = "in_series"
	RelationHasParent      Relation = "has_parent"
	RelationHasPart        Relation = "has_part"
	// TODO:
	// - followed_by
	// - derived_from
	// - translation_of

)

var relationLabels = map[string][2]string{
	"has_contributor": {"Has contributor", "Har bidrag fra"},
	"has_subject":     {"Has subject", "Har som emne"},
	"published_by":    {"Published by", "Utgitt av"},
	"in_series":       {"In series", "I serien"},
	"has_parent":      {"Has parent", "Hører til under"}, //  TODO norwegian label sounds odd
	"has_part":        {"Has part", "Innehodler del"},
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
	default:
		return RelationInvalid
	}
}

func (r Relation) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && relationLabels[string(r)][1] != "" {
		return relationLabels[string(r)][1]
	}
	return relationLabels[string(r)][0]
}
