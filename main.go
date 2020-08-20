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
	flagMtime      int
	flagSize       float64
	flagPath       string
	flagVerbose    bool
	flagNoColors   bool
	flagSkipDirs   string
	flagSkipFiles  string
	flagCheckMode  bool
	flagVersion    bool
	exceptionPaths []string
	ttlDuration    time.Duration
	minSize        int64
	toDel          []fileToDel
)

type fileToDel struct {
	path string
	info os.FileInfo
}

func main() {
	flag.StringVar(&flagPath, "path", ".", "Working dir path")
	flag.StringVar(&flagSkipDirs, "skip-dirs", "", "A comma-separated list of dirs for deletion skip")
	flag.StringVar(&flagSkipFiles, "skip-files", "", "A comma-separated list of full path to files for deletion skip")
	flag.IntVar(&flagMtime, "mtime", 0, "Remove files by age (in days)")
	flag.Float64Var(&flagSize, "size", 0, "Remove files by age (in MB)")
	flag.BoolVar(&flagCheckMode, "check", false, "Run app in check mode. Only list files to delete")
	flag.BoolVar(&flagNoColors, "no-colors", false, "Disable colors in app output")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose output")
	flag.BoolVar(&flagVersion, "version", false, "Prints app version")

	flag.Parse()

	if flagVersion {
		fmt.Println(info.ForPrint())
		os.Exit(0)
	}

	// Configure logging
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

	// Exceptions dirs
	if flagSkipDirs != "" {
		exceptList := strings.Split(flagSkipDirs, ",")
		// Remove unnessesary slash at the end of path
		for _, e := range exceptList {
			if strings.HasSuffix(e, "/") {
				e = e[:len(e)-1]
			}
			exceptionPaths = append(exceptionPaths, e)
		}
	}
	log.Debugf("Exception dirs: %v", exceptionPaths)

	ttlDuration = time.Duration(flagMtime) * 24 * time.Hour
	if flagMtime > 0 {
		log.Debug("Searching files older than ", time.Now().Add(-ttlDuration))
	}
	if flagSize > 0 {
		minSize = int64(flagSize * 1024 * 1024)
		log.Debugf("Searching files bigger than %d bytes", minSize)
	}

	if err := filepath.Walk(flagPath, walkFunc); err != nil {
		log.Fatal(err)
	}

	log.Infof("Found %d files for delete.", len(toDel))

	if len(toDel) == 0 {
		log.Info("Nothing to delete.")
		return
	}

	if flagCheckMode {
		log.Info("List of files:")
		for _, f := range toDel {
			log.WithFields(log.Fields{"size": f.info.Size(), "mod_time": f.info.ModTime()}).Infof("\t%s", f.path)
		}
		return
	}

	for _, f := range toDel {
		if err := os.Remove(f.path); err != nil {
			log.Errorf("Can't delete %s : %s", f.path, err.Error())
			continue
		}
		log.WithFields(log.Fields{"path": f.path}).Debug("File deleted")
	}
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	var skip bool

	relPath, err := filepath.Rel(flagPath, path)
	if err != nil {
		log.Errorf("Got invalid path from filepath.Walk: %s, err: %s", path, err)
		skip = true
	}

	if info.IsDir() {
		if skip {
			return nil
		}
		if generigo.StringInSlice(relPath, exceptionPaths) {
			// log.Debugf("Skipping dir: %s", path)
			return filepath.SkipDir
		}
	}

	if skip {
		return nil
	}

	if info.Mode().IsRegular() {
		if fitsToDelete(path, info) {
			toDel = append(toDel, fileToDel{path: path, info: info})
		}
	}

	return nil
}

func fitsToDelete(path string, info os.FileInfo) bool {
	var fits bool
	var skipList []string

	if flagSkipFiles != "" {
		skipList = strings.Split(flagSkipFiles, ",")
	}

	if generigo.StringInSlice(path, skipList) {
		fits = false
	} else if flagMtime > 0 && flagSize > 0 {
		if time.Since(info.ModTime()) > ttlDuration && info.Size() > minSize {
			fits = true
		}
	} else if flagMtime > 0 {
		if time.Since(info.ModTime()) > ttlDuration {
			fits = true
		}
	} else if flagSize > 0 {
		if info.Size() > minSize {
			fits = true
		}
	}
	log.WithFields(log.Fields{"size": info.Size(), "mod_time": info.ModTime(), "fits": fits}).Debugf("Processing: %s", path)
	return fits
}
