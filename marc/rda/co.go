package rda

import "golang.org/x/text/language"

// RDA Content Type (rdaco)
// https://bibliotekutvikling.no/content/uploads/sites/8/2019/12/RDAContentType_rdaco.pdf
// https://github.com/RDARegistry/RDA-Vocabularies/blob/master/jsonld/termList/RDAContentType.jsonld
// https://github.com/RDARegistry/RDA-Vocabularies/blob/master/nt/termList/RDAContentType.nt
type CO string

const (
	CO1001 CO = "1001"
	CO1002 CO = "1002"
	CO1003 CO = "1003"
	CO1004 CO = "1004"
	CO1005 CO = "1005"
	CO1006 CO = "1006"
)

func (c CO) URI() string {
	return "http://rdaregistry.info/termList/RDAContentType/" + string(c)
}

func (c CO) String() string {
	return string(c)
}

func (c CO) Label(tag language.Tag) string {
	return "TODO"
}
