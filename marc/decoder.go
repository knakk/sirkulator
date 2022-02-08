package marc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// Unmarshal parsers the MARCXML-encoded data and stores the result in the given Record.
func Unmarshal(b []byte, rec *Record) error {
	dec := xml.NewDecoder(bytes.NewBuffer(b))

	for {
		t, err := dec.Token()
		if err != nil {
			return fmt.Errorf("marc: Unmarshal: %w", err)
		}

		switch elem := t.(type) {
		case xml.SyntaxError:
			return fmt.Errorf("marc: Unmarshal: %s", elem.Error())
		case xml.StartElement:
			if elem.Name.Local == "record" {
				if err := dec.DecodeElement(&rec, &elem); err != nil {
					return fmt.Errorf("marc: Unmarshal: %w", err)
				}

				return nil
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
// If there is an error, it will return, together with the successfully
// parsed MARC records up til then.
func (d *Decoder) DecodeAll() ([]Record, error) {
	res := make([]Record, 0)
	for r, err := d.Decode(); !errors.Is(err, io.EOF); r, err = d.Decode() {
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
			return rec, fmt.Errorf("marc: Decoder.Decode: %w", err)
		}
		switch elem := t.(type) {
		case xml.SyntaxError:
			return rec, fmt.Errorf("marc: Decoder.Decode: %s", elem.Error())
		case xml.StartElement:
			if elem.Name.Local == "record" {
				if err := d.xmlDec.DecodeElement(&rec, &elem); err != nil {
					return rec, fmt.Errorf("marc: Decoder.Decode: %w", err)
				}

				return rec, nil
			}
		}
	}
}
