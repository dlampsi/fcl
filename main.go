package main

import (
	"fcl/info"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dlampsi/generigo"
	log "github.com/sirupsen/logrus"
)

var (
	flagMtime            int
	flagSize             float64
	flagPath             string
	flagVerbose          bool
	flagNoColors         bool
	flagSkip             string
	flagCheckMode        bool
	flagVersion          bool
	flagCleanupEmptyDirs bool
)

type osFile struct {
	path string
	info os.FileInfo
}

func main() {
	flag.StringVar(&flagPath, "path", ".", "Working dir path")
	flag.StringVar(&flagSkip, "skip", "", "A comma-separated list of dirs or files for deletion skip (full path)")
	flag.IntVar(&flagMtime, "mtime", 0, "Remove files by age (in days)")
	flag.Float64Var(&flagSize, "size", 0, "Remove files by age (in MB)")
	flag.BoolVar(&flagCheckMode, "check", false, "Run app in check mode. Only list files to delete")
	flag.BoolVar(&flagNoColors, "no-colors", false, "Disable colors in app output")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose output")
	flag.BoolVar(&flagVersion, "version", false, "Prints app version")
	flag.BoolVar(&flagCleanupEmptyDirs, "cleanup-empty-dirs", false, "Cleanup empty dirs after files will be deleted")

	flag.Parse()

	if flagVersion {
		fmt.Println(info.ForPrint())
		os.Exit(0)
	}

	// Logging
	logFormater := &log.TextFormatter{
		DisableLevelTruncation: true,
		DisableColors:          false,
		ForceColors:            true,
		FullTimestamp:          false,
		DisableTimestamp:       true,
	}
	if flagNoColors {
		logFormater.DisableColors = true
	}
	log.SetFormatter(logFormater)
	if flagVerbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stdout)

	// Remove unnessesary slash at the end of path flag
	if strings.HasSuffix(flagPath, "/") {
		flagPath = flagPath[:len(flagPath)-1]
	}
	log.Debugf("Working dir: %s", flagPath)

	ttlDuration := daysToHoursDuration(flagMtime)
	if flagMtime > 0 {
		log.Debug("Searching files older than ", time.Now().Add(-ttlDuration))
	}

	var minSize int64
	if flagSize > 0 {
		minSize = int64(flagSize * 1024 * 1024)
		log.Debugf("Searching files bigger than %d bytes", minSize)
	}

	// Exclusions
	var exclusions []string
	if flagSkip != "" {
		exceptList := strings.Split(flagSkip, ",")
		// Remove unnessesary slash at the end of path
		for _, e := range exceptList {
			if strings.HasSuffix(e, "/") {
				e = e[:len(e)-1]
			}
			exclusions = append(exclusions, e)
		}
	}

	toDel, err := getFilesToDel(flagPath, ttlDuration, minSize, exclusions)
	if err != nil {
		log.Fatalf("Can't get files to delete: %v", err)
	}
	if len(toDel) == 0 {
		log.Info("No files found to delete")
		return
	}
	log.Infof("Found %d files to delete.", len(toDel))

	if flagCheckMode {
		log.Info("List of files:")
		for _, f := range toDel {
			log.WithFields(log.Fields{"size": f.info.Size(), "mod_time": f.info.ModTime()}).Infof("\t%s", f.path)
		}
		return
	}

	var deletedFiles int
	for _, f := range toDel {
		if err := os.Remove(f.path); err != nil {
			log.Errorf("Can't delete %s : %s", f.path, err.Error())
			continue
		}
		deletedFiles++
		log.WithFields(log.Fields{"path": f.path}).Debug("File deleted")
	}
	log.Infof("Deleted %d files", deletedFiles)

	var deletedDirs int
	if flagCleanupEmptyDirs {
		emptyDirs, err := getEmptyDirs(flagPath, exclusions)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Found %d empty dirs to delete.", len(toDel))
		for _, d := range emptyDirs {
			if err := os.Remove(d); err != nil {
				log.Errorf("Can't remove dir %s : %v", d, err.Error())
				continue
			}
			deletedDirs++
		}
		log.Infof("Deleted %d files", deletedFiles)
	}
}

// Returns duration in hours from days numbers.
func daysToHoursDuration(days int) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}

// Returns list of files info for delete.
func getFilesToDel(rootPath string, mtime time.Duration, msize int64, exclusions []string) ([]osFile, error) {
	var toDel []osFile

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var skip bool

		if _, err := filepath.Rel(rootPath, path); err != nil {
			log.Errorf("Got invalid path from filepath.Walk: %s, err: %s", path, err)
			skip = true
		}

		if info.IsDir() {
			if skip {
				return nil
			}
			if generigo.StringInSlice(path, exclusions) {
				return filepath.SkipDir
			}
		}

		if skip {
			return nil
		}

		if info.Mode().IsRegular() {
			if !generigo.StringInSlice(path, exclusions) {
				f := osFile{path: path, info: info}
				if fitsForDelete(f, mtime, msize) {
					toDel = append(toDel, f)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return toDel, nil
}

// Determines if a file is eligible for deletion by modification time and/or min size in bytes.
func fitsForDelete(f osFile, mtime time.Duration, msize int64) bool {
	var fits bool

	if mtime > 0 && msize > 0 {
		if time.Since(f.info.ModTime()) > mtime && f.info.Size() > msize {
			fits = true
		}
	} else if mtime > 0 {
		if time.Since(f.info.ModTime()) > mtime {
			fits = true
		}
	} else if msize > 0 {
		if f.info.Size() > msize {
			fits = true
		}
	}

	log.WithFields(log.Fields{
		"size":     f.info.Size(),
		"mod_time": f.info.ModTime(),
		"fits":     fits,
	}).Debugf("Processed: %s", f.path)

	return fits
}

// Returns list of empty dirs from base root path.
func getEmptyDirs(rootPath string, exclusions []string) ([]string, error) {
	emptyDirs := map[string]bool{}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var skip bool

		if _, err := filepath.Rel(rootPath, path); err != nil {
			log.Errorf("Got invalid path from filepath.Walk: %s, err: %s", path, err)
			skip = true
		}

		if skip {
			return nil
		}

		if info.IsDir() {
			if generigo.StringInSlice(path, exclusions) {
				return filepath.SkipDir
			}
			// Add all dirs, except root dir
			if path != rootPath {
				emptyDirs[path] = true
			}
		} else {
			// Remove dir from empty dirs map if there some files in there
			parent := filepath.Dir(path)
			delete(emptyDirs, parent)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var result []string
	for d := range emptyDirs {
		result = append(result, d)
	}

	return result, nil
}
