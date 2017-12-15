# be_exam_candidate_ro
This repo contains an exercise intended for Back-End Engineers.

## How-To
This simple daemon will listen for file changes on an input directory, and write any input CSV-files containing Contact
details to the specified output directory.  Errors will be written to the specified error directory.

All command-line arguments are required.

Current usage output is below for reference:

```
$ ./be_exam_candidate_ro --help

Usage of ./be_exam_candidate_ro:
  -e string
    	
  -error-dir string
    	Full directory path where output error files will be written.
  -i string
    	
  -input-dir string
    	Full directory path to watch for new input CSV files.
  -o string
    	
  -output-dir string
    	Full directory path where output JSON translations will be written.
```

## Assumptions
The following assumptions regarding, or derivations from, requirements were made:

* Files will not be considered processed, by name, unless the entire file is tranlated to JSON without error.
* Error output is not written in CSV format, but includes all other informational requirements.
* The "processed file" cache is managed with an internal map that is allowed to grow endlessly (there were no age-out requirements noted).
