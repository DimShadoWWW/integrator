
all: build

build:
	go build
	rice append --exec integrator