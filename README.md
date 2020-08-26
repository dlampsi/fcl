# fcl

[![Actions Status](https://github.com/dlampsi/fcl/workflows/default/badge.svg)](https://github.com/dlampsi/fcl/actions)

File cleaner (fcl) is app for cleanup files from host folder (and subfolders) by modification time and/or file size.

## Setup

You can download releases [here](https://github.com/dlampsi/fcl/releases).

Install commands example for version `0.0.1` on linux OS:
```bash
wget https://github.com/dlampsi/fcl/releases/download/0.0.1/fcl_0.0.1_linux_amd64.zip
unzip fcl_0.0.1_linux_amd64.zip
mv fcl_0.0.1_linux_amd64 /usr/local/bin/fcl
chmod +x /usr/local/bin/fcl
```

## Usage

All usage commands available on help flag:

```bash
fcl -h
```

Some command examples:

```bash
# Remove all files in CURRENT dir and subdirs older than 1 day
fcl -mtime 1

# List (not delete) all files in current dir and subdirs older than 1 day
fcl -mtime 1 -check

# Remove all files in specific folder dir and subdirs older than 1 day
fcl -path /dummy/path -mtime 1

# Remove all files in specific folder dir and subdirs older than 1 day with exception dir and file
fcl -path /dummy/path -mtime 1 -skip /dummy/path/subdir1,/dummy/path/file.log

# Delete files bigger than 10MB
fcl -path /dummy/path -size 10

# Files bigger than 25,5MB
fcl -path /dummy/path -size 25.5

# Delete files bigger than 10MB AND older than 5 days
fcl -path /dummy/path -size 10 -mtime 5
```