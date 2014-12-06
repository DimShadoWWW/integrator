package integratorlib

import (
	"github.com/DimShadoWWW/integrator/dockerlib"
)

type ContainerStatus struct {
	Containers []dockerlib.APIContainers
	Status     map[string]int
}

type ImagesStatus struct {
	Images []dockerlib.APIImages
	Status map[string]int
}

type Status struct {
	Containers ContainerStatus
	Images     ImagesStatus
}

type IntegratorStruct struct {
	APIKey  string
	Client  dockerlib.Lib
	Basedir string
}

func NewIntegrator(address, apikey, basedir string) (IntegratorStruct, error) {

	client, err := dockerlib.NewDockerLib(address)
	if err != nil {
		return IntegratorStruct{}, err
	}

	return IntegratorStruct{
		Client:  client,
		APIKey:  apikey,
		Basedir: basedir,
	}, nil
}
