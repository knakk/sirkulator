package rda

import "golang.org/x/text/language"

// RDA Type Of Binding (rdatb)
type TB string

const (
	TB1001 TB = "1001"
	TB1002 TB = "1002"
	TB1003 TB = "1003"
	TB1004 TB = "1004"
	TB1005 TB = "1005"
	TB1006 TB = "1006"
	TB1007 TB = "1007"
	TB1008 TB = "1008"
	TB1009 TB = "1009"
	TB1010 TB = "1010"
)

func (t TB) URI() string {
	return "http://rdaregistry.info/termList/RDATypeOfBinding/" + string(t)
}

func (t TB) String() string {
	return string(t)
}

func (t TB) Label(tag language.Tag) string {
	return "TODO"
}
