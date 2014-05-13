
all: build

deps:
	go get github.com/GeertJohan/go.rice
	go get github.com/gorilla/mux
	go get github.com/wsxiaoys/terminal/color
	go get github.com/deckarep/golang-set
	go get github.com/dotcloud/docker/engine
	go get github.com/dotcloud/docker/nat
	go get github.com/dotcloud/docker/utils
	go get github.com/fsouza/go-dockerclient
	go get github.com/stevedomin/termtable
	go get github.com/coreos/locksmith/lock
	go get github.com/coreos/go-etcd/etcd

build:
	go build
	rice append --exec integrator
