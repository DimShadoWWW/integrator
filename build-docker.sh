#!/bin/bash

docker run --rm -v "$(pwd)":/go/src/github.com/DimShadoWWW/integrator/ -w /go/src/github.com/DimShadoWWW/integrator golang:latest make
