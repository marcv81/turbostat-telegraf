package main

import (
	"github.com/influxdata/telegraf/plugins/common/shim"
	"github.com/marcv81/turbostat-telegraf-plugin/plugins/inputs/turbostat"

	"fmt"
	"os"
)

func fatalf(s string, a ...any) {
	fmt.Fprintf(os.Stderr, s, a...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fatalf("Usage: turbostat-telegraf-plugin [command] <args>\n")
	}
	s := shim.New()
	err := s.AddInput(&turbostat.Turbostat{
		Cmd:  os.Args[1],
		Args: os.Args[2:],
	})
	if err != nil {
		fatalf("Error adding plugin to shim: %s\n", err)
	}
	err = s.Run(shim.PollIntervalDisabled)
	if err != nil {
		fatalf("Error running plugin: %s\n", err)
	}
}
