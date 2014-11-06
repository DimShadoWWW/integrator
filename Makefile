
have_apt := $(wildcard /usr/bin/apt-get)
have_zip := $(shell which zip)

all: deps build

deps:
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.incremental
	go get github.com/akavel/rsrc
	go get github.com/jessevdk/go-flags
	go install github.com/GeertJohan/go.rice/rice
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
	go get github.com/cihub/seelog
	go get github.com/alecthomas/kingpin
	go get gopkg.in/redis.v2

build: builddnsctl buildproxyctl buildintegratorctl buildintegrator

buildintegrator:
	go build
	/go/bin/rice append --exec integrator
	mv integrator bin

builddnsctl:
	cd dnsctl && go build && mv dnsctl ../bin

buildproxyctl:
	cd proxyctl && go build && mv proxyctl ../bin

buildintegratorctl:
	cd integratorctl && go build && mv integratorctl ../bin

