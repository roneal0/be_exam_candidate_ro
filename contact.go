package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"regexp"
	"strconv"
)

const (
	// expected CSV column index order
	COL_INTERNAL_ID = iota
	COL_FIRST_NAME
	COL_MIDDLE_NAME
	COL_LAST_NAME
	COL_PHONE_NUM
)

// This type will wrap up some contact information
type Contact struct {
	Id     int    `json:"id"`
	First  string `json:"first"`
	Middle string `json:"middle,omitempty"` // optional
	Last   string `json:"last"`
	Phone  string `json:"phone"`
}

// a simple type to house parsed CSV data, and any errors
type ContactCsvData struct {
	Records []*Contact
	Errors  []*CsvErrorRecord
}

// returns a newly initialized ContactCsvData instance
func NewContactCsvData() *ContactCsvData {
	return &ContactCsvData{
		Records: make([]*Contact, 0),
		Errors:  make([]*CsvErrorRecord, 0),
	}
}

// Opens filename, which is expected to be a CSV-file (with a header record) of
// Contact type fields. The expected order of these fields is defined by the COL_*
// constant set.
func ParseCsvContactData(filename string) (*ContactCsvData, error) {
	// try to load the input CSV file
	file, err := os.Open(filename)
	if err != nil {
		ERROR("Failed to open input file [ ", filename, " ].")
		return nil, err
	}

	i := 0
	result := NewContactCsvData()
	reader := csv.NewReader(bufio.NewReader(file))
	for {
		records, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			ERROR("Error reading record from input file [ ", filename, " ].")
			return nil, err
		}

		// we'll assume we need to skip the header record
		if i == 0 {
			i++
			continue
		}
		i++

		// capture any errors
		tmpError := NewCsvErrorRecord(i, records)

		// validate the input record
		fail := false
		isMatch, err := regexp.MatchString("[0-9]{8}", records[COL_INTERNAL_ID])
		if err != nil {
			ERROR("Error during regexp processing for file [ ", filename, " ]")
			return nil, err
		}

		if !isMatch {
			fail = true
			tmpError.Errors = append(tmpError.Errors, "INTERNAL_ID must be an 8-digit integer.")
		}

		id, err := strconv.ParseInt(records[COL_INTERNAL_ID], 10, 64)
		if err != nil {
			fail = true
			tmpError.Errors = append(tmpError.Errors, "Failed to parse PHONE_NUM: "+err.Error())
		}

		// validate the phone number
		isMatch, err = regexp.MatchString("[0-9]{3}-[0-9]{3}-[0-9]{4}", records[COL_PHONE_NUM])
		if err != nil {
			ERROR("Error during regexp processing for file [ ", filename, " ]")
			return nil, err
		}

		if !isMatch {
			fail = true
			tmpError.Errors = append(tmpError.Errors, "PHONE_NUM must be in the format ###-###-####.")
		}

		// if there were any validation errors, add the current CsvErrorRecord
		if fail {
			result.Errors = append(result.Errors, tmpError)
			continue
		}

		// maximum length of 15
		if len(records[COL_FIRST_NAME]) >= 14 {
			records[COL_FIRST_NAME] = records[COL_FIRST_NAME][0:14]
		}
		if len(records[COL_MIDDLE_NAME]) >= 14 {
			records[COL_MIDDLE_NAME] = records[COL_MIDDLE_NAME][0:14]
		}
		if len(records[COL_LAST_NAME]) >= 14 {
			records[COL_LAST_NAME] = records[COL_LAST_NAME][0:14]
		}

		// build a new Contact instance from the successfully parsed data
		c := new(Contact)
		c.Id = int(id)
		c.First = records[COL_FIRST_NAME]
		c.Middle = records[COL_MIDDLE_NAME]
		c.Last = records[COL_LAST_NAME]
		c.Phone = records[COL_PHONE_NUM]

		// add the decoded record
		result.Records = append(result.Records, c)
	}
	return result, nil
}
