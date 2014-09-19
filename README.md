integrator
==========

Docker systems integrator

## Install

### From source

```bash
$ go get github.com/DimShadoWWW/integrator
$ cd $GOPATH/src/github.com/DimShadoWWW/integrator
$ make deps
$ make && scp bin/* core@CoreOS:/home/core/
```


## TODO

* Some testing would be good :)
* Add Container Start, Stop, run and Remove support
* Add container dependencies (working in integratorctl)
* Better support to cleanup container and images in web interface
