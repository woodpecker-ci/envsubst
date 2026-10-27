package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/woodpecker-ci/envsubst"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	stdin := bufio.NewScanner(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)

	for stdin.Scan() {
		line, err := envsubst.EvalEnv(stdin.Text())
		if err != nil {
			return fmt.Errorf("error while envsubst: %w", err)
		}
		_, err = fmt.Fprintln(stdout, line)
		if err != nil {
			return fmt.Errorf("error while writing to stdout: %w", err)
		}
		stdout.Flush()
	}

	return nil
}
