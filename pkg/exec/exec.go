package exec

import (
	"fmt"
	"os"
	"os/exec"
)

type Option func(*exec.Cmd)

func WithEnv(env ...string) Option {
	return func(cmd *exec.Cmd) {
		cmd.Env = append(cmd.Env, env...)
	}
}

func Run(arg []string, opts ...Option) error {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Env = os.Environ()
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stdout
	for _, opt := range opts {
		opt(cmd)
	}

	if buff, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run: '%v' cmd\nerror: %v\noutput: %v\n", arg, err, string(buff))
	}
	return nil
}

func runCmd(arg []string) error {
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
