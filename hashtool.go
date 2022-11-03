package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	hasher      hash.Hash                                //sha256 Hasher
	digitRegexp = regexp.MustCompile(`\d+,[a-zA-Z-]*.+`) // Regular Expression For Matching Lines That Start With Number (Used For Counting Number Of NFTs)
)

// CHIP-0007 Object Model
type chipJson struct {
	Format           string      `json:"format"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	MintingTool      string      `json:"minting_tool"`
	SensitiveContent bool        `json:"sensitive_content"`
	SeriesNumber     int         `json:"series_number"`
	SeriesTotal      int         `json:"series_total"`
	Attributes       []Attribute `json:"attributes"`
	Collection       Collection  `json:"collection"`
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

type Collection struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	Attributes struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
}

func main() {

	// Check Whether The Number Of Arguments To The Program Is Up To 2
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <input.csv>\n", os.Args[0])
		return
	}

	// Splits The FileName Passed To The Program Into An Array Using '.' As A Separator
	ext := strings.Split(os.Args[1], ".")

	// Checks The Last Element Of The Array (Which Should Be The File Extension) And Ensures It's A CSV File
	if strings.ToLower(ext[len(ext)-1]) != "csv" {
		log.Printf("Invalid File Type Specified, CSV File Expected")
		return
	}

	// Instantiates A New Hash
	hasher = sha256.New()

	outputFilename, err := process(strings.Join(ext[:len(ext)-1], "."))

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Successfully Created Output File: %s\n", outputFilename)
}

/*
	process : Reads Records From A CSV File, Converts It To JSON, Hashes The JSON And Writes It To An Output File (At Once) For Memory Efficiency
*/

func process(fileName string) (string, error) {
	//Opens The CSV File For Reading
	file, err := os.Open(fmt.Sprintf("%s.csv", fileName))

	if err != nil {
		return "", fmt.Errorf("An Error Occurred While Opening Input File: %v\n", err)
	}

	// Closes The File When The Program Is Done Running
	defer func() { _ = file.Close() }()

	// Due To How Streams Work In Golang, Once A Reader Has Been Read From, It Is Emptied (i.e A Buffer Cannot Be Read From More Than Once).
	// This Byte Buffer Is Used To Duplicate The File Reader, So There's Two Distinct Readers Containing The Same Thing And Can Be Used Differently.
	var b bytes.Buffer
	dupReader := io.TeeReader(file, &b)

	numLines, err := countLines(dupReader) // Count The Number Of Valid Lines In The CSV (i.e. Skips Blanks And Team Names)

	if err != nil {
		return "", fmt.Errorf("An Error Occurred While Retrieving The Number Of NFTs In File: %v\n", err)
	}

	outputFileName := fmt.Sprintf("%s.output.csv", fileName)
	outputFile, err := os.Create(outputFileName) // Create The Output CSV File With Format inputfilename.output.csv

	if err != nil {
		return "", fmt.Errorf("An Error Occurred While Creating The Output File: %v\n", err)
	}

	reader := csv.NewReader(&b) // Creates A New CSV Reader From The Byte Buffer Duplicated Above.
	reader.FieldsPerRecord = -1 // Allows The CSV Files To Have A Variable Amount Of Columns.

	writer := csv.NewWriter(outputFile) // Creates A New CSV Writer

	header, err := reader.Read() // Reads The Header Of The Input CSV File (Mainly To Check If The File Is Empty)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("Empty CSV File Passed\n")
		}
		return "", fmt.Errorf("An Error Occurred While Reading From Specified CSV File: %v\n", err)
	}

	header = append(header, "Hash") // Append An Extra Column Hash To The Header Above And Write It To The Output File.

	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("An Error Occurred Trying To Write Headers TI Output File: %v\n", err)
	}

	//===========================================================================
	/*
		Due To Earlier Versions Of The Teams' CSV Files, The Position Of Each Header Wasn't Consistent
		Each Teams Put The Position Of Their Headers In Different Places (e.g. Team Bevel Might Put Their Filename Header As The Second Column, While Team Headlight Puts It As The Fourth Column).
		Hence, This Solution To Get The Index Of The Required Headers For Accuracy.

		So The Program Doesn't End Up Using Description For Filename Because It Expected A Consistent Formatting.
	*/
	nftNameIndex := getIndex(header, "name")
	nftDescriptionIndex := getIndex(header, "description")
	nftGenderIndex := getIndex(header, "gender")
	nftAttributeIndex := getIndex(header, "attributes")
	nftSeriesNumberIndex := getIndex(header, "series number")
	nftTeamNameIndex := getIndex(header, "team names")

	//===========================================================================

	teamName := ""
	line := 1

	// Infinite Loop To Read From The CSV File Until EOF Is Reached.
	for {
		record, err := reader.Read()

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("An Error Occurred While Reading From CSV File: %v\n", err)
		}

		// Check Whether The Beginning Of The Line Is A Team-Name And Set It Accordingly
		if tn := record[nftTeamNameIndex]; tn != "" {
			teamName = tn
		}

		// If The Name Of The NFT Isn't Present (As Some Teams Have It In The CSV File), Skip To The Next Record.
		if record[nftNameIndex] == "" {
			if record[nftSeriesNumberIndex] != "" {
				line++
			}
			if err := writer.Write(record); err != nil {
				return "", fmt.Errorf("An Error Occurred While Writing To Destination File: %v\n", err)
			}
			continue
		}

		// Gender Is A Consistent Attribute Across All NFTs
		attributes := []Attribute{
			{
				TraitType: "gender",
				Value:     strings.TrimSpace(record[nftGenderIndex]),
			},
		}

		// Although This Might No Longer Be Needed, But There Were Some Variations Of Each Team's CSVs Which Didn't Have This Column.
		// Hence, The Check If It Exists.
		if nftAttributeIndex > -1 {
			attributes = append(attributes, serializeAttributes(record[getIndex(header, "attributes")])...)
		}

		// Construction Of The Struct With The Necessary Details.
		cJson := chipJson{
			Format:           "CHIP-0007",
			Name:             strings.TrimSpace(record[nftNameIndex]),
			Description:      strings.TrimSpace(record[nftDescriptionIndex]),
			MintingTool:      teamName,
			SensitiveContent: false,
			SeriesNumber:     line,
			SeriesTotal:      numLines,
			Attributes:       attributes,
			Collection: Collection{
				Name: "Zuri NFT Tickets for Free Lunch",
				ID:   "b774f676-c1d5-422e-beed-00ef5510c64d",
				Attributes: struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{Type: "description", Value: "Rewards for accomplishments during HNGi9."},
			},
		}

		line++

		hashed, err := serializeAndHash(cJson)

		if err != nil {
			return "", fmt.Errorf("An Error Occurred While Serializing And Hashing Object: %v\n", err)
		}

		record = append(record, hashed) // Add The SHA256 Hash To The Current Record So It Can Be Written To Output

		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("An Error Occurred While Writing To Destination File: %v\n", err)
		}
	}

	writer.Flush() // Ensure That The Buffer Is Written To File

	return outputFileName, nil
}

/*
serializeAndHash : This Takes In A CHIP-0007 Object, Converts It To JSON And Returns The SHA256 Hashsum
*/
func serializeAndHash(entry chipJson) (string, error) {
	rawMessage, err := json.Marshal(entry) // Convert To Raw JSON

	if err != nil {
		return "", err
	}

	fmt.Println(string(rawMessage))

	hasher.Write(rawMessage)

	h := hasher.Sum(nil)

	hasher.Reset()

	return strings.ToUpper(fmt.Sprintf("%x", h)), nil

}

/*
getIndex : Get The Index Of A Column From A Provided Header
*/
func getIndex(header []string, key string) int {
	for index, elem := range header {
		if strings.ToLower(elem) == key {
			return index
		}
	}

	log.Printf("No Column Named: %v\n", key)
	return -1
}

/*
serializeAttributes : This Takes In The String From The Attribute Column And Parses It To Get The Attribute And Value
e.g. Converts 'Teeth Color: Brown' To Attribute{Trait Type: Teeth Color, Value: Brown}
*/
func serializeAttributes(attributes string) []Attribute {
	var result []Attribute
	attributes = strings.NewReplacer("; ", ";", " ;", ";", ",", ";", ", ", ";").Replace(attributes)
	attrs := strings.Split(attributes, ";")

	if len(attrs) == 0 {
		return nil
	}

	for _, attribute := range attrs {
		attributeAndValue := strings.Split(attribute, ":")

		if len(attributeAndValue) != 2 {
			continue
		}

		result = append(result, Attribute{TraitType: strings.TrimSpace(attributeAndValue[0]), Value: strings.Trim(strings.TrimSpace(attributeAndValue[1]), ",")})
	}

	return result
}

/*
countLines : Counts The Number Of Valid Lines In A CSV File.
*/
func countLines(r io.Reader) (int, error) {
	var count int

	buf := make([]byte, 64*1024)

	for {
		bufferSize, err := r.Read(buf)

		count += bytes.Count(buf[:bufferSize], []byte{'\n'})

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}
	}

	fmt.Println(count)

	return count, nil
}
