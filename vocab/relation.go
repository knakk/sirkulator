package vocab

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

// TODO rename to just Relation
type Relation string

const (
	RelationInvalid        Relation = "invalid"
	RelationHasContributor Relation = "has_contributor"
	RelationHasSubject     Relation = "has_subject"
	RelationPublishedBy    Relation = "published_by"
	// TODO:
	// - in_series
	// - followed_by
	// - derived_from
	// - translation_of

)

var relationLabels = map[string][2]string{
	"has_contributor": {"Has contributor", "Har bidrag fra"},
	"has_subject":     {"Has subject", "Har som emne"},
	"published_by":    {"Published by", "Utgitt av"},
}

func ParseRelation(s string) Relation {
	switch s {
	case "has_contributor":
		return RelationHasContributor
	case "has_subject":
		return RelationHasSubject
	case "published_by":
		return RelationPublishedBy
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
