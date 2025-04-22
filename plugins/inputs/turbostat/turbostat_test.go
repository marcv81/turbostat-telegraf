package turbostat

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
	"time"
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
		assert.Equal(t, tt.output, cleanKey(tt.input))
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
		assert.Equal(t, tt.output, isTag(tt.input))
	}
}

func TestToTagsAndFields(t *testing.T) {
	tests := []struct {
		keys   []string
		values []string
		tags   map[string]string
		fields map[string]any
		err    string
	}{
		{
			keys:   []string{"CPU", "Core", "Busy%", "C1", "PkgWatt"},
			values: []string{"0", "1", "0.53", "634"},
			tags:   map[string]string{"cpu": "0", "core": "1"},
			fields: map[string]any{"busy_percent": 0.53, "c1": 634.0},
			err:    "",
		},
		{
			keys:   []string{"CPU", "Core"},
			values: []string{"0", "1", "0.53"},
			tags:   nil,
			fields: nil,
			err:    "too many values: keys=[CPU Core], values=[0 1 0.53]",
		},
		{
			keys:   []string{"CPU", "Core", "Busy%"},
			values: []string{"0", "1"},
			tags:   nil,
			fields: nil,
			err:    "no value for any field: keys=[CPU Core Busy%], values=[0 1]",
		},
		{
			keys:   []string{"CPU", "Busy%"},
			values: []string{"?", "0.53"},
			tags:   nil,
			fields: nil,
			err:    "invalid tag: ?",
		},
		{
			keys:   []string{"CPU", "Busy%"},
			values: []string{"0", "xyz"},
			tags:   nil,
			fields: nil,
			err:    "strconv.ParseFloat: parsing \"xyz\": invalid syntax",
		},
	}
	for _, tt := range tests {
		tags, fields, err := toTagsAndFields(tt.keys, tt.values)
		if tt.err == "" {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, tt.err, err.Error())
		}
		assert.Equal(t, tt.tags, tags)
		assert.Equal(t, tt.fields, fields)
	}
}

func TestProcessStdout(t *testing.T) {
	tests := []struct {
		input  string
		output []telegraf.Metric
	}{
		{
			input: "testdata/in",
			output: []telegraf.Metric{
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "-", "cpu": "-"},
					map[string]any{"busy_percent": 0.99, "corwatt": 1.05},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "0", "cpu": "0"},
					map[string]any{"busy_percent": 0.89, "corwatt": 0.06},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "0", "cpu": "8"},
					map[string]any{"busy_percent": 0.26},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "-", "cpu": "-"},
					map[string]any{"busy_percent": 1.34, "corwatt": 1.13},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "0", "cpu": "0"},
					map[string]any{"busy_percent": 1.30, "corwatt": 0.09},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"turbostat",
					map[string]string{"core": "0", "cpu": "8"},
					map[string]any{"busy_percent": 0.42},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		f, err := os.Open(tt.input)
		assert.NoError(t, err)
		defer f.Close()
		a := &testutil.Accumulator{}
		assert.NoError(t, processStdout(f, a))
		testutil.RequireMetricsEqual(t, tt.output, a.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}
