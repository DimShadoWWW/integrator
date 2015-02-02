package fleet

import (
	"github.com/DimShadoWWW/fleet-client-go"
	"github.com/wsxiaoys/terminal/color"
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
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return err
	}

	return nil
}
