package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func main() {
	cmd := exec.Command("turbostat", os.Args[1:]...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		Fatal(err)
	}
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		Fatal(err)
	}
	go func() {
		err = ProcessLines(pipe, func(s string) { fmt.Println(s) })
		if err != io.EOF {
			Fatal(err)
		}
	}()
	err = cmd.Wait()
	if err != nil {
		Fatal(err)
	}
}
