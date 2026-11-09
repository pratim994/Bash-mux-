
package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)


func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState )

	cmd := exec.Command("bash")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	rows, cols, err := pty.Getsize(os.Stdin)

	if err != nil {
		rows, cols = 24, 80
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})

	if err != nil {
		return err
	}

	defer ptmx.Close()

	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	defer signal.Stop(resize)

	go func() {
		for range resize {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()

	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()


	_, copyErr := io.Copy(os.Stdout, ptmx)

	_ = cmd.Wait()

	return copyErr
}

