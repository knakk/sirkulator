// Package marc implements decoding of MARCXML (MarcXchange (ISO25577)
// bibliographic MARC records, and convenience methods for extracting
// values from record fields.
package marc

import (
	"encoding/xml"
)

// Record represents a MARC record.
type Record struct {
	XMLName       xml.Name       `xml:"record"`
	Leader        string         `xml:"leader"` // 24 chars
	ControlFields []ControlField `xml:"controlfield"`
	DataFields    []DataField    `xml:"datafield"`
}

// ControlField represents a control field in a MARC record.
type ControlField struct {
	Tag   string `xml:"tag,attr"`  // 3 chars
	Value string `xml:",chardata"` // if Tag == "000"; 40 chars
}

// DataField represents a data field in a MARC record.
type DataField struct {
	Tag       string     `xml:"tag,attr"`  // 3 chars
	Ind1      string     `xml:"ind1,attr"` // 1 char
	Ind2      string     `xml:"ind2,attr"` // 1 char
	SubFields []SubField `xml:"subfield"`
}

// ValueAt returns the first value at the first occurrence
// of the code in the the DataField. It returns an empty
// string if none is found.
func (d DataField) ValueAt(code string) string {
	for _, f := range d.SubFields {
		if f.Code == code {
			return f.Value
		}
	}

	return ""
}

// ValuesAt returns a string slice of all occurrences of
// code in the DataField.
func (d DataField) ValuesAt(code string) (res []string) {
	for _, f := range d.SubFields {
		if f.Code == code {
			res = append(res, f.Value)
		}
	}

	return res
}

// SubField represents a sub field in a data field.
type SubField struct {
	Code  string `xml:"code,attr"` // 1 char
	Value string `xml:",chardata"`
}

// IsEmpty returns true if the MARC Record i empty (zero value).
func (r Record) IsEmpty() bool {
	return r.Leader == "" && len(r.ControlFields) == 0 && len(r.DataFields) == 0
}

// ValueAt returns the first occurrence of code in the requested DataField, along
// with a boolean denoting if a value was found or not.
func (r Record) ValueAt(tag, code string) (string, bool) {
	for _, f := range r.DataFields {
		if f.Tag == tag {
			for _, sf := range f.SubFields {
				if sf.Code == code {
					return sf.Value, true
				}
			}
		}
	}

	return "", false
}

// ValuesAt returns a string slice of all occurrences of code in
// all occurrences of DataFields of the given tag.
func (r Record) ValuesAt(tag, code string) (res []string) {
	for _, f := range r.DataFields {
		if f.Tag == tag {
			for _, sf := range f.SubFields {
				if sf.Code == code {
					res = append(res, sf.Value)
				}
			}
		}
	}

	return res
}

// ControlFieldAt returns the control fiel of the given tag, along
// with a boolean denoting if it was found or not.
func (r Record) ControlFieldAt(tag string) (ControlField, bool) {
	for _, f := range r.ControlFields {
		if f.Tag == tag {
			return f, true
		}
	}

	return ControlField{}, false
}

// DataFieldAt returns the first occurrence of DataField with
// the given tag, along with a boolean denoting if it was found or not.
func (r Record) DataFieldAt(tag string) (DataField, bool) {
	for _, f := range r.DataFields {
		if f.Tag == tag {
			return f, true
		}
	}

	return DataField{}, false
}

// DataFieldsAt returns all occurrences of DataFields with
// the given tags.
func (r Record) DataFieldsAt(tags ...string) (res []DataField) {
	for _, tag := range tags {
		for _, f := range r.DataFields {
			if f.Tag == tag {
				res = append(res, f)
			}
		}
	}

	return res
}
