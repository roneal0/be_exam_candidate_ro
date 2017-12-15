package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const (
	// flag descriptions, for convenience
	HELP_INPUT_DIR  = "Full directory path to watch for new input CSV files."
	HELP_OUTPUT_DIR = "Full directory path where output JSON translations will be written."
	HELP_ERROR_DIR  = "Full directory path where output error files will be written."
)

// this type will pair a CSV record with any associated validation errors
type CsvErrorRecord struct {
	// line on which this record appears in the parent file
	Line int

	// fields parsed from the record
	Fields []string

	// all errors associated with the record
	Errors []string
}

// Creates a newly initialized CsvErrorRecord, ready to have error strings
// appended to its internal Errors slice
func NewCsvErrorRecord(line int, fields []string) *CsvErrorRecord {
	return &CsvErrorRecord{
		Line:   line,
		Fields: fields,
		Errors: make([]string, 0),
	}
}

// entry point
func main() {
	// define and parse command-line arguments
	argInputDir := flag.String("i", "", "")
	flag.StringVar(argInputDir, "input-dir", "", HELP_INPUT_DIR)
	argOutputDir := flag.String("o", "", "")
	flag.StringVar(argOutputDir, "output-dir", "", HELP_OUTPUT_DIR)
	argErrorDir := flag.String("e", "", "")
	flag.StringVar(argErrorDir, "error-dir", "", HELP_ERROR_DIR)
	flag.Parse()

	fail := false
	if len(*argInputDir) == 0 {
		ERROR("An input directory must be provided.")
		fail = true
	} else if len(*argOutputDir) == 0 {
		ERROR("An output directory must be supplied.")
		fail = true
	} else if len(*argErrorDir) == 0 {
		ERROR("An error directory must be supplied.")
		fail = true
	}

	if fail {
		print("\n")
		flag.Usage()
		os.Exit(1)
	}

	// pre-process inputs only slightly
	inputDir := *argInputDir
	outputDir := strings.TrimSuffix(*argOutputDir, string(filepath.Separator))
	errorDir := strings.TrimSuffix(*argErrorDir, string(filepath.Separator))

	// confirm the input and output directories exist.
	if _, err := os.Stat(inputDir); err != nil && os.IsNotExist(err) {
		ERROR("Input directory [ ", inputDir, " ] doesn't exist.")
		os.Exit(1)
	} else if _, err := os.Stat(outputDir); err != nil && os.IsNotExist(err) {
		ERROR("Output directory [ ", outputDir, " ] doesn't exist.")
		os.Exit(1)
	} else if _, err := os.Stat(errorDir); err != nil && os.IsNotExist(err) {
		ERROR("Error directory [ ", errorDir, " ] doesn't exist.")
		os.Exit(1)
	}

	if err := os.Chdir(inputDir); err != nil {
		ERROR("Failed to change directory to [ ", inputDir, " ].")
		os.Exit(1)
	}

	// let's establish an "inotify" watch on the input directory
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ERROR("Error creating watcher: ", err.Error())
		os.Exit(1)
	}

	// watch out for os.exit...not that it matters much here on shutdown
	defer watcher.Close()

	// we'll use this channel to close the watcher on exit
	done := make(chan bool)

	// this map will keep track of file names we've seen (depending on usage,
	// this should probably be offloaded to an external key/value store)
	fileMap := make(map[string]struct{})

	// meat and potatoes below
	go func() {
		for {
			select {
			// capture any inbound events
			case ev := <-watcher.Events:
				INFO("Received fsnotify event: ", ev)

				// we'll only concern ourselves with CSV files
				if !strings.HasSuffix(ev.Name, ".csv") {
					break
				}

				// have we seen this file before?
				if _, ok := fileMap[ev.Name]; ok {
					break
				}

				INFO("Processing file [ ", ev.Name, " ].")

				// capture this file in our map (which will also act as a lock
				// in the case of duplicate fsnotify events)
				fileMap[ev.Name] = struct{}{}

				// process the file in a new goroutine
				go func() {
					shouldDelete := false

					stat, err := os.Stat(ev.Name)
					filename := strings.TrimSuffix(ev.Name, ".csv") + ".json"
					if err != nil {
						ERROR("Failed to get status for file [ ", ev.Name, " ].")
					} else {
						// translate the input file
						if contacts, err := ParseCsvContactData(ev.Name); err != nil {
							ERROR(err.Error())
						} else {
							// we'll only consider this processed if there were more than zero records parsed successfully
							if len(contacts.Records) > 0 {
								byteData, err := json.MarshalIndent(contacts.Records, "", "    ")
								if err != nil {
									ERROR("Error marshalling file [ ", ev.Name, " ] records to json: ", err.Error())
								} else {
									ioutil.WriteFile(outputDir+"/"+filename, byteData, stat.Mode())
									shouldDelete = true
								}
							}

							// write any errors
							if len(contacts.Errors) > 0 {
								byteData, err := json.MarshalIndent(contacts.Errors, "", "    ")
								if err != nil {
									ERROR("Error marshalling file [ ", ev.Name, " ] errors to json: ", err.Error())
								} else {
									ioutil.WriteFile(errorDir+"/"+filename, byteData, stat.Mode())
								}
							}
						}

						// remove the input file (only if processed)
						if shouldDelete {
							//os.Remove(ev.Name)
						} else {
							// remove this file from the map...we didn't process it successfully
							delete(fileMap, ev.Name)
						}
					}
				}()
			case err := <-watcher.Errors:
				ERROR("Error during watch: ", err.Error())
			case _ = <-done:
				return
			}
		}
	}()

	// watch the input directory for any file changes
	if err := watcher.Add("."); err != nil {
		ERROR("Error trying to watch input directory: ", err.Error())
		return
	}

	// sit and wait for signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c

	// kill our watcher
	done <- true
	INFO("Exiting...")
}
