// Copyright 2024 Matthew P. Dargan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Yesterday prints file names from the dump.
//
// Usage:
//
//	yesterday [-c | -C | -d] [-n daysago | -t [[yy]yy]mm]dd] file...
//
// Yesterday prints the names of the files from the most recent dump. Since
// dumps are done early in the morning, yesterday's files are really in today's
// dump. For example, if today is February 11, 2003,
//
//	$ yesterday /home/am3/rsc/.profile
//
// prints
//
//	/dump/am/2003/0211/home/am3/rsc/.profile
//
// In fact, the implementation is to select the most recent dump in the current
// year, so the dump selected may not be from today. Yesterday does not
// guarantee that the string it prints represents an existing file.
//
// By default, yesterday prints the names of the dump files corresponding to
// the named files. The first set of options changes this behavior.
//
// The -c flag causes yesterday to copy the dump files over the named files.
//
// The -C flag causes yesterday to copy the dump files over the named files
// only when they differ.
//
// The -d flag causes yesterday to run “diff” to compare the dump files
// with the named files.
//
// The -n flag causes yesterday to select the dump daysago prior to the current
// day.
//
// The -t flag causes yesterday to select other day’s dumps, with a format of
// 1, 2, 4, 6, or 8 digits of the form d, dd, mmdd, yymmdd, or yyyymmdd.
//
// Examples:
//
// See what’s changed in the last week in your profile:
//
//	$ yesterday −d −n 7 ~/.profile
//	diff -c /dump/am/2024/0211/home/mpd/.profile /home/mpd/.profile
//
// Restore your profile from yesterday:
//
//	$ yesterday −c ~/.profile
//	cp /dump/am/2024/0217/home/mpd/.profile /home/mpd/.profile
package main

import (
	"bytes"
	"crypto/sha512"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	cp       = flag.Bool("c", false, "copy dump files over named files")
	cpIfDiff = flag.Bool("C", false, "copy dump files over named files if they differ")
	diff     = flag.Bool("d", false, "compare dump files with named files")
	daysAgo  = flag.Uint("n", 0, "selects dump days prior to the current day")
	date     = flag.String("t", "", "selects other day's dumps")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: yesterday [-c | -C | -d] [-n daysago | -t [[yy]yy]mm]dd] file...\n")
	os.Exit(2)
}

func main() {
	log.SetPrefix("yesterday: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) < 1 || (*cp && (*cpIfDiff || *diff)) || (*daysAgo > 0 && *date != "") {
		usage()
	}
	t := time.Now()
	if *daysAgo > 0 {
		t = t.AddDate(0, 0, -int(*daysAgo))
	} else if *date != "" {
		var err error
		t, err = parseDate(t, *date)
		if err != nil {
			log.Fatal(err)
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	dump := filepath.Join("/dump", hostname)
	if _, err := os.Stat(dump); os.IsNotExist(err) {
		log.Fatal(err)
	}
	for _, f := range flag.Args() {
		if !filepath.IsAbs(f) {
			f = filepath.Join(dir, f)
		}
		dp, err := datePath(dump, t)
		if err != nil {
			log.Fatal(err)
		}
		if err = processFile(filepath.Join(dp, f), f); err != nil {
			log.Fatal(err)
		}
	}
}

const layout = "20060102"

func parseDate(t time.Time, d string) (time.Time, error) {
	refDate := t.Format(layout)
	switch len(d) {
	case 1:
		refDate = refDate[:len(refDate)-2] + "0" + d
	case 2, 4, 6, 8:
		refDate = refDate[:len(refDate)-len(d)] + d
	default:
		return time.Time{}, fmt.Errorf("invalid date: %s", d)
	}
	return time.Parse(layout, refDate)
}

func datePath(dump string, t time.Time) (string, error) {
	y := fmt.Sprint(t.Year())
	dump = filepath.Join(dump, y)
	if *daysAgo > 0 || *date != "" {
		d := fmt.Sprintf("%02d%02d", t.Month(), t.Day())
		return filepath.Join(dump, d), nil
	}
	ents, err := os.ReadDir(dump)
	if err != nil {
		return "", err
	}
	var recentDir os.DirEntry
	var recentModTime time.Time
	for _, e := range ents {
		if e.IsDir() {
			info, err := e.Info()
			if err != nil {
				continue
			}
			modTime := info.ModTime()
			if modTime.After(recentModTime) {
				recentDir = e
				recentModTime = modTime
			}
		}
	}
	if recentDir != nil {
		return filepath.Join(dump, recentDir.Name()), nil
	}
	return "", fmt.Errorf("no directory entries in %s", dump)
}

func processFile(dump, f string) error {
	switch {
	case *cp:
		return cpFile(dump, f)
	case *cpIfDiff:
		return cpIfDifferent(dump, f)
	case *diff:
		diffFiles(dump, f)
	default:
		fmt.Println(dump)
	}
	return nil
}

func cpFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	fmt.Printf("cp %s %s\n", src, dst)
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func cpIfDifferent(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	hSrc := sha512.New()
	if _, err := io.Copy(hSrc, srcFile); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	hDst := sha512.New()
	if _, err := io.Copy(hDst, dstFile); err != nil {
		return err
	}
	if sSrc, sDst := hSrc.Sum(nil), hDst.Sum(nil); bytes.Equal(sSrc, sDst) {
		return nil
	}
	if _, err = srcFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err = dstFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	fmt.Printf("cp %s %s\n", src, dst)
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func diffFiles(f1, f2 string) {
	cmd := exec.Command("diff", "-c", f1, f2)
	fmt.Println(cmd)
	data, _ := cmd.CombinedOutput()
	fmt.Print(string(data))
}
