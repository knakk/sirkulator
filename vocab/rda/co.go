// Code generated by go generate; DO NOT EDIT.
// This file was generated at 2022-02-10T19:26:30+01:00
// using data from the RDA Registry maintained by the RDA Steering Committee (http://www.rda-rsc.org/)
package rda

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/vocab"
	"golang.org/x/text/language"
)

//go:generate go run gen_codes.go

var labelsCO = map[string][2]string{
	"1001": {"cartographic dataset", "kartografisk datasett"},
	"1002": {"cartographic image", "kartografisk bilde"},
	"1003": {"cartographic moving image", "kartografisk levende bilde"},
	"1004": {"cartographic tactile image", "kartografisk taktilt bilde"},
	"1005": {"cartographic tactile three-dimensional form", "kartografisk taktil tredimensjonal form"},
	"1006": {"cartographic three-dimensional form", "kartografisk tredimensjonal form"},
	"1007": {"computer dataset", "datasett"},
	"1008": {"computer program", "dataprogram"},
	"1009": {"notated movement", "bevegelsesnotasjon"},
	"1010": {"notated music", "nedskrevet musikk"},
	"1024": {"performed movement", ""},
	"1011": {"performed music", "framført musikk"},
	"1012": {"sounds", "lyder"},
	"1013": {"spoken word", "tale"},
	"1014": {"still image", "stillbilde"},
	"1015": {"tactile image", "taktilt bilde"},
	"1017": {"tactile notated movement", "taktil bevegelsesnotasjon"},
	"1016": {"tactile notated music", "taktil musikknotasjon"},
	"1018": {"tactile text", "taktil tekst"},
	"1019": {"tactile three-dimensional form", "taktil tredimensjonal form"},
	"1020": {"text", "tekst"},
	"1021": {"three-dimensional form", "tredimensjonal form"},
	"1022": {"three-dimensional moving image", "tredimensjonalt levende bilde"},
	"1023": {"two-dimensional moving image", "todimensjonalt levende bilde"},
}

var deprecatedCO = map[string]string{}

// CO is a RDA Term Code.
type CO string

func (t CO) Code() string {
	return string(t)
}

func (t CO) URI() string {
	return "http://rdaregistry.info/termList/RDAContentType/" + string(t)
}

func (t CO) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && labelsCO[string(t)][1] != "" {
		return labelsCO[string(t)][1]
	}
	return labelsCO[string(t)][0]
}

func ParseCO(s string) (CO, error) {
	if _, ok := labelsCO[s]; ok {
		return CO(s), nil
	}
	if _, ok := deprecatedCO[s]; ok {
		return CO("unknown"), vocab.ErrDeprecated
	}
	return CO("unknown"), vocab.ErrUnknown
}
