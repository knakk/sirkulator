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

var labelsCT = map[string][2]string{
	"1021": {"aperture card", "vinduskort"},
	"1070": {"audio belt", "audio belt"},
	"1002": {"audio cartridge", "lyd-cartridge"},
	"1003": {"audio cylinder", "fonografrull"},
	"1004": {"audio disc", "lydplate"},
	"1006": {"audio roll", "pianorull"},
	"1071": {"audio wire reel", "lydtrådspole"},
	"1007": {"audiocassette", "lydkassett"},
	"1008": {"audiotape reel", "lydbåndspole"},
	"1045": {"card", "kort"},
	"1011": {"computer card", "datakort"},
	"1012": {"computer chip cartridge", "kretskortkassett"},
	"1013": {"computer disc", "dataplate"},
	"1014": {"computer disc cartridge", "dataplatekassett"},
	"1015": {"computer tape cartridge", "databånd-cartridge"},
	"1016": {"computer tape cassette", "databåndkassett"},
	"1017": {"computer tape reel", "databåndspole"},
	"1032": {"film cartridge", "film-cartridge"},
	"1033": {"film cassette", "filmkassett"},
	"1034": {"film reel", "filmspole"},
	"1069": {"film roll", "filmrull"},
	"1035": {"filmslip", "filmstrimmel"},
	"1036": {"filmstrip", "filmremse"},
	"1037": {"filmstrip cartridge", "filmremsekassett"},
	"1046": {"flipchart", "flippover"},
	"1022": {"microfiche", "mikrofilmkort"},
	"1023": {"microfiche cassette", "mikrofilmkortkassett"},
	"1024": {"microfilm cartridge", "mikrofilm-cartridge"},
	"1025": {"microfilm cassette", "mikrofilmkassett"},
	"1026": {"microfilm reel", "mikrofilmspole"},
	"1056": {"microfilm roll", "mikrofilmrull"},
	"1027": {"microfilm slip", "mikrofilmremse"},
	"1028": {"microopaque", "mikro-opak"},
	"1030": {"microscope slide", "mikroskopdia"},
	"1059": {"object", "gjenstand"},
	"1018": {"online resource", "online (nettilkoblet) ressurs"},
	"1039": {"overhead transparency", "overheadtransparent"},
	"1047": {"roll", "rull"},
	"1048": {"sheet", "ark"},
	"1040": {"slide", "lysbilde"},
	"1005": {"sound-track reel", "lydfilmspole"},
	"1042": {"stereograph card", "stereobilde"},
	"1043": {"stereograph disc", "stereografisk plate"},
	"1051": {"video cartridge", "video-cartridge"},
	"1052": {"videocassette", "videokassett"},
	"1060": {"videodisc", "videodisk"},
	"1053": {"videotape reel", "videobåndspole"},
	"1049": {"volume", "bind"},
}

var deprecatedCT = map[string]string{
	"1001": "Audio carriers (Deprecated)",
	"1010": "Computer carriers (Deprecated)",
	"1020": "Microform carriers (Deprecated)",
	"1029": "Microscopic carriers (Deprecated)",
	"1031": "Projected image carriers (Deprecated)",
	"1041": "Stereographic carriers (Deprecated)",
	"1044": "Unmediated carriers (Deprecated)",
	"1050": "Video carriers (Deprecated)",
}

// CT is a RDA Term Code.
type CT string

func (t CT) Code() string {
	return string(t)
}

func (t CT) URI() string {
	return "http://rdaregistry.info/termList/RDACarrierType/" + string(t)
}

func (t CT) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && labelsCT[string(t)][1] != "" {
		return labelsCT[string(t)][1]
	}
	return labelsCT[string(t)][0]
}

func ParseCT(s string) (CT, error) {
	if _, ok := labelsCT[s]; ok {
		return CT(s), nil
	}
	if _, ok := deprecatedCT[s]; ok {
		return CT("unknown"), vocab.ErrDeprecated
	}
	return CT("unknown"), vocab.ErrUnknown
}
