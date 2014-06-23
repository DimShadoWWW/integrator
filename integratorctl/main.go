package main

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/dnslib"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/alecthomas/kingpin"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// import . "github.com/DimShadoWWW/integrator/db"

var (
	etcdclient etcdlib.EtcdClient
	client     dockerlib.Lib

	// debug    = app.Flag("debug", "enable debug mode").Default("false").Bool()
	app         = kingpin.New("integratorctl", "A command-line application to generate and deploy multiple services in CoreOS.")
	serverIP    = app.Flag("server", "Etcd address").Default("127.0.0.1").IP()
	pretend     = app.Flag("pretend", "Instead of actually performing the deploy, simply generate what *would* have been deployed if pretend is false").Bool()
	add         = app.Command("add", "Generate and deploy a new service.")
	servicejson = add.Arg("jsonfile", "jsonfile for input").Required().File()
	serviceid   = add.Arg("id", "id for service").String()
	cleandns    = app.Command("cleardns", "Look for skydns2's etcd directories and remove the empty ones.")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Register user
	case "add":

		if *servicejson != nil {

			// DB.CreateTable(fleet.SystemdService{})
			// DB.CreateTable(fleet.Port{})
			// DB.CreateTable(fleet.Volume{})
			// DB.CreateTable(fleet.Link{})
			// DB.CreateTable(fleet.Env{})

			rand.Seed(time.Now().UnixNano())
			var id int64
			var err error

			if *serviceid != "" {
				id, err = strconv.ParseInt(*serviceid, 10, 64)
				if err != nil {
					fmt.Println("Fatal error ", err.Error())
					for i := 0; i < 10; i++ {
						id = rand.Int63() + 1
					}
				}
			} else {
				for i := 0; i < 10; i++ {
					id = rand.Int63() + 1
				}
			}

			myServices := fleet.SystemdServiceList{}

			err = myServices.FromJSON(*servicejson)
			if err != nil {
				panic(err)
			}

			for _, serv := range myServices.Services {
				fmt.Println(serv)
				serv.Id = id

				service_files := fleet.CreateSystemdFiles(serv, "./")
				// fmt.Println(fleet.CreateSystemdFiles(f, "./"))

				fmt.Println(*pretend)
				if *pretend == false {
					fmt.Println("DEPLOY")
					for _, s := range service_files {
						err = fleet.Deploy(s, "")
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
	case "cleandns":
		machines := []string{"http://" + string(*serverIP) + ":4001"}

		etcdclient, err := etcdlib.NewEtcdClient(machines)
		if err != nil {
			fmt.Println(err)
		}

		err = dnslib.CleanDNS(etcdclient)
		if err != nil {
			fmt.Println(err)
		}
	case "showdns":
		machines := []string{"http://" + string(*serverIP) + ":4001"}

		etcdclient, err := etcdlib.NewEtcdClient(machines)
		if err != nil {
			fmt.Println(err)
		}

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
}
