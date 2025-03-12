package client

import (
	"flag"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(0)
	}

	cmd := exec.Command("docker", "build", "-t", "soultest:latest", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	go func() {
		cmd := exec.Command("docker", "run", "--name", "soultest", "-p", "2242:2242", "soultest:latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	m.Run()

	cmd = exec.Command("docker", "stop", "soultest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("docker", "rm", "soultest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
