package marc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
)

// Unmarshal parsers the MARCXML-encoded data and stores the result in the given Record.
func Unmarshal(b []byte, rec *Record) error {
	dec := xml.NewDecoder(bytes.NewBuffer(b))
	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		switch elem := t.(type) {
		case xml.SyntaxError:
			return errors.New(elem.Error())
		case xml.StartElement:
			if elem.Name.Local == "record" {
				return dec.DecodeElement(&rec, &elem)
			}
		}
	}
}

// MustParse parses the give MARCXML-encoded data and records a Record, but panics on errors.
func MustParse(b []byte) Record {
	var r Record
	if err := Unmarshal(b, &r); err != nil {
		panic(err)
	}
	return r
}

// MustParseString is like MustParse, except it accepts a string instead.
func MustParseString(s string) Record {
	var r Record
	if err := Unmarshal([]byte(s), &r); err != nil {
		panic(err)
	}
	return r
}

// Decoder can decode MARC records from a stream.
type Decoder struct {
	xmlDec *xml.Decoder
}

// NewDecoder returns a new Decoder for the given reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		xmlDec: xml.NewDecoder(r),
	}
}

// DecodeAll consumes the input stream and returns all decoded records.
// If there is an error, it will return, together with the succesfully
// parsed MARC records up til then.
func (d *Decoder) DecodeAll() ([]Record, error) {
	res := make([]Record, 0)
	for r, err := d.Decode(); err != io.EOF; r, err = d.Decode() {
		if err != nil {
			return res, err
		}
		res = append(res, r)
	}
	return res, nil
}

// Decode decodes and returns a single MARC Record, or and error.
func (d *Decoder) Decode() (Record, error) {
	var rec Record
	for {
		t, err := d.xmlDec.Token()
		if err != nil {
			return rec, err
		}
		switch elem := t.(type) {
		case xml.SyntaxError:
			return rec, errors.New(elem.Error())
		case xml.StartElement:
			if elem.Name.Local == "record" {
				err := d.xmlDec.DecodeElement(&rec, &elem)
				return rec, err
			}
		}
	}
}
