package etl

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/isbn"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/vocab"
)

// TODO more candidates: https://gist.github.com/boutros/221499f95397a4987a70ce726455319e
var relatorsFrom700e = map[string]string{
	"arranger of music":           "arr",
	"author of introduction, etc": "aui",
	"author":                      "aut",
	"autor":                       "aut", // misspelling?
	"contributor":                 "ctb",
	"cover designer":              "bjd",
	"director":                    "drt",
	"dirigent":                    "cnd",
	"editor":                      "edt",
	"elgitar":                     "", // TODO instrument
	"fiolin":                      "", // TODO instrument
	"gitar":                       "", // TODO instrument
	"illustr":                     "ill",
	"illustrator":                 "ill",
	"klaver":                      "", // TODO instrument
	"kurator":                     "cur",
	"medarb":                      "ctb", // ?
	"narrator":                    "nrt",
	"overs":                       "trl",
	"photographer":                "pht",
	"producer":                    "pro",
	"produsent":                   "pro",
	"red":                         "edt",
	"sang":                        "sng",
	"slagverk":                    "", // TODO instrument
	"sopran":                      "", // TODO intsrument
	"speaker":                     "", // ?
	"tekstforf":                   "lyr",
	"tenor":                       "", // TODO instrument
	"tenorsaksofon":               "", // TODO instrument
	"translator":                  "trl",
	"trompet":                     "",    // TODO instrument
	"utg":                         "pbl", // ?
	"utøver":                      "prf",
	"writer of foreword":          "aui", // TODO foreword != introduction ?
	"writer of introduction":      "aui",
	// "ethnographer": "",
	//"attributed name", "",
	//"coordinador": ""
	//"dir": "drt", // eller dirigent?
	//"traductor": ""
}

// createAgent creates a Resource of type Person/Corporation from the given MARC datafield.
// If the returned Resource has an empty ID, it is to be considered invalid.
func createAgent(f marc.DataField, idFunc func() string) (res sirkulator.Resource) {
	name := f.ValueAt("a")
	if name == "" {
		// No name means invalid resource, return without ID
		return res
	}

	switch f.Tag {
	case "100", "600", "700":
		res.Type = sirkulator.TypePerson
		res.Label = invertName(name)
		person := sirkulator.Person{
			Name: res.Label,
		}
		if lifespan := f.ValueAt("d"); lifespan != "" {
			person.YearRange = parseYearRange(lifespan)
			// TODO consider dropping lifespan from persons born before 0 CE.
			// It is mostly needed for disambiguating betwen different
			// persons with similar or identical names. The further we
			// go back in time, the less likely this is. There is only one Herodotus (or?)
			// In addition, we don't have a language-independent way of
			// denoting BCE/CE.
			res.Label = fmt.Sprintf("%s (%s)", res.Label, person.YearRange)
		}
		res.Data = person
	case "110", "610", "710":
		res.Type = sirkulator.TypeCorporation
		res.Label = name
	default:
		panic("createAgent: unhandled Marc data field " + f.Tag)
	}

	for _, v := range f.ValuesAt("0") {
		// TODO strings.ToLower(v)
		if strings.HasPrefix(v, "(NO-TrBIB)") {
			// Ex: "(NO-TrBIB)90086277"
			// TODO validate ID ^\d+$
			res.Links = append(res.Links,
				[2]string{"bibsys", strings.TrimPrefix(v, "(NO-TrBIB)")})
		}
		if strings.HasPrefix(v, "(orcid)") {
			// Ex: "(orcid)0000-0003-1274-907"
			res.Links = append(res.Links,
				[2]string{"orcid", strings.TrimPrefix(v, "(orcid)")})

		}
		// TODO (DE-588) Deutsche Nationalbibliothek
	}

	res.ID = idFunc()
	return res
}

func matchOrCreate(agents *[]sirkulator.Resource, f marc.DataField, idFunc func() string) sirkulator.Resource {
	// TODO maybe return err, eg. if given marcfield is gibberish?
	name := invertName(f.ValueAt("a"))
	for _, agent := range *agents {
		// First try ID match
		// TODO check agent.Links and match against subfield '0'

		// Second match on name
		if strings.HasPrefix(agent.Label, name) {
			return agent
		}
	}
	agent := createAgent(f, idFunc)
	if agent.ID != "" {
		*agents = append(*agents, agent)
	}
	// TODO take reviews as param, and add review if no id links in agent?
	return agent
}

// TODO it currently never returns an error: when should it fail? For example not enough data
// to construct a valid resource? No label?
func ingestMarcRecord(source string, rec marc.Record, idFunc func() string) (Ingestion, error) {
	// TODO switch on source => one fn per source

	p := sirkulator.Publication{}
	pID := idFunc()
	var label string
	var agents []sirkulator.Resource
	var relations []sirkulator.Relation
	var ing Ingestion
	if len(rec.Leader) > 33 {
		switch rec.Leader[33] {
		case '0', 'e':
			p.Nonfiction = true
		case '1', 'd', 'f', 'j':
			p.Fiction = true
		default:

			// TODO:
			/*
				33 - Literary form (006/16)

					0 - Not fiction (not further specified)
					1 - Fiction (not further specified)
					c - Comic strips
					d - Dramas
					e - Essays
					f - Novels
					h - Humor, satires, etc.
					i - Letters
					j - Short stories
					m - Mixed forms
					p - Poetry
					s - Speeches
					u - Unknown
					| - No attempt to code
			*/
		}
	}
	if f, ok := rec.ControlFieldAt("008"); ok {
		// Fiction/Nonfiction
		if len(f.Value) > 33 {
			switch f.Value[33] {
			case '0', 'e':
				p.Nonfiction = true
			case '1', 'd', 'f', 'j':
				p.Fiction = true

			}
		}
		// Language
		if len(f.Value) > 38 {
			lang := f.Value[35:38]
			// Validate Marc language
			if _, err := marc.ParseLanguage(lang); err == nil {
				p.Language = lang
			}
		}
	}

	if f, ok := rec.DataFieldAt("041"); ok {
		lang := f.ValueAt("h")
		if _, err := marc.ParseLanguage(lang); err == nil {
			p.LanguageOriginal = lang
		}
	}
	if f, ok := rec.DataFieldAt("245"); ok {
		if title := f.ValueAt("a"); title != "" {
			p.Title = cleanTitle(title)
			label = p.Title
		}
		if subtitle := f.ValueAt("b"); subtitle != "" {
			p.Subtitle = subtitle
			label = fmt.Sprintf("%s: %s", label, subtitle)
		}
	}
	if f, ok := rec.DataFieldAt("246"); ok {
		if title := f.ValueAt("a"); title != "" && strings.Contains(f.ValueAt("i"), "ginaltittel") {
			// Matches both "Orignaltittel" and the misspelled "Orginaltittel"
			p.TitleOriginal = title
		}
	}
	// Publisher and published year
	f, ok := rec.DataFieldAt("260")
	if !ok {
		f, ok = rec.DataFieldAt("264")
	}
	// TODO handle multiple 260/264 fields
	if ok {
		if publisher := f.ValueAt("b"); publisher != "" {
			p.Publisher = publisher
			ing.Reviews = append(ing.Reviews, sirkulator.Relation{
				FromID: pID,
				Type:   "published_by",
				Data:   map[string]interface{}{"label": p.Publisher},
			})
		}
		if year := parseYear(f.ValueAt("c")); year != 0 {
			p.Year = year
			label = fmt.Sprintf("%s (%d)", label, year)
		}
	}
	// Publisher series
	for _, f := range rec.DataFieldsAt("490") {
		if series := f.ValueAt("a"); series != "" {
			p.Series = append(p.Series, series)
			data := map[string]interface{}{"label": series}
			if num := f.ValueAt("v"); num != "" {
				n, err := strconv.Atoi(num)
				if err == nil {
					data["number"] = n
				}
			}
			if p.Publisher != "" {
				data["publisher"] = p.Publisher
			}
			ing.Reviews = append(ing.Reviews, sirkulator.Relation{
				FromID: pID,
				Type:   "in_series",
				Data:   data,
			})
		}

	}
	// Physical properties
	if f, ok := rec.DataFieldAt("300"); ok {
		if n := parsePages(f.ValueAt("a")); n != 0 {
			p.NumPages = n
		}
	}
	// Creator/Main entry
	// 100=Person, 110=Corporation
	for _, f := range rec.DataFieldsAt("100", "110") {
		agent := createAgent(f, idFunc)
		agents = append(agents, agent)

		// Add relation from agent to publication
		relator, _ := marc.ParseRelator(f.ValueAt("4"))
		role := relator.Code()
		if role == "" {
			// default
			// TODO different default for mediatypes other than books/monographs
			// TODO different default for other conditions (110?)
			role = "aut"
		}
		relations = append(relations, sirkulator.Relation{
			FromID: pID,
			ToID:   agent.ID,
			Type:   "has_contributor",
			Data:   map[string]interface{}{"role": role, "main_entry": true},
		})

		if role == "aut" {
			switch data := agent.Data.(type) {
			case sirkulator.Person:
				label = fmt.Sprintf("%s - %s", data.Name, label)
			case sirkulator.Corporation:
				// TODO
			}

		}
	}

	// Audience
	for _, f := range rec.DataFieldsAt("385") {
		if audience, err := vocab.ParseAudienceURL(f.ValueAt("0")); err == nil {
			p.Audiences = append(p.Audiences, audience)
		}
	}

	// Subjects
	// https://rdakatalogisering.unit.no/6xx-emneinnforsler/
	// 600 Subject of person
	// 610 Subject of organization
	for _, f := range rec.DataFieldsAt("600", "610") {
		if agent := matchOrCreate(&agents, f, idFunc); agent.ID != "" {
			relations = append(relations, sirkulator.Relation{
				FromID: pID,
				ToID:   agent.ID,
				Type:   "has_subject",
			})
		}
	}
	// 610 Subject of corporation
	// 611 Subject of meeting, conference, event, exhibition etc.
	// 653 Nøkkelord og stikkord (Index term)
	/*
		<datafield ind1=" " ind2=" " tag="653">
			<subfield code="a">skjønnlitteratur</subfield>
			<subfield code="a">roman</subfield>
			<subfield code="a">svensk-litteratur</subfield>
		</datafield>
	*/
	// 655 Genre/literary form
	for _, f := range rec.DataFieldsAt("655") {
		if lang := f.ValueAt("9"); lang == "nno" {
			// skip versions in nynorsk
			continue
		}
		if val := f.ValueAt("a"); val != "" {
			// TODO lowercase first letter?
			p.GenreForms = append(p.GenreForms, val)
		}
	}

	// 7xx contributors
	for _, f := range rec.DataFieldsAt("700", "710") {

		// No role, ignore those entries for now
		// TODO default to "ctb" contributor
		if f.ValueAt("4") == "" && f.ValueAt("e") == "" {
			continue
		}

		if agent := matchOrCreate(&agents, f, idFunc); agent.ID != "" {
			for _, rel := range f.ValuesAt("4") {
				relator, _ := marc.ParseRelator(rel)
				role := relator.Code()
				if role != "" { // We only add relation if we know the role TODO assume default role on certain conditions?
					// TODO check that we don't already have an identical relation, except with main-entry:true
					relations = append(relations, sirkulator.Relation{
						FromID: pID,
						ToID:   agent.ID,
						Type:   "has_contributor",
						Data:   map[string]interface{}{"role": role},
					})
				}
			}
			for _, rel := range f.ValuesAt("e") {
				rel = strings.ToLower(rel)
				rel = strings.TrimSuffix(rel, ".")
				rel = strings.TrimSuffix(rel, ",")
				if match := relatorsFrom700e[rel]; match != "" {
					rel = match
				}
				relator, _ := marc.ParseRelator(rel)
				role := relator.Code()
				if role != "" {
					// TODO check that we don't already have an identical relation, except with main-entry:true
					relations = append(relations, sirkulator.Relation{
						FromID: pID,
						ToID:   agent.ID,
						Type:   "has_contributor",
						Data:   map[string]interface{}{"role": role},
					})
				}
			}
		}
	}

	res := sirkulator.Resource{
		ID:    pID,
		Type:  sirkulator.TypePublication,
		Label: label,
		Data:  p,
	}

	// Publication identifiers: ISBN, ISSN, GTIN (EAN)
	for _, id := range rec.ValuesAt("020", "a") {
		res.Links = append(res.Links, [2]string{"isbn", isbn.Clean(id)})
	}
	for _, issn := range rec.ValuesAt("022", "a") {
		// TODO clean ISSN number
		res.Links = append(res.Links, [2]string{"issn", issn})
	}
	for _, gtin := range rec.ValuesAt("024", "a") {
		// TODO clean GTIN (EAN) number
		res.Links = append(res.Links, [2]string{"gtin", gtin})
	}

	ing.Resources = append(ing.Resources, res)
	ing.Resources = append(ing.Resources, agents...)
	ing.Relations = relations

	var covers []FileFetch
	for _, f := range rec.DataFieldsAt("856") {
		if url := f.ValueAt("u"); url != "" {
			if mime := f.ValueAt("q"); strings.HasPrefix(mime, "image") || strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".jpeg") {
				covers = append(covers, FileFetch{
					ResourceID: pID,
					URL:        url,
				})

				// Bibsys records sometimes omit the original version and just gives us the small one,
				// but we'll also try to get original size image:
				if strings.HasPrefix(url, "https://contents.bibs.aws.unit.no/files/images/small/") {
					covers = append(covers, FileFetch{
						ResourceID: pID,
						URL:        "https://contents.bibs.aws.unit.no/files/images/original/" + url[53:],
					})
				}
			}

		}
	}
	sort.Slice(covers, func(i, j int) bool {
		return strings.Contains(covers[i].URL, "original") && !strings.Contains(covers[j].URL, "original")
	})
	ing.Covers = covers

	// TODO verify that we have enough data for a valid record, ie with Label != ""

	return ing, nil
}

var rxpYear = regexp.MustCompile(`\d{4}`)

func parseYear(s string) int {
	matches := rxpYear.FindAllString(s, -1)
	if len(matches) == 1 {
		n, _ := strconv.Atoi(matches[0])
		return n
	}
	return 0
}

var rxpNumbers = regexp.MustCompile(`[0-9]+`)

func parsePages(s string) int {
	for _, match := range rxpNumbers.FindAllString(s, -1) {
		n, _ := strconv.Atoi(match)
		return n
	}
	return 0
}

func parseYearRange(s string) sirkulator.YearRange {
	//parsingFrom := false
	parsingTo := false
	res := sirkulator.YearRange{}
	start := 0
	pos := 0
	s = strings.ToLower(s)
	var r rune
	var w int
	peekHas := func(sub string) bool {
		if pos-w < 0 || len(s) < pos-w {
			return false
		}
		if strings.HasPrefix(s[pos-w:], sub) {
			pos += len(sub) - w // consume substring
			return true
		}
		return false
	}
	consumeYear := func() int {
		for {
			r, w = utf8.DecodeRuneInString(s[pos:])
			pos += w
			if r < 48 || r > 57 {
				// not a number
				pos -= w
				break
			}
		}
		n, _ := strconv.Atoi(s[start:pos]) // ignoring err since we know we got all digits
		return n
	}
	for pos <= len(s) {
		r, w = utf8.DecodeRuneInString(s[pos:])
		pos += w
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start = pos - w
			if parsingTo {
				res.To = consumeYear()
			} else {
				res.From = consumeYear()
			}
		case '-':
			if peekHas("-tallet") {
				res.Approx = true
				res.To = res.From + 100
			} else {
				parsingTo = true
			}
		case 'c':
			if peekHas("ca") {
				res.Approx = true
			}
		case 'd':
			if peekHas("død ") || peekHas("d. ") {
				parsingTo = true
			}
		case 'f':
			if peekHas("f.kr.") {
				res.From *= -1
				res.To *= -1
			}
		case 't':
			if peekHas("th cent") {
				res.From = (res.From - 1) * 100
				res.To = res.From + 100
				res.Approx = true
			}
		case 'å':
			if peekHas("årh. f.kr") {
				res.From *= -100
				res.To = res.From + 100
				res.Approx = true
			} else if peekHas("årh.") {
				res.From = (res.From - 1) * 100
				res.To = res.From + 100
				res.Approx = true
			}
		default:
			if r == utf8.RuneError { // eof
				return res
			}
		}
	}
	return res
}

func cleanTitle(s string) string {
	s = strings.TrimSuffix(s, " :")
	s = strings.TrimSuffix(s, " : ")
	return s
}

func invertName(s string) string {
	if i := strings.Index(s, ", "); i != -1 {
		return s[i+2:] + " " + s[:i]
	}
	return s
}

func PersonFromAuthority(rec marc.Record) (sirkulator.Resource, error) {
	var (
		res    sirkulator.Resource
		person sirkulator.Person
	)

	res.Type = sirkulator.TypePerson

	if name := rec.ValueAt("100", "a"); name != "" {
		person.Name = invertName(name)
	} // TODO else authority without name should be an error

	if lifespan := rec.ValueAt("100", "d"); lifespan != "" {
		person.YearRange = parseYearRange(lifespan)
	}

	for _, name := range rec.ValuesAt("400", "a") {
		person.NameVariations = append(person.NameVariations, invertName(name))
	}

	for _, d := range rec.DataFieldsAt("024") {
		code := d.ValueAt("2")
		val := d.ValueAt("a")
		if val == "" {
			continue
		}
		switch strings.ToLower(code) {
		case "viaf":
			res.Links = append(res.Links, [2]string{"viaf", strings.TrimPrefix(val, "http://viaf.org/viaf/")})
		case "isni":
			res.Links = append(res.Links, [2]string{"isni", val})
		case "bibbi":
			res.Links = append(res.Links, [2]string{"bibbi", strings.TrimPrefix(val, "https://id.bs.no/bibbi/")})
		case "orcid":
			res.Links = append(res.Links, [2]string{"orcid", val})
		case "no-trbib":
			res.Links = append(res.Links, [2]string{"bibsys", strings.TrimPrefix(val, "x")})
		}
	}

	for _, country := range rec.ValuesAt("043", "c") {
		person.PlaceAssociations = append(person.PlaceAssociations, "marc/"+country)
	}
	if f, ok := rec.DataFieldAt("386"); ok && f.ValueAt("2") == "bs-nasj" {
		for _, nationality := range strings.Split(f.ValueAt("a"), "-") {
			person.PlaceAssociations = append(person.PlaceAssociations, "bs/"+nationality)
		}
	}

	person.Gender = vocab.ParseGender(rec.ValueAt("375", "a"))

	res.Data = person
	res.Label = person.Label()

	return res, nil
}

func CorporationFromAuthority(rec marc.Record) (sirkulator.Resource, error) {
	var (
		res  sirkulator.Resource
		corp sirkulator.Corporation
	)

	res.Type = sirkulator.TypeCorporation

	if name := rec.ValueAt("110", "b"); name != "" {
		// subdivision often is a more descriptinve name than parent corporation name (field a)
		corp.Name = strings.TrimSpace(name)
		// if b is found, a should also be present
		corp.ParentName = strings.TrimSpace(rec.ValueAt("110", "a"))
	} else if name := rec.ValueAt("110", "a"); name != "" {
		corp.Name = strings.TrimSpace(name)
	} // TODO else authority without name should be an error

	for _, d := range rec.DataFieldsAt("024") {
		// TODO code same as in PersonFromAuthority, extract to fn?
		code := d.ValueAt("2")
		val := d.ValueAt("a")
		if val == "" {
			continue
		}
		switch strings.ToLower(code) {
		case "viaf":
			res.Links = append(res.Links, [2]string{"viaf", strings.TrimPrefix(val, "http://viaf.org/viaf/")})
		case "isni":
			res.Links = append(res.Links, [2]string{"isni", val})
		case "bibbi":
			res.Links = append(res.Links, [2]string{"bibbi", strings.TrimPrefix(val, "https://id.bs.no/bibbi/")})
		case "orcid":
			res.Links = append(res.Links, [2]string{"orcid", val})
		case "no-trbib":
			res.Links = append(res.Links, [2]string{"bibsys", strings.TrimPrefix(val, "x")})
		}
	}

	for _, f := range rec.DataFieldsAt("410") {
		if subdivision := f.ValueAt("b"); subdivision != "" {
			corp.NameVariations = append(corp.NameVariations,
				fmt.Sprintf("%s / %s", strings.TrimSpace(subdivision), strings.TrimSpace(f.ValueAt("a"))))
		} else if name := f.ValueAt("a"); name != corp.ParentName && name != corp.Name {
			corp.NameVariations = append(corp.NameVariations, strings.TrimSpace(name))
		}
	}

	res.Data = corp
	res.Label = corp.Label()

	return res, nil
}
