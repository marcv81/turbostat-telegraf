package turbostat

import (
	"github.com/influxdata/telegraf"

	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Converts an arbitrary string to a snake case key.
func cleanKey(s string) string {
	words := []string{}
	word := []rune{}
	flush := func() {
		if len(word) > 0 {
			words = append(words, string(word))
			word = []rune{}
		}
	}
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			c = c - ('A' - 'a')
			word = append(word, c)
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			word = append(word, c)
		} else {
			flush()
			if c == '%' {
				words = append(words, "percent")
			}
		}
	}
	flush()
	return strings.Join(words, "_")
}

// Returns whether a string represents a tag or not.
// Turbostat appears to only use integers and "-".
func isTag(s string) bool {
	if s == "-" {
		return true
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Converts keys and values to tags and fields.
// Keys and values are easy to read from Turbostat stdout.
// Tags and fields are easy to submit to Telegraf.
func toTagsAndFields(keys, values []string) (map[string]string, map[string]any, error) {
	if len(values) > len(keys) {
		msg := "too many values: keys=%s, values=%s"
		err := fmt.Errorf(msg, keys, values)
		return nil, nil, err
	}
	tags := map[string]string{}
	fields := map[string]any{}
	for i := range values {
		k := cleanKey(keys[i])
		if k == "cpu" || k == "core" || k == "apic" || k == "x2apic" {
			if !isTag(values[i]) {
				return nil, nil, fmt.Errorf("invalid tag: %s", values[i])
			}
			tags[k] = values[i]
		} else {
			v, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				return nil, nil, err
			}
			fields[k] = v
		}
	}
	if len(fields) == 0 {
		msg := "no value for any field: keys=%s, values=%s"
		return nil, nil, fmt.Errorf(msg, keys, values)
	}
	return tags, fields, nil
}

// Reads metrics from Turbostat stdout and adds them to an accumulator.
func processStdout(r io.Reader, a telegraf.Accumulator) error {
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return s.Err()
	}
	firstLine := s.Text()
	keys := strings.Split(firstLine, "\t")
	for s.Scan() {
		line := s.Text()
		if line == firstLine {
			continue
		}
		values := strings.Split(line, "\t")
		tags, fields, err := toTagsAndFields(keys, values)
		if err != nil {
			return err
		}
		a.AddFields("turbostat", fields, tags)
	}
	return s.Err()
}

// Reads error lines from Turbostat stderr and logs them.
func processStderr(r io.Reader, log telegraf.Logger) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		log.Info(line)
	}
	return s.Err()
}
