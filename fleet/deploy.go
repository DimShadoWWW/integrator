package fleet

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Deploy(fname string, address string) error {
	_, err := os.Stat(fname)
	if err != nil {
		return err
	}

	fName := filepath.Base(fname)

	fmt.Println("/usr/bin/fleetctl", "start", fName)
	cmd := exec.Command("/usr/bin/fleetctl", "start", fName)

	if address != "" {
		fmt.Println("/usr/bin/fleetctl", "-endpoint=\""+address+"\"", "start", fName)
		cmd = exec.Command("/usr/bin/fleetctl", "-endpoint=\""+address+"\"", "start", fName)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		fmt.Errorf("error: %s\n", err)
		fmt.Printf("stderr: %s\n", stderr.String())
		fmt.Printf("stdout: %s\n", stdout.String())
		return err
	}

	return nil
}
