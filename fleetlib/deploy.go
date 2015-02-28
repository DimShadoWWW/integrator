package fleet

import (
	"github.com/DimShadoWWW/fleet-client-go"
	"github.com/wsxiaoys/terminal/color"
	"log"
	"os"
	"path/filepath"
)

func Deploy(fname string, address string) error {
	_, err := os.Stat(fname)
	if err != nil {
		return err
	}

	fName := filepath.Base(fname)

	fleetClient := client.NewClientAPI()
	log.Println("Loading: ", fname)
	err = fleetClient.Submit(fName, fname)

	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Println(err.Error())
		return err
	}

	return nil
}
