package fleet

import (
	"fmt"
	"github.com/DimShadoWWW/fleet-client-go"
	"os"
	"path/filepath"
)

func Deploy(fname string, address string) error {
	_, err := os.Stat(fname)
	if err != nil {
		return err
	}

	fName := filepath.Base(fname)
	ext := filepath.Ext(fName)
	name := fName[0 : len(fName)-len(ext)]

	fleetClient := client.NewClientAPI()
	err = fleetClient.Submit(name, fname)

	if err != nil {
		fmt.Errorf("error: %s\n", err)
		return err
	}

	return nil
}
