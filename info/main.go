package info

import "fmt"

// Full application name.
var AppName = "fcl"

// Short application name (namespace).
var NameSpace = "fcl"

// Version number.
var Version = "dev"

// App build number
var BuildNumber = ""

// BuildTime label of build time.
var BuildTime = ""

// Repository commit hash from HEAD on build branch.
var CommitHash = ""

// ForPrint Returns formated version string for print.
func ForPrint() string {
	return fmt.Sprintf("%s %s\nBuild time %s\n", AppName, Version, BuildTime)
}
