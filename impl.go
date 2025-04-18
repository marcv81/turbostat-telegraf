package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Turns an arbitrary string into a snake case key.
func CleanKey(s string) string {
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

// Returns whether a string represents a number or not.
func IsNumeric(s string) bool {
	var err error
	_, err = strconv.Atoi(s)
	if err == nil {
		return true
	}
	_, err = strconv.ParseFloat(s, 64)
	if err == nil {
		return true
	}
	return false
}

// Returns whether a string represents a tag or not.
// Turbostat appears to only use integers and "-".
func IsTag(s string) bool {
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

// Formats a data row in InfluxDB format.
// The keys should go through CleanKey() before calling FormatRow().
func FormatRow(keys, values []string) (string, error) {
	if len(values) > len(keys) {
		return "", fmt.Errorf("too many values")
	}
	tags := []string{}
	fields := []string{}
	for i, _ := range values {
		k := keys[i]
		if k == "cpu" || k == "core" || k == "apic" || k == "x2apic" {
			if !IsTag(values[i]) {
				return "", fmt.Errorf("invalid tag value: %s", values[i])
			}
			tags = append(tags, fmt.Sprintf("%s=%s", k, values[i]))
		} else {
			if !IsNumeric(values[i]) {
				return "", fmt.Errorf("invalid field value: %s", values[i])
			}
			fields = append(fields, fmt.Sprintf("%s=%s", k, values[i]))
		}
	}
	return "turbostat," + strings.Join(tags, ",") + " " + strings.Join(fields, ","), nil
}

// Reads a single line from a bufio.Reader.
func ReadLine(br *bufio.Reader) (string, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = line[:len(line)-1]
	return line, nil
}

// Transforms Turbostat output lines into InfluxDB data rows.
// Calls a function on each of the data rows.
func ProcessLines(r io.Reader, process func(s string)) error {
	br := bufio.NewReader(r)
	firstLine, err := ReadLine(br)
	if err != nil {
		return err
	}
	keys := []string{}
	for _, k := range strings.Split(firstLine, "\t") {
		keys = append(keys, CleanKey(k))
	}
	for {
		line, err := ReadLine(br)
		if err != nil {
			return err
		}
		if line == firstLine {
			continue
		}
		values := strings.Split(line, "\t")
		s, err := FormatRow(keys, values)
		if err != nil {
			return err
		}
		process(s)
	}
}
