package main

import (
	"encoding/json"
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerio"
	"github.com/vmihailenco/redis/v2"
	"io/ioutil"
	"os"
)

type SiteConfig struct {
	host   string
	values []string
}

var sites []SiteConfig

func main() {
	fmt.Println("Starting Integrator")
}
