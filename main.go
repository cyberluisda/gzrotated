package main

import (
	"path/filepath"
	"io"
	"compress/gzip"
	"io/fs"
	"os"
  "flag"
	"time"
	"fmt"
)

type CandidateFiles struct {
	RootPath string
	Pattern string
	Before time.Duration
}

type Level int
type Log struct {
	Level Level
}
const (
	Verbose Level = iota
	Info
	Quiet
)

var logger *Log

func main() {
	path := flag.String("path", "", "Directory path were start to look for files")
	pattern := flag.String("pattern", "*", "Pattern files to compress")
	before := flag.Duration("before", time.Minute * 15, "Time file was modified/accesed before now to set for gzip compress")
	recursive := flag.Bool("recursive", true, "Scan directories in recursive mode")
	verbose := flag.Bool("verbose", false, "Enable verbosing mode")
	quiet := flag.Bool("quiet", false, "Quiet mode")
	keep := flag.Bool("keep", false, "Keep original files")

	flag.Parse()

	if *path == "" {
		fmt.Fprintln(os.Stderr, "path parameter can not be empty")
		os.Exit(1)
	}

	level := Info
	if *verbose {
		level = Verbose
	}
	if *quiet {
		level = Quiet
	}
	logger = &Log{level}

	cf := &CandidateFiles {
		RootPath: *path,
		Pattern: *pattern,
		Before: *before,
	}

	var files []string
	if *recursive{
		files = cf.ScanRecursive()
	} else {
		panic("Not implemented jet")
	}

	for _, f := range files {
		CompressGZ(f, !(*keep))
	}
}

func (cf *CandidateFiles) ScanRecursive() []string{
	result := make([]string, 0)
	// Scan path
	filepath.WalkDir(
		cf.RootPath,
		func(p string, d fs.DirEntry, e error) error {
			if checkCandidateFile(cf.Pattern, cf.Before, p, d, e) {
				result = append(result, p)
			}
			return nil
		},
	)
	return result
}

func CompressGZ(fileName string, remove bool) error {

	// Source
	src, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("when open source file, %s: %w", fileName, err)
	}
	defer src.Close()

	// Destination
	dstFileName := fmt.Sprintf("%s.gz", fileName)
	dst, err := os.Create(dstFileName)
	if err != nil {
		return fmt.Errorf("when open destination file, %s: %w", fileName, err)
	}
	defer dst.Close()

	logger.log(Info, "Compressing", fileName, "to", dstFileName)

	// Archiver
	archiver := gzip.NewWriter(dst)
	archiver.Name = filepath.Base(fileName)
	defer archiver.Close()

	_, err = io.Copy(archiver, src)
	if err != nil {
		return fmt.Errorf("when compress file %s -> %s: %w", fileName, dstFileName, err)
	}

	if remove {
		err = src.Close()
		if err != nil {
			logger.log(Info, "Error when close orig file", err)
		}

		err := os.Remove(fileName)
		if err != nil {
			return fmt.Errorf("when remove original file %s: %w", fileName, err)
		}
		logger.log(Verbose, "Original file", fileName, "removed")
	} else {
		logger.log(Verbose, "File", fileName, "kept")
	}

	return nil
}

func (l *Log)log(level Level, v ...interface{}) {
	if l.Level != Quiet {
		if l.Level == Verbose || l.Level == level {
			fmt.Println(v...)
		}
	}
}

func checkCandidateFile(pattern string, before time.Duration, p string, d fs.DirEntry, e error) bool {
	now := time.Now()
	if e != nil {
		logger.log(Info, "Ignoring", p, "because we get error:", e)
	} else {
		if ! d.IsDir(){
			base := filepath.Base(p)
			m, err := filepath.Match(pattern, base)

			if err != nil {
				logger.log(Info, "Ignoring", p, "because we get error when match pattern", pattern, ". Err:", err)
			} else if m {
				if filepath.Ext(p) == ".gz" {
					logger.log(Info, "Ignoring", p, "because its extension is '.gz'")
				} else {
					inf, err := d.Info()
					if err != nil {
						logger.log(Info, "Ignoring", p, "because we get error when extract info", err)
					} else {
						modTime := inf.ModTime()
						if modTime.Add(before).Before(now) {
							destName := fmt.Sprintf("%s.gz", p)
							if _, err := os.Stat(destName); !os.IsNotExist(err) {
								logger.log(Info, "Ignoring", p, "because compressed file", destName, "exists")
							} else {
								logger.log(
									Verbose,
									"File", p, "compliant with requirements. Modtime",
										modTime.Format(time.RFC3339),
										"is before more than", before, "from current time:", now.Format(time.RFC3339))
								return true
							}
						} else {
							logger.log(Info, "Ignoring", p, "because file modification time", modTime, "is after that thresold", before)
						}
					}
				}
			}
		}
	}
	return false
}
