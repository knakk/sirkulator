package oai

import (
	"time"

	"github.com/knakk/sirkulator/marc"
)

// RemoteRecord represents a OAI record retrieved from an OAI/PMH repository.
type RemoteRecord struct {
	Header struct {
		Status     string    `xml:"status,attr"`
		Identifier string    `xml:"identifier"`
		Datestamp  time.Time `xml:"datestamp"`
	} `xml:"header"`
	Metadata []byte `xml:",innerxml"` // marcxml
}

type listRecordsResponse struct {
	Error struct {
		Code    string `xml:"code,attr"`
		Message string `xml:",chardata"`
	} `xml:"error"`
	ListRecords struct {
		Records         []RemoteRecord `xml:"record"`
		ResumptionToken string         `xml:"resumptionToken"`
	} `xml:"ListRecords,omitempty"`
}

type getRecordResponse struct {
	Error struct {
		Code    string `xml:"code,attr"`
		Message string `xml:",chardata"`
	} `xml:"error"`
	GetRecord struct {
		Record          RemoteRecord `xml:"record"`
		ResumptionToken string       `xml:"resumptionToken"`
	} `xml:"GetRecord,omitempty"`
}

// ProcessedRecord represent a OAI Record which has been processed extracting
// relevant information needed for storage in DB and indexing in search index.
type ProcessedRecord struct {
	DBRecord

	// For search indexing:
	// TODO maybe remove these, not really needed except for testing now
	Type  string
	Label string

	// [0] = key (isbn/ismn/viaf), [1] = id
	Identifiers [][2]string
}

type ProcessFunc func(RemoteRecord) (ProcessedRecord, error)

// DBRecord represents an OAI record as it is stored in local DB.
type DBRecord struct {
	Source     string
	ID         string
	Data       []byte // gzipped XML
	NewData    []byte // gzipped XML
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt time.Time
}

type Record struct {
	Source string
	ID     string
	Data   marc.Record
}
