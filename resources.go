package sirkulator

import (
	"time"

	"github.com/teris-io/shortid"
)

/*type PersistableResource interface {
	Validate() error
}*/

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

func (r ResourceType) String() string {
	if r > 7 || r < 0 {
		r = 0 // "unknown"
	}
	return [...]string{"unknown", "publication", "publisher", "person", "corporation", "literary_award", "series"}[r]
}

// 1) Abstract/generic/shared types:

type Resource struct {
	Type  ResourceType
	ID    string
	Label string // Synthesized from Data properties
	Links [][2]string
	//Aliases []string    // To function as "synonyms", giving hits when searching, but not for display?
	Data interface{}

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt time.Time
}

type SimpleResource struct {
	Type  ResourceType
	ID    string // If ID == "", considered an "unmapped resource"; use Links to map to existing or create a new resource.
	Label string
	Links [][2]string // [2]string{"wikidata", "q213"} [2]string{"viaf", "234234"} etc
}

type YearRange struct {
	FromYear   int  `json:"from_year"`
	ToYear     int  `json:"to_year"`
	Aproximate bool `json:"approx"`
	// TODO consider AproxFrom and AproxTo instead of Aproximate
	// TODO consider "Virksom" bool
}

func (yr YearRange) String() string {
	if yr.Aproximate {

	}
	if yr.FromYear != 0 {

	}
	return "" // yr.FromYear == 0 && yr.ToYear == 0
}

type Contribution struct {
	Role  string
	Agent SimpleResource // Corporation|Person
}

// 2) Concrete types

type Publication struct {
	// Title and publishing info
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle,omitempty"`
	Publisher string `json:"publisher"`
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
	Language       string   `json:"language"`
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
	YearRange `json:"year_range"`
	Name      string `json:"name"`
}

type Corporation struct {
	YearRange
}

// 3) Circulation: Item, User, Staff etc

// 4) Various

type Relation struct {
	FromID string
	ToID   string
	Type   string
	Data   map[string]interface{} // TODO consider map[string]string or [][2]string
}

/*
{
	Type: TypePublication,
	ID:   "Rae943afe",
	Prop: "Publisher",
	Value: "Aschehough"

}

*/

// GetNewID returns a new string ID which can be used to persist a
// new Resource to DB.
// NOTE: Currently this is a variable so that it can be overwritten with a
//       deterministic function for test purposes. TODO revise.
var GetNewID func() string = shortid.MustGenerate
