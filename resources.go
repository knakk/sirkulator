package sirkulator

import (
	"fmt"
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
)

func AllResourceTypes() []ResourceType {
	return []ResourceType{
		TypePublication,
		TypePublisher,
		TypePerson,
		TypeCorporation,
		TypeLiteraryAward,
		TypeSeries,
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
	return [...]string{"unknown", "publication", "publisher", "person", "corporation", "literary_award", "series"}[r]
}

func (r ResourceType) enLabel() string {
	if r > 7 || r < 0 {
		r = 0 // "unknown"
	}
	return [...]string{"unknown", "publication", "publisher", "person", "corporation", "literary award", "series"}[r]
}

func (r ResourceType) noLabel() string {
	if r > 7 || r < 0 {
		r = 0 // "ukjent"
	}
	return [...]string{"ukjent", "utgivelse", "forlag", "person", "korporasjon", "pris", "serie"}[r]
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
	Data  interface{} // TODO PersistableResource?
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
// The year value 0 is interpreted as unknown/no value, which means that
// we cannot represent the acutaly year 0.
// Either From or To can be 0, denoting unknown start or end of span.
// A zero value of the struct, where both From and To are 0, is not really
// usefull and is to be interpreted as unknown/no value.
// TODO consider pointer to int so that we can distinquish between unkown
// and year 0.
type YearRange struct {
	From   int  `json:"from"`
	To     int  `json:"to"`
	Approx bool `json:"approx"` // TODO or "CA"?
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
	if yr.From != 0 {
		if yr.From < 0 {
			s.WriteString(strconv.Itoa(yr.From * -1))
			if yr.To > 0 {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–") // or -
	if yr.To != 0 {
		if yr.To < 0 {
			s.WriteString(strconv.Itoa(yr.To * -1))
			if yr.From < 0 {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.To))
			if yr.From < 0 {
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

func (yr YearRange) noLabel() string {
	var s strings.Builder
	if yr.Approx {
		if yr.From%100 == 0 && yr.To%100 == 0 {
			if yr.To-yr.From > 100 {
				s.WriteString(strconv.Itoa(yr.From / 100))
				s.WriteString("/")
				s.WriteString(strconv.Itoa(yr.To - 100))
			} else {
				if yr.From < 0 {
					s.WriteString(strconv.Itoa(yr.From * -1))
				} else {
					s.WriteString(strconv.Itoa(yr.From))
				}
			}
			s.WriteString("-tallet")
			if yr.To < 0 {
				s.WriteString(" f.Kr")
			}
			return s.String()
		}
		s.WriteString("ca. ")
	}
	if yr.From != 0 {
		if yr.From < 0 {
			s.WriteString(strconv.Itoa(yr.From * -1))
			if yr.To > 0 {
				s.WriteString(" f.Kr")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–") // or -
	if yr.To != 0 {
		if yr.To < 0 {
			s.WriteString(strconv.Itoa(yr.To * -1))
			if yr.From < 0 {
				s.WriteString(" f.Kr")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.To))
			if yr.From < 0 {
				s.WriteString(" e.Kr")
			}
		}

	}
	return s.String()
}

func (yr YearRange) enLabel() string {
	var s strings.Builder
	if yr.Approx {
		if yr.From%100 == 0 && yr.To%100 == 0 {
			if yr.To-yr.From > 100 {
				s.WriteString(strconv.Itoa((yr.From / 100) + 1))
				s.WriteString("/")
				s.WriteString(strconv.Itoa(yr.To / 100))
			} else {
				if yr.From < 0 {
					s.WriteString(strconv.Itoa((yr.From/100)*-1 + 1))
				} else {
					s.WriteString(strconv.Itoa(yr.To / 100))
				}
			}
			s.WriteString("th century")
			if yr.To < 0 {
				s.WriteString(" BCE")
			}
			return s.String()
		}
		s.WriteString("ca. ")
	}
	if yr.From != 0 {
		if yr.From < 0 {
			s.WriteString(strconv.Itoa(yr.From * -1))
			if yr.To > 0 {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.From))
		}
	} else {
		s.WriteString("?")
	}
	s.WriteString("–") // or -
	if yr.To != 0 {
		if yr.To < 0 {
			s.WriteString(strconv.Itoa(yr.To * -1))
			if yr.From < 0 {
				s.WriteString(" BCE")
			}
		} else {
			s.WriteString(strconv.Itoa(yr.To))
			if yr.From < 0 {
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
	Year int `json:"year,omitempty"`

	// "Work" / orignal title info
	// WorkRepresentative bool  / WorkExample - prefer first edition in original language
	// WorkClassic bool // homer, bible, pre 1500 books etc
	YearFirst        int    `json:"year_first,omitempty"`
	TitleOriginal    string `json:"title_original,omitempty"`
	LanguageOriginal string `json:"language_original,omitempty"`

	// Content info
	Language       string   `json:"language,omitempty"`
	LanguagesOther []string `json:"languages_other"`
	GenreForms     []string `json:"genre_forms"`
	Fiction        bool     `json:"fiction"`
	Nonfiction     bool     `json:"nonfiction"`
	Subjects       []string

	// Physical info
	Format   string `json:"format"` // hardcover, paperback, innbundet, heftet etc... Binding?
	NumPages int    `json:"numpages,omitempty"`
	// Weight   int    `json:"weight,omitempty"` // in grams
	// Height   int    `json:"height,omitempty"` // in mm
	// Width    int    `json:"width,omitempty"`  // in mm
	// Depth    int    `json:"depth,omitempty"`  // in mm
	// Ex physical numbers: https://www.akademika.no/liv-koltzow/koltzow-liv/9788203365133
}

type Publisher struct {
	YearRange
}

type Person struct {
	YearRange      YearRange    `json:"year_range,omitempty"` // TODO pointer *YearRange?
	Name           string       `json:"name"`
	Gender         vocab.Gender `json:"gender"`
	NameVariations []string     `json:"name_variations"`
	// PlaceAssociations denotes places where the person lived and worked.
	// It can be a country, region, historical empire, city etc, but it
	// will most commonly be one or more countries (~ "nationality")
	PlaceAssociations []string
}

func (p Person) Label() string {
	if p.YearRange.From != 0 || p.YearRange.To != 0 {
		return fmt.Sprintf("%s (%s)", p.Name, p.YearRange)
	}
	return p.Name
}

// Corporation TODO rename Organization?
type Corporation struct {
	YearRange      YearRange `json:"year_range,omitempty"`
	Name           string    `json:"name"`
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

// 3) Circulation: Item, User, Staff etc

// 4) Various

type Relation struct {
	FromID string
	ToID   string
	Type   string
	Data   map[string]interface{} // TODO consider map[string]string or [][2]string
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
