package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type banDataType struct {
	UserName string `json:"username"`
	Reason   string `json:"reason,omitempty"`
}

func main() {
	var composite []banDataType
	exe, _ := os.Executable()

	//Help
	if len(os.Args) <= 1 {
		fmt.Println("Usage: " + filepath.Base(exe) + " <file1> <file2> ...")
		fmt.Print("Output file: composite.json\n")
		os.Exit(1)
	}

	//Get list of files
	filesToProcess := os.Args[1:]

	//Loop through files
	for _, file := range filesToProcess {

		file, err := os.Open(file)

		if err != nil {
			log.Println(file, err)
			return
		}
		defer file.Close()

		var bData []banDataType

		data, err := ioutil.ReadAll(file)

		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		/* This area deals with 'array of strings' format */
		var names []string
		err = json.Unmarshal(data, &names)

		if err != nil {
			//Not really an error, just empty array
			//Only needed because Factorio will write some bans as an array for some unknown reason.
		} else {

			for _, name := range names {
				if name != "" {
					bData = append(bData, banDataType{UserName: strings.ToLower(name)})
				}
			}
		}

		/* This area deals with standard format bans */
		var bans []banDataType
		_ = json.Unmarshal(data, &bans)

		for _, item := range bans {
			if item.UserName != "" {
				bData = append(bData, banDataType{UserName: strings.ToLower(item.UserName), Reason: item.Reason})
			}
		}
		log.Println("Read " + fmt.Sprintf("%v", len(bData)) + " bans from banlist.")

		/* Combine all the data, dealing with duplicates */
		dupes := 0
		diffDupes := 0
		for apos, aBan := range bData {
			found := false
			dupReason := ""
			for bpos, bBan := range bData {
				if strings.EqualFold(aBan.UserName, bBan.UserName) && apos != bpos {
					found = true
					dupReason = bBan.Reason
					dupes++
					break
				}

			}
			if !found {
				composite = append(composite, aBan)
			} else {
				if !strings.EqualFold(aBan.Reason, dupReason) && !strings.HasPrefix(dupReason, "[dup]") {
					if aBan.Reason != "" && dupReason != "" {
						bData[apos].Reason = "[dup] " + aBan.Reason + ", " + dupReason
					} else {
						bData[apos].Reason = "[dup] " + aBan.Reason + dupReason
					}
					diffDupes++
					composite = append(composite, bData[apos])
				}
			}
		}

		log.Printf("Removed %v duplicates from banlist, %v dupes had multiple reasons (reasons combined)\n", dupes, diffDupes)

	}

	/* Write out composite banlist */
	file, err := os.Create("composite.json")

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	err = enc.Encode(composite)

	if err != nil {
		log.Println("Error encoding ban list file: " + err.Error())
		os.Exit(1)
	}

	wrote, err := file.Write(outbuf.Bytes())

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Printf("Wrote banlist (%v) of %v bytes.\n", len(composite), wrote)

}
