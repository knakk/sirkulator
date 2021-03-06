// Code generated by go generate; DO NOT EDIT.
// This file was generated at 2022-02-10T19:26:31+01:00
// using data from the RDA Registry maintained by the RDA Steering Committee (http://www.rda-rsc.org/)
package rda

import (
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/vocab"
	"golang.org/x/text/language"
)

//go:generate go run gen_codes.go

var labelsTB = map[string][2]string{
	"1004": {"closed ring binding", ""},
	"1002": {"case binding", ""},
	"1005": {"open ring binding", ""},
	"1001": {"perfect binding", ""},
	"1003": {"spiral binding", ""},
	"1006": {"springback binding", ""},
	"1007": {"saddle stitch binding", ""},
	"1008": {"board book binding", ""},
	"1009": {"slide binding", ""},
	"1010": {"comb binding", ""},
}

var deprecatedTB = map[string]string{}

// TB is a RDA Term Code.
type TB string

func (t TB) Code() string {
	return string(t)
}

func (t TB) URI() string {
	return "http://rdaregistry.info/termList/RDATypeOfBinding/" + string(t)
}

func (t TB) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && labelsTB[string(t)][1] != "" {
		return labelsTB[string(t)][1]
	}
	return labelsTB[string(t)][0]
}

func ParseTB(s string) (TB, error) {
	if _, ok := labelsTB[s]; ok {
		return TB(s), nil
	}
	if _, ok := deprecatedTB[s]; ok {
		return TB("unknown"), vocab.ErrDeprecated
	}
	return TB("unknown"), vocab.ErrUnknown
}
