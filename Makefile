
have_apt := $(wildcard /usr/bin/apt-get)
have_zip := $(shell which zip)

all: deps build

deps:
	go get -u github.com/GeertJohan/go.rice
	go get -u github.com/GeertJohan/go.incremental
	go get -u github.com/akavel/rsrc
	go get -u github.com/jessevdk/go-flags
	go install github.com/GeertJohan/go.rice/rice
	go get -u github.com/gorilla/mux
	go get -u github.com/rakyll/gometry
	go get -u github.com/wsxiaoys/terminal/color
	go get -u github.com/deckarep/golang-set
	go get -u github.com/dotcloud/docker/engine
	go get -u github.com/dotcloud/docker/nat
	go get -u github.com/dotcloud/docker/utils
	go get -u github.com/fsouza/go-dockerclient
	go get -u github.com/stevedomin/termtable
	go get -u github.com/coreos/locksmith/lock
	go get -u github.com/coreos/go-etcd/etcd
	go get -u github.com/cihub/seelog
	go get -u github.com/alecthomas/kingpin
	go get -u gopkg.in/redis.v2

build: builddnsctl buildproxyctl buildintegratorctl buildintegrator

buildintegrator:
	go build
	rice append --exec integrator
	mv integrator bin

builddnsctl:
	cd dnsctl && go build && mv dnsctl ../bin

buildproxyctl:
	cd proxyctl && go build && mv proxyctl ../bin

buildintegratorctl:
	cd integratorctl && go build && mv integratorctl ../bin

