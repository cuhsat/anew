package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/edsrzf/mmap-go"
	"github.com/zeebo/xxh3"
)

func main() {
	var quietMode bool
	var dryRun bool
	var trim bool

	flag.BoolVar(&quietMode, "q", false, "quiet mode (no output at all)")
	flag.BoolVar(&dryRun, "d", false, "don't append anything to the file, just print the new lines to stdout")
	flag.BoolVar(&trim, "t", false, "trim leading and trailing whitespace before comparison")

	flag.Parse()

	fn := flag.Arg(0)

	lines := make(map[uint64]struct{})

	var f io.WriteCloser

	if fn != "" {
		// read the hashed file into a map
		f, err := os.Open(fn)

		if err != nil {
			log.Fatalln(err)
		}

		m, err := mmap.Map(f, mmap.RDONLY, 0)

		if err != nil {
			log.Fatalln(err)
		}

		sc := bufio.NewScanner(bytes.NewReader(m))

		for sc.Scan() {
			line := sc.Text()
			if trim {
				line = strings.TrimSpace(line)
			}
			lines[xxh3.HashString(line)] = struct{}{}
		}

		_ = m.Unmap()
		_ = f.Close()

		if !dryRun {
			// re-open the file for appending new stuff
			f, err = os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to open file for writing: %s\n", err)
				return
			}

			defer func() {
				_ = f.Close()
			}()
		}
	}

	// read the lines, append and output them if they're new
	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		line := sc.Text()

		if trim {
			line = strings.TrimSpace(line)
		}

		if _, ok := lines[xxh3.HashString(line)]; ok {
			continue
		}

		// add the line to the map so we don't get any duplicates from stdin
		lines[xxh3.HashString(line)] = struct{}{}

		if !quietMode {
			fmt.Println(line)
		}

		if !dryRun {
			if fn != "" {
				_, _ = fmt.Fprintf(f, "%s\n", line)
			}
		}
	}
}
