package main

import (
	"flag"
	"fmt"
	"github.com/DimShadoWWW/integrator/dnslib"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"math/rand"
	"strconv"
	"time"
	// "encoding/json"
	// "github.com/DimShadoWWW/integrator/coreoslib"
	// "io"
	// "os"
	// "path/filepath"
	// "strings"
	// "github.com/coreos/go-etcd/etcd"
)

import . "github.com/DimShadoWWW/integrator/db"

// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

var (
	etcdclient etcdlib.EtcdClient
	client     dockerlib.Lib
)

func main() {

	flag.Parse()

	// client = dockerlib.NewDockerLib("unix:///var/run/docker.sock")

	// err := client.GetContainerHostPort("postgresql-3599648947757342585", 5432)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	var deploy_all bool
	deploy_all = false

	DB.CreateTable(fleet.SystemdService{})
	DB.CreateTable(fleet.Port{})
	DB.CreateTable(fleet.Volume{})
	DB.CreateTable(fleet.Link{})
	DB.CreateTable(fleet.Env{})

	machines := []string{"http://127.0.0.1:4001"}
	rand.Seed(time.Now().UnixNano())
	var id int64
	var err error

	if len(flag.Args()) > 0 {
		myid, err := strconv.ParseInt(flag.Args()[0], 10, 64)
		if err != nil {
			fmt.Println("Fatal error ", err.Error())
			for i := 0; i < 10; i++ {
				id = rand.Int63() + 1
			}
		}

		id = myid

	} else {
		for i := 0; i < 10; i++ {
			id = rand.Int63() + 1
		}
	}

	fmt.Println(deploy_all)
	fmt.Println(id)
	fmt.Println(machines)

	var myServices []fleet.SystemdService

	// var myports []fleet.Port
	// var myvolumes []fleet.Volume
	// var mylinks []fleet.Link

	// v1 := fleet.Volume{Loca1lDir: "/local/dir", ContainerDir: "/container/dir"}
	// myvolumes = append(myvolumes, v1)

	// // Gogeta
	// var ggports []fleet.Port

	// p := fleet.Port{
	// 	HostPort:      "80",
	// 	ContainerPort: "7777",
	// }

	// ggports = append(ggports, p)

	// f := fleet.SystemdService{
	// 	Id:          id,
	// 	Name:        "gogeta",
	// 	Description: "Proxy gogeta",
	// 	Command:     "gogeta -port=80 -etcdAddress=\"http://172.17.42.1:4001\" -domainDir=\"/domains\" -envDir=\"/services\" -templateDir=\"/gogeta/templates\"",
	// 	ImageName:   "shadow/gogeta",
	// 	Hostname:    "gogeta",
	// 	Ports:       ggports,
	// 	Domain:      "local",
	// 	Region:      "east",
	// 	// Variables:   []string,
	// 	// IncludeFleet bool
	// }
	// myServices = append(myServices, f)

	// Vulcand
	var vuports []fleet.Port
	var vuvolumes []fleet.Volume
	p := fleet.Port{
		HostPort:      "80",
		ContainerPort: "8181",
	}

	vuports = append(vuports, p)

	p = fleet.Port{
		HostPort:      "8182",
		ContainerPort: "8182",
	}

	vuports = append(vuports, p)

	f := fleet.SystemdService{
		Id:          id,
		Name:        "vulcand",
		Description: "Proxy vulcand",
		Command:     "/opt/vulcan/vulcand -apiInterface=\"0.0.0.0\" -interface=\"0.0.0.0\" --etcd=http://172.17.42.1:4001",
		ImageName:   "shadow/vulcand",
		Hostname:    "vulcand",
		Ports:       vuports,
		Volumes:     vuvolumes,
		Domain:      "local",
		Region:      "east",
		Conflicts:   []string{"vulcand*"},
		// Variables:   []string,
		// IncludeFleet bool
	}
	myServices = append(myServices, f)

	// Postgresql
	var pgports []fleet.Port
	var pgvolumes []fleet.Volume

	p = fleet.Port{
		HostPort:      "5432",
		ContainerPort: "5432",
		Protocol:      "tcp",
	}
	pgports = append(pgports, p)

	v1 := fleet.Volume{LocalDir: "/home/core/datos/db/", ContainerDir: "/var/lib/pgsql/9.3/"}
	pgvolumes = append(pgvolumes, v1)

	f = fleet.SystemdService{
		Id:          id,
		Name:        "postgresql-" + strconv.FormatInt(id, 10),
		Description: "Postgresql",
		ImageName:   "shadow/centos-postgresql",
		Hostname:    "postgresql",
		Ports:       pgports,
		Volumes:     pgvolumes,
		Domain:      "local",
		Region:      "east",
		// Variables:   []string,
		// IncludeFleet bool
	}
	myServices = append(myServices, f)

	// Deliver
	var deports []fleet.Port
	var devolumes []fleet.Volume
	var delinks []fleet.Link

	l := fleet.Link{
		Name:  "postgresql-" + strconv.FormatInt(id, 10),
		Alias: "postgresql",
	}
	delinks = append(delinks, l)

	p = fleet.Port{
		HostPort:      "1026",
		ContainerPort: "22",
		Protocol:      "tcp",
	}
	deports = append(deports, p)

	p = fleet.Port{
		ContainerPort: "80",
	}
	deports = append(deports, p)

	v1 = fleet.Volume{LocalDir: "/home/core/code/deliver/", ContainerDir: "/app/"}
	devolumes = append(devolumes, v1)

	f = fleet.SystemdService{
		Id:          id,
		Name:        "deliver-" + strconv.FormatInt(id, 10),
		Description: "deliver",
		ImageName:   "shadow/centos-php-deliver",
		Hostname:    "deliver",
		Ports:       deports,
		Volumes:     devolumes,
		Links:       delinks,
		Domain:      "local",
		Region:      "east",
		Deps:        []string{"postgresql-" + strconv.FormatInt(id, 10)},
		HttpPort:    80,
		// Variables:   []string,
		// IncludeFleet bool
	}
	myServices = append(myServices, f)

	deploy_all = true

	// // logspout
	// v1 := fleet.Volume{LocalDir: "/var/run/docker.sock", ContainerDir: "/tmp/docker.sock"}
	// myvolumes = append(myvolumes, v1)

	// p := fleet.Port{
	// 	HostPort:      "8000",
	// 	ContainerPort: "8000",
	// }

	// myports = append(myports, p)

	// f := fleet.SystemdService{
	// 	Id:          id,
	// 	Name:        "logspout",
	// 	Description: "Logs",
	// 	ImageName:   "progrium/logspout",
	// 	Hostname:    "logspout",
	// 	Ports:       myports,
	// 	Volumes:     myvolumes,
	// 	Domain:      "local",
	// 	HttpPort:    8000,
	// 	Region:      "east",
	// }

	// var f1 fleet.SystemdService
	// DB.Where(&f).First(&f1)
	// DB.Save(&f)

	// f := fleet.SystemdService{
	// 	Id:          strconv.Itoa(id),
	// 	Name:        "a",
	// 	Description: "Servicio A ",
	// 	// Ports:       {"8080"},
	// }

	// f := fleet.SystemdService{}

	// err := f.FromJSON("postgresql.json")
	// if err != nil {
	// 	panic(err)
	// }

	// f.Id = id

	// if err = json.Unmarshal(byt, &dat); err != nil {
	// 	panic(err)
	// }

	etcdclient, err := etcdlib.NewEtcdClient(machines)
	if err != nil {
		fmt.Println(err)
	}

	// err = dnslib.CleanDNS(etcdclient)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	for _, serv := range myServices {
		fmt.Println(serv)

		err := serv.ToJSON(serv.Name + ".json")
		if err != nil {
			panic(err)
		}

		service_files := fleet.CreateSystemdFiles(serv, "./")
		// fmt.Println(fleet.CreateSystemdFiles(f, "./"))

		if deploy_all == true {
			fmt.Println("DEPLOY")
			for _, s := range service_files {
				err = fleet.Deploy(s, "")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	// err = dnslib.DeleteHostnameDNS(etcdclient, "f20e.east.deliver.local", "172.17.0.5", 80)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	hostnames, _ := dnslib.GetHostnamesDNS(etcdclient, "/skydns/")

	for _, hostname := range hostnames {
		fmt.Println(hostname.Hostname)
		fmt.Println(hostname.EtcdKey)
		for _, entry := range hostname.Entry {
			fmt.Println(entry.Host)
			fmt.Println(entry.Port)
			fmt.Println(entry.Priority)
		}
	}

	// coreoslib.CoreOsLocksmith(machines, 1)

}
