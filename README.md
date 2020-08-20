# fcl

[![Actions Status](https://github.com/dlampsi/fcl/workflows/default/badge.svg)](https://github.com/dlampsi/fcl/actions)

Simple app for cleanup files on you hosts.

## Setup

You can download releases [here](https://github.com/dlampsi/fcl/releases).

Actions example for darwin OS, version 1.0.0:
```bash
wget https://github.com/dlampsi/fcl/releases/download/1.0.0/fcl_1.0.0_darwin_amd64.zip
unzip fcl_1.0.0_darwin_amd64.zip
mv fcl_darwin_amd64 /usr/local/bin/fcl
chmod +x /usr/local/bin/fcl
```

## Usage

```bash
-check
    Run app in check mode. Only list files to delete
-mtime int
    Remove files by age (in days)
-no-colors
    Disable colors in app output
-path string
    Working dir path (default ".")
-size float
    Remove files by age (in MB)
-skip-dirs string
    A comma-separated list of dirs for deletion skip
-skip-files string
    A comma-separated list of full path to files for deletion skip
-v    Verbose output
```

## Examples

```bash
# Remove all files in current dir and subdirs older than 1 day
fcl -mtime 1

# Remove all files in specific folder dir and subdirs older than 1 day
fcl -path /dummy/path -mtime 1

# List (not delete) all files in current dir and subdirs older than 1 day
fcl -mtime 1 -check

# Delete files bigger than 10MB
fcl -path /dummy/path -size 10

# Files bigger than 25,5MB
fcl -path /dummy/path -size 25.5

# Delete files bigger than 10MB AND older than 5 days
fcl -path /dummy/path -size 10 -mtime 5
```