package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestCleanKey(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{input: "CPU", output: "cpu"},
		{input: "Bzy_MHz", output: "bzy_mhz"},
		{input: "Busy%", output: "busy_percent"},
		{input: "CPU%c1", output: "cpu_percent_c1"},
		{input: "*Abc-DEF!", output: "abc_def"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.output, CleanKey(tt.input))
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input  string
		output bool
	}{
		{input: "0", output: true},
		{input: "123", output: true},
		{input: "123.45", output: true},
		{input: "", output: false},
		{input: "abc", output: false},
		{input: "x123", output: false},
		{input: "123x", output: false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.output, IsNumeric(tt.input))
	}
}

func TestIsTag(t *testing.T) {
	tests := []struct {
		input  string
		output bool
	}{
		{input: "0", output: true},
		{input: "12", output: true},
		{input: "-", output: true},
		{input: "abc", output: false},
		{input: "*", output: false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.output, IsTag(tt.input))
	}
}

func TestFormatRow(t *testing.T) {
	tests := []struct {
		keys   []string
		values []string
		output string
		err    error
	}{
		{
			keys:   []string{"cpu", "core", "busy_percent", "c1"},
			values: []string{"0", "1", "0.53", "634"},
			output: "turbostat,cpu=0,core=1 busy_percent=0.53,c1=634",
			err:    nil,
		},
		{
			keys:   []string{"cpu"},
			values: []string{"0", "1"},
			output: "",
			err:    errors.New("too many values"),
		},
		{
			keys:   []string{"cpu", "busy_percent"},
			values: []string{"?", "0.53"},
			output: "",
			err:    errors.New("invalid tag value: ?"),
		},
		{
			keys:   []string{"cpu", "busy_percent"},
			values: []string{"0", "bar"},
			output: "",
			err:    errors.New("invalid field value: bar"),
		},
	}
	for _, tt := range tests {
		output, err := FormatRow(tt.keys, tt.values)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.output, output)
	}
}

func TestProcessLines(t *testing.T) {
	tests := []struct {
		input  string
		output []string
		err    error
	}{
		{
			input: "testdata/in1",
			output: []string{
				"turbostat,core=-,cpu=- busy_percent=0.99,bzy_mhz=3104,corwatt=1.05,pkgwatt=24.93",
				"turbostat,core=0,cpu=0 busy_percent=0.89,bzy_mhz=2592,corwatt=0.06,pkgwatt=24.93",
				"turbostat,core=0,cpu=8 busy_percent=0.26,bzy_mhz=2639",
				"turbostat,core=-,cpu=- busy_percent=1.34,bzy_mhz=3272,corwatt=1.13,pkgwatt=24.90",
				"turbostat,core=0,cpu=0 busy_percent=1.30,bzy_mhz=2771,corwatt=0.09,pkgwatt=24.90",
				"turbostat,core=0,cpu=8 busy_percent=0.42,bzy_mhz=3744",
			},
			err: io.EOF,
		},
	}
	for _, tt := range tests {
		f, err := os.Open(tt.input)
		assert.NoError(t, err)
		defer f.Close()
		output := []string{}
		err = ProcessLines(f, func(s string) { output = append(output, s) })
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.output, output)
	}
}
