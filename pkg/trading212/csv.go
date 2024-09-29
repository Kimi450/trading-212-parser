package trading212

// https://stackoverflow.com/questions/24999079/reading-csv-file-in-go

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"

	"github.com/ansel1/merry/v2"
)

type Scanner struct {
	Reader *csv.Reader
	Head   []string
	Row    []string
}

func NewScanner(o io.Reader) Scanner {
	csv_o := csv.NewReader(o)
	header, err := csv_o.Read()
	if err != nil {
		return Scanner{}
	}
	return Scanner{Reader: csv_o, Head: header}
}

func (o *Scanner) Scan() bool {
	row, err := o.Reader.Read()
	if err != nil {
		return false
	}
	o.Row = row
	return true
}

func (o *Scanner) toJson() string {
	json := ""
	for i, value := range o.Row {
		value = strings.TrimSpace(value)
		if i != 0 {
			json += ",\n"
		}
		json += "\"" + o.Head[i] + "\":\"" + value + "\""
	}
	json = "{\n" + json + "\n}"
	return json
}

func (o *Scanner) ToRecord() (Record, error) {
	jsonString := []byte(o.toJson())
	record := Record{}
	if err := json.Unmarshal(jsonString, &record); err != nil {
		return record, merry.Errorf("failed to unmarshall json to record: %w", err)
	}

	return record, nil
}
