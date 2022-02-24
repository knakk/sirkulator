package vocab

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

// TODO rename to just Relation
type RelationType string

const (
	RelationInvalid        RelationType = "invalid"
	RelationHasContributor RelationType = "has_contributor"
	RelationHasSubject     RelationType = "has_subject"
	RelationPublishedBy    RelationType = "published_by"
	// TODO:
	// - in_series
	// - followed_by
	// - derived_from
	// - translation_of

)

var relationTypeLabels = map[string][2]string{
	"has_contributor": {"Has contributor", "Har bidrag fra"},
	"has_subject":     {"Has subject", "Har som emne"},
	"published_by":    {"Published by", "Utgitt av"},
}

func ParseRelationType(s string) RelationType {
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

func (r RelationType) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && relationTypeLabels[string(r)][1] != "" {
		return relationTypeLabels[string(r)][1]
	}
	return relationTypeLabels[string(r)][0]
}
