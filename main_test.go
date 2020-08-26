package main

import (
	"os"
	"testing"
	"time"

	"github.com/dlampsi/generigo"
)

var (
	testsExclusions = []string{
		"testdata/files_del/.gitkeep",
		"testdata/files_del/subdir/.gitkeep",
		"testdata/files_del/subdir/.gitkeep",
		"testdata/empty_dirs/.gitkeep",
	}

	testFiles = []struct {
		path    string
		daysAgo int
	}{
		{path: "testdata/files_del/today1", daysAgo: 0},
		{path: "testdata/files_del/today2", daysAgo: 0},
		{path: "testdata/files_del/2daysAgo", daysAgo: -2},
		{path: "testdata/files_del/3daysAgo", daysAgo: -3},
		{path: "testdata/files_del/10daysAgo", daysAgo: -10},
		{path: "testdata/files_del/subdir/today3", daysAgo: 0},
		{path: "testdata/files_del/subdir/15daysAgo", daysAgo: -15},
	}

	testEmptryDirs = []string{
		"testdata/empty_dirs/empty1",
		"testdata/empty_dirs/notempty1",
		"testdata/empty_dirs/notempty1/notempty2",
		"testdata/empty_dirs/notempty1/empty2",
	}
	testEmptryDirsFiles = []string{
		"testdata/empty_dirs/notempty1/file1",
		"testdata/empty_dirs/notempty1/file2",
		"testdata/empty_dirs/notempty1/notempty2/file3",
		"testdata/empty_dirs/notempty1/notempty2/file5",
	}
)

// Creates test file with specified modtime - days ago.
func createTestFile(path string, daysAgo int) error {
	if _, err := os.Create(path); err != nil {
		return err
	}
	curTime := time.Now().Local()
	then := curTime.AddDate(0, 0, daysAgo)

	return os.Chtimes(path, curTime, then)
}

func Test_getFilesToDel(t *testing.T) {
	// Prepare test data
	for _, f := range testFiles {
		if err := createTestFile(f.path, f.daysAgo); err != nil {
			t.Fatalf("can't create test file %s : %v", f.path, err)
		}
	}
	defer func() {
		for _, f := range testFiles {
			if err := os.Remove(f.path); err != nil {
				t.Fatalf("can't remove test file %s : %v", f.path, err)
			}
		}
	}()

	f := func(root string, daysMtime int, expectedCount int, exc []string) {
		t.Helper()
		ttlDuration := daysToHoursDuration(daysMtime)
		files, err := getFilesToDel(root, ttlDuration, 0, exc)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != expectedCount {
			t.Fatalf("unexpected lengh of files to delete; want: %d, get: %d - %v", expectedCount, len(files), files)
		}
	}

	f("testdata/files_del", 0, 0,
		testsExclusions)

	f("testdata/files_del", 2, 4,
		testsExclusions)

	f("testdata/files_del", 2, 3,
		append(testsExclusions, "testdata/files_del/subdir"),
	)

	f("testdata/files_del", 2, 2,
		append(testsExclusions, []string{"testdata/files_del/subdir", "testdata/files_del/3daysAgo"}...))

	f("testdata/files_del", 3, 2,
		append(testsExclusions, "testdata/files_del/subdir"))

	f("testdata/files_del", 7, 1,
		append(testsExclusions, "testdata/files_del/subdir"))
}

func Test_getEmptyDirs(t *testing.T) {
	// Prepare test data
	for _, d := range testEmptryDirs {
		if err := os.MkdirAll(d, os.FileMode(0700)); err != nil {
			t.Fatalf("can't create test directory %s : %v", d, err)
		}
	}
	for _, f := range testEmptryDirsFiles {
		if _, err := os.Create(f); err != nil {
			t.Fatalf("can't create test file %s : %v", f, err)
		}
	}
	defer func() {
		for _, f := range testEmptryDirsFiles {
			if err := os.Remove(f); err != nil {
				t.Fatalf("can't remove test file %s : %v", f, err)
			}
		}
		if err := os.Remove("testdata/empty_dirs/notempty1/notempty2"); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove("testdata/empty_dirs/notempty1/empty2"); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove("testdata/empty_dirs/notempty1"); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove("testdata/empty_dirs/empty1"); err != nil {
			t.Fatal(err)
		}
	}()

	f := func(root string, excludeDirs []string, expected []string) {
		t.Helper()
		d, err := getEmptyDirs(root, excludeDirs)
		if err != nil {
			t.Fatal(err)
		}
		if !generigo.CompareStringSlices(d, expected) {
			t.Fatalf("unexpected empty dirs; get: %v, want: %v", d, expected)
		}
	}

	f("testdata/empty_dirs/notempty1/notempty2",
		nil,
		nil)

	f("testdata/empty_dirs",
		nil,
		[]string{"testdata/empty_dirs/empty1", "testdata/empty_dirs/notempty1/empty2"})

	f("testdata/empty_dirs",
		[]string{"testdata/empty_dirs/notempty1"},
		[]string{"testdata/empty_dirs/empty1"})

	f("testdata/empty_dirs/notempty1",
		nil,
		[]string{"testdata/empty_dirs/notempty1/empty2"})
}
