package sirkulator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/vocab"
	"github.com/teris-io/shortid"
	"golang.org/x/text/language"
)

type PersistableResource interface {
	Valid() bool
	Label() string
}

type ResourceType int

const (
	// TODO consider string consts instead
	TypeUnknown ResourceType = iota
	TypePublication
	TypePublisher
	TypePerson
	TypeCorporation
	TypeLiteraryAward
	TypeSeries
	TypeDewey
)

func AllResourceTypes() []ResourceType {
	return []ResourceType{
		TypePublication,
		TypePublisher,
		TypePerson,
		TypeCorporation,
		TypeLiteraryAward,
		TypeSeries,
		TypeDewey, // TODO TypeClassification?
	}
}

func ParseResourceType(s string) ResourceType {
	for _, t := range AllResourceTypes() {
		if s == t.String() {
			return t
		}
	}
	return TypeUnknown
}

func (r ResourceType) String() string {
	if r > 7 || r < 0 {
		r = 0 // "unknown"
	}
	return [...]string{"unknown", "publication", "publisher", "person", "corporation", "literary_award", "series", "dewey"}[r]
}

func (r ResourceType) enLabel() string {
	if r > 7 || r < 0 {
		r = 0 // "unknown"
	}
	return [...]string{"Unknown", "Publication", "Publisher", "Person", "Corporation", "Literary award", "Series", "Dewey number"}[r]
}

func (r ResourceType) noLabel() string {
	if r > 7 || r < 0 {
		r = 0 // "ukjent"
	}
	return [...]string{"Ukjent", "Utgivelse", "Forlag", "Person", "Korporasjon", "Litterær pris", "Serie", "Deweynummer"}[r]
}

// Label returns a localized string representation of ResourceType.
func (r ResourceType) Label(tag language.Tag) string {
	lang, _, _ := localizer.Matcher.Match(tag)
	switch lang {
	case language.English:
		return r.enLabel()
	case language.Norwegian:
		return r.noLabel()
	default:
		panic("ResourceType.Label: unsupported language " + lang.String())
	}
}

// 1) Abstract/generic/shared types:

type Resource struct {
	Type  ResourceType
	ID    string
	Label string // Synthesized from Data properties
	Links [][2]string
	Data  any // TODO Persistable?
	// TODO candidates/thinking:
	// Description string   // Short description (max 50 characters)
	// Aliases     []string // To function as "synonyms", giving hits when searching, but not for display?
	// Tags        []string // anything that's interesting for discovery/categorization: subject, suptype, genre, etc

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt time.Time
}

// SimpleResource is a minimal representation of a Resource that can be
// displayed and referenced to (i.e generate a URL/link from type+ID).
// If ID is empty it is not persisted or intended for persistence.
type SimpleResource struct {
	Type  ResourceType
	ID    string
	Label string
}

// YearRange represents a span of years, with a from and to year,
// and a specification if it is accurate (Approx=false) or approximate (Approx=true).
// Using json.Number instead of int allow us to distinguish between year 0 and no value.
// It also simplifies form validation.
type YearRange struct {
	From   json.Number `json:"from,omitempty"`
	To     json.Number `json:"to,omitempty"`
	Approx bool        `json:"approx"`
	// TODO consider ApproxFrom and ApproxTo instead of Approx
	// TODO consider Active (=Virksom) bool
}

// String returns a (ideally language-agnostic) string representation of YearRange.
// TODO find language independent representation of BC/AD
func (yr YearRange) String() string {
	var s strings.Builder
	if yr.Approx {
		s.WriteString("ca. ")
	}
	if yr.From != "" {
		if strings.HasPrefix(string(yr.From), "-") {
			s.WriteString(string(yr.From)[1:])
			if !strings.HasPrefix(string(yr.To), "-") {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(string(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–")
	if yr.To != "" {
		if strings.HasPrefix(string(yr.To), "-") {
			s.WriteString(string(yr.To)[1:])
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(string(yr.To))
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" AD")
			}
		}

	}
	return s.String()
}

// Label returns a localized string representation of YearRange.
func (yr YearRange) Label(tag language.Tag) string {
	lang, _, _ := localizer.Matcher.Match(tag)
	switch lang {
	case language.English:
		return yr.enLabel()
	case language.Norwegian:
		return yr.noLabel()
	default:
		panic("YearRange.Label: unsupported language " + lang.String())
	}
}

var rxpYear = regexp.MustCompile(`^-?\d{1,4}$`)

func (yr YearRange) Valid() bool {
	return (yr.From == "" || rxpYear.MatchString(string(yr.From))) &&
		(yr.To == "" || rxpYear.MatchString(string(yr.To)))
}

func (yr YearRange) noLabel() string {
	var s strings.Builder
	if yr.Approx {
		if strings.HasSuffix(string(yr.From), "00") && strings.HasSuffix(string(yr.To), "00") {
			from, _ := strconv.Atoi(string(yr.From))
			to, _ := strconv.Atoi(string(yr.To))
			if to-from > 100 {
				s.WriteString(strconv.Itoa(from / 100))
				s.WriteString("/")
				s.WriteString(strconv.Itoa(to - 100))
			} else {
				if strings.HasPrefix(string(yr.From), "-") {
					s.WriteString(string(yr.From)[1:])
				} else {
					s.WriteString(string(yr.From))
				}
			}
			s.WriteString("-tallet")
			if strings.HasPrefix(string(yr.To), "-") {
				s.WriteString(" f.Kr")
			}
			return s.String()
		}
		s.WriteString("ca. ")
	}
	if yr.From != "" {
		if strings.HasPrefix(string(yr.From), "-") {
			s.WriteString(string(yr.From)[1:])
			if !strings.HasPrefix(string(yr.To), "-") {
				s.WriteString(" f.Kr")
			}
		} else {
			s.WriteString(string(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–")
	if yr.To != "" {
		if strings.HasPrefix(string(yr.To), "-") {
			s.WriteString(string(yr.To)[1:])
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" f.Kr")
			}
		} else {
			s.WriteString(string(yr.To))
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" e.Kr")
			}
		}

	}
	return s.String()
}

func (yr YearRange) enLabel() string {
	var s strings.Builder
	if yr.Approx {
		if strings.HasSuffix(string(yr.From), "00") && strings.HasSuffix(string(yr.To), "00") {
			from, _ := strconv.Atoi(string(yr.From))
			to, _ := strconv.Atoi(string(yr.To))
			if to-from > 100 {
				s.WriteString(strconv.Itoa((from / 100) + 1))
				s.WriteString("/")
				s.WriteString(strconv.Itoa(to / 100))
			} else {
				if strings.HasPrefix(string(yr.From), "-") {
					s.WriteString(strconv.Itoa((from/100)*-1 + 1))
				} else {
					s.WriteString(strconv.Itoa(to / 100))
				}
			}
			s.WriteString("th century")
			if strings.HasPrefix(string(yr.To), "-") {
				s.WriteString(" BCE")
			}
			return s.String()
		}
		s.WriteString("ca. ")
	}
	if yr.From != "" {
		if strings.HasPrefix(string(yr.From), "-") {
			s.WriteString(string(yr.From)[1:])
			if !strings.HasPrefix(string(yr.To), "-") {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(string(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–")
	if yr.To != "" {
		if strings.HasPrefix(string(yr.To), "-") {
			s.WriteString(string(yr.To)[1:])
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(string(yr.To))
			if strings.HasPrefix(string(yr.From), "-") {
				s.WriteString(" AD")
			}
		}
	}
	return s.String()
}

// AgentContribution is a contribution from the viewpoint of an Agent (Person|Corporation).
type AgentContribution struct {
	SimpleResource
	Year  int // TODO publication year or first published year?
	Roles []marc.Relator
}

// PublicationContribution is a contribution from the view point of a Publication.
type PublicationContribution struct {
	Agent SimpleResource
	Roles []marc.Relator
}

// 2) Concrete types

type Publication struct {
	// Title and publishing info
	Title     string   `json:"title"`
	Subtitle  string   `json:"subtitle,omitempty"`
	Publisher string   `json:"publisher,omitempty"`
	Series    []string `json:"series"`
	// Note, Year|YearFirst=0 means we cannot have a publication published in year 0,
	// assumed this to be not a practical problem, not a lot of known classics published that year (TODO any?)
	// https://en.wikipedia.org/wiki/Ancient_literature
	Year json.Number `json:"year,omitempty"`

	// "Work" / orignal title info
	// WorkRepresentative bool  / WorkExample - prefer first edition in original language
	// WorkClassic bool // homer, bible, pre 1500 books etc
	YearFirst        json.Number `json:"year_first,omitempty"`
	TitleOriginal    string      `json:"title_original,omitempty"`
	LanguageOriginal string      `json:"language_original,omitempty"`

	// Content info
	Language       string   `json:"language,omitempty"`
	LanguagesOther []string `json:"languages_other"`
	GenreForms     []string `json:"genre_forms"`
	Audiences      []string `json:"audiences"`
	Fiction        bool     `json:"fiction"`
	Nonfiction     bool     `json:"nonfiction"`
	Subjects       []string

	// Physical info
	Format   string      `json:"format"` // hardcover, paperback, innbundet, heftet etc... Binding?
	NumPages json.Number `json:"numpages,omitempty"`
	// Weight   int    `json:"weight,omitempty"` // in grams
	// Height   int    `json:"height,omitempty"` // in mm
	// Width    int    `json:"width,omitempty"`  // in mm
	// Depth    int    `json:"depth,omitempty"`  // in mm
	// Ex physical numbers: https://www.akademika.no/liv-koltzow/koltzow-liv/9788203365133
}

// TODO how to get author into the picture?
func (p Publication) Label() string {
	return "TODO"
}

type Publisher struct {
	YearRange
}

type Person struct {
	YearRange      YearRange    `json:"year_range"` // TODO pointer *YearRange?
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	NameVariations []string     `json:"name_variations"`
	Gender         vocab.Gender `json:"gender"`
	Countries      []string     `json:"countries"`
	Nationalities  []string     `json:"nationalities"`
}

func (p Person) Label() string {
	if p.YearRange.From != "" || p.YearRange.To != "" {
		return fmt.Sprintf("%s (%s)", p.Name, p.YearRange)
	}
	return p.Name
}

// Corporation TODO rename Organization?
type Corporation struct {
	YearRange      YearRange `json:"year_range"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ParentName     string    `json:"parent_name,omitempty"`
	NameVariations []string  `json:"name_variations"`
	//Type           vocab.CorporationType `json:"type"` // University, Municipality, Music gorup, Record label etc
}

func (c Corporation) Label() string {
	if c.ParentName != "" {
		return fmt.Sprintf("%s / %s", c.Name, c.ParentName)
	}
	return c.Name
}

// Character is a fictional or mythical person/character.
// Examples: Ulysses, Donald Duck, Harry Hole
type Character struct {
	Name           string   `json:"name"`
	NameVariations []string `json:"name_variations"`
}

type Dewey struct {
	Number string   `json:"number"` // same as resource.ID
	Name   string   `json:"name"`   // Only norwegian label for now
	Terms  []string `json:"terms"`  // Henvisningstermer
}

func (d Dewey) Label() string {
	return fmt.Sprintf("%s %s", d.Number, d.Name)
}

// 3) Circulation: Item, User, Staff etc

// 4) Various

type Relation struct {
	ID     int64
	FromID string
	ToID   string
	Type   string
	Data   map[string]any // TODO consider map[string]string or [][2]string
}

type RelationExp struct {
	Relation
	From SimpleResource
	To   SimpleResource
}

type Image struct {
	ID     string
	Type   string // MIME type, but stored without "image/" prefix
	Height int
	Width  int
}

// GetNewID returns a new string ID which can be used to persist a
// new Resource to DB.
// NOTE: Currently this is a variable so that it can be overwritten with a
//       deterministic function for test purposes. TODO revise.
var GetNewID func() string = shortid.MustGenerate
