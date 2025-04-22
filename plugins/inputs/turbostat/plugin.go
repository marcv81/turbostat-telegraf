package turbostat

import (
	"github.com/influxdata/telegraf"

	"context"
	"os/exec"
	"sync"
)

type Turbostat struct {
	Cmd  string
	Args []string

	Log telegraf.Logger

	cancel context.CancelFunc
}

func (t *Turbostat) Gather(a telegraf.Accumulator) error {
	return nil
}

func (t *Turbostat) SampleConfig() string {
	return ""
}

// Starts a child process and goroutines to process stdout/stderr.
func (t *Turbostat) Start(a telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	cmd := exec.CommandContext(ctx, t.Cmd, t.Args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err = processStdout(stdout, a)
		if err != nil {
			t.Log.Errorf("error processing stdout: %s", err)
			cancel()
		}
	}()
	go func() {
		defer wg.Done()
		err = processStderr(stderr, t.Log)
		if err != nil {
			t.Log.Errorf("error processing stderr: %s", err)
			cancel()
		}
	}()
	go func() {
		wg.Wait()
		err := cmd.Wait()
		if err != nil {
			t.Log.Errorf("child process stopped: %s", err)
		}
		t.Log.Error("metrics emission stopped")
	}()
	return nil
}

// Stops the child process.
func (t *Turbostat) Stop() {
	t.cancel()
}
