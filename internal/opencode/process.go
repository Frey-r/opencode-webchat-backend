package opencode

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type Process struct {
	cmd    *exec.Cmd
	pty    *os.File
	stdin  io.Writer
	stdout *bufio.Reader

	done chan struct{}
}

func StartProcess(ctx context.Context, binaryPath string, args ...string) (*Process, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Env = os.Environ()

	ptyF, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &Process{
		cmd:    cmd,
		pty:    ptyF,
		stdin:  ptyF,
		stdout: bufio.NewReader(ptyF),
		done:   make(chan struct{}),
	}, nil
}

func (p *Process) Write(ctx context.Context, data string) error {
	_, err := p.stdin.Write([]byte(data + "\n"))
	return err
}

func (p *Process) ReadLine(ctx context.Context) (string, error) {
	line, err := p.stdout.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return line, nil
}

func (p *Process) Wait() error {
	err := p.cmd.Wait()
	close(p.done)
	return err
}

func (p *Process) Kill() error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *Process) IsDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}