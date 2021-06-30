package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/// strip unused imports in rust

/// Sample warning text by cargo b

/**
warning: unused import: `common_exception::ErrorCode`
  --> common/aggregate_functions/src/aggregate_arg_min.rs:11:5
   |
11 | use common_exception::ErrorCode;
   |     ^^^^^^^^^^^^^^^^^^^^^^^^^^^
warning: unused import: `std::convert::TryFrom`
 --> common/aggregate_functions/src/aggregate_max.rs:6:5
  |
6 | use std::convert::TryFrom;
  |     ^^^^^^^^^^^^^^^^^^^^^

**/

var (
	mp    = make(map[string][]MsgLine)
	files = make(map[string]bool)

	importReg   = regexp.MustCompile("`(\\S+)`")
	rustFileReg = regexp.MustCompile(`[\w_\-/\d]+.rs`)

	pattern = ""
	root    = ""
)

type MsgLine struct {
	Line int
	Msg  string
}

func main() {
	flag.StringVar(&pattern, "pattern", "", "files glob patterns to process")
	flag.StringVar(&root, "root", "", "cargo root module")
	flag.Parse()

	if pattern != "" {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			log.Fatal(err)
		}

		for _, v := range matches {
			files[v] = true
		}
	}
	if err := buildMap(); err != nil {
		log.Fatal(err)
		return
	}

	if err := remove(); err != nil {
		log.Fatal(err)
		return
	}
}

func buildMap() error {
	cmd := exec.Command("cargo", "b")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	var sc = bufio.NewScanner(bytes.NewReader(bs))
	for sc.Scan() {
		var text = sc.Text()
		if strings.Contains(text, "warning: unused import:") {
			var matches = importReg.FindStringSubmatch(text)
			if len(matches) > 1 {
				importMsg := matches[1]
				sc.Scan()
				arrs := strings.Split(sc.Text(), ":")
				lineNum, err := strconv.Atoi(arrs[1])
				if err != nil {
					return err
				}
				file := rustFileReg.FindString(sc.Text())
				file = filepath.Join(root, file)

				var m = MsgLine{Line: lineNum, Msg: importMsg}
				if vs, ok := mp[file]; ok {
					vs = append(vs, m)
					mp[file] = vs
				} else {
					mp[file] = []MsgLine{m}
				}

				fmt.Printf("file: %s, msg --> %s, line --> %d \n", file, importMsg, lineNum)
			} else {
				log.Printf("Ignore %s \n", text)
			}

		}
	}
	return nil
}

func remove() error {
	for f, msgLines := range mp {
		if satisfies(f) {
			if err := removeLines(f, msgLines); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeLines(file string, msgLines []MsgLine) error {
	mod, err := os.Lstat(file)
	if err != nil {
		return err
	}
	mod.Mode()
	var newFileName = file + "_strip_new"

	newFile, err := os.OpenFile(newFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mod.Mode())
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(newFile)

	reader, err := os.Open(file)
	if err != nil {
		return err
	}

	lineIdx := 1
	msgIdx := 0

	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		if msgIdx >= len(msgLines) || lineIdx != msgLines[msgIdx].Line {
			bw.WriteString(sc.Text())
			bw.WriteByte('\n')
		} else if msgIdx < len(msgLines) {
			msgIdx += 1
		}
		lineIdx += 1
	}

	bw.Flush()
	if err := newFile.Close(); err != nil {
		return err
	}

	return os.Rename(newFileName, file)
}

func satisfies(f string) bool {
	if pattern == "" {
		return true
	}
	_, ok := files[f]
	return ok
}
