package hipache

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"gopkg.in/redis.v2"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

var (
	ipaddress        string
	toclean_etcdpath []string
	container_name   string
	external_domain  = "production"
)

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func HostAdd(redisAccess, dockeruri string, service fleet.SystemdService, port int, path string) error {

	// Docker internal ip address
	fmt.Println("Adding internal PROXY entry")
	// if service.Id == 0 {
	container_name = service.Name
	// } else {
	// 	container_name = service.Name + "-" + strconv.FormatInt(service.Id, 10)
	// }

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain

	rand.Seed(time.Now().UnixNano())

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	retries := 31
	for i := 0; i < retries; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			// retry until this
			if i == retries-1 {
				fmt.Println("Failed to get docker's container ip address")
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	re := regexp.MustCompile(".*//(?P<user>[^:@]*)?:?(?P<pass>[^:@]*)?@?(?P<hostname>[A-Za-z0-9._-]+)?:?(?P<port>[0-9]*)?/(?P<db>[0-9])?")

	res := re.FindStringSubmatch(redisAccess)
	// redisuser := res[1]
	fmt.Printf("%+v\n", res)
	redispasswd := res[2]
	redishost := res[3]
	if redishost == "" {
		redishost = "127.0.0.1"
	}

	redisport := res[4]
	if redisport == "" {
		redisport = "6379"
	}
	redisdbname := res[5]
	if redisdbname == "" {
		redisdbname = "0"
	}
	dbname, err := strconv.ParseInt(redisdbname, 10, 64)
	if err != nil {
		return err
	}

	client := redis.NewTCPClient(&redis.Options{
		Addr:     redishost + ":" + redisport,
		Password: redispasswd, // no password set
		DB:       dbname,      // use default DB
	})

	// try to find it first
	if client.LLen("frontend:"+fullname).Val() > 0 {
		// try to find it first
		vlist := client.LRange("frontend:"+fullname, 0, -1)
		if vlist.Err() != nil {
			return vlist.Err()
		}
		for _, v := range vlist.Val() {
			if v == "http://"+ipaddress+":"+strconv.Itoa(port) {
				// it's already there
				return nil
			}
		}
	} else {
		// if empty, add hostname as first value
		client.RPush("frontend:"+fullname, service.Hostname)
	}

	client.RPush("frontend:"+fullname, "http://"+ipaddress+":"+strconv.Itoa(port))

	return nil
}

func HostDel(redisAccess, dockeruri string, service fleet.SystemdService, port int, path string) error {

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain

	// Docker internal ip address
	fmt.Println("Remove internal PROXY entry")
	container_name = service.Name

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	retries := 10

	for i := 0; i < retries; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			if i == retries-1 {
				fmt.Println("Failed to get docker's container ip address")
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	re := regexp.MustCompile(".*//(?P<user>[^:@]*)?:?(?P<pass>[^:@]*)?@?(?P<hostname>[A-Za-z0-9._-]+)?:?(?P<port>[0-9]*)?/(?P<db>[0-9])?")

	res := re.FindStringSubmatch(redisAccess)
	// redisuser := res[1]
	redispasswd := res[2]
	redishost := res[3]
	if redishost == "" {
		redishost = "127.0.0.1"
	}

	redisport := res[4]
	if redisport == "" {
		redisport = "6379"
	}
	redisdbname := res[5]
	if redisdbname == "" {
		redisdbname = "0"
	}
	dbname, err := strconv.ParseInt(redisdbname, 10, 64)
	if err != nil {
		return err
	}

	client := redis.NewTCPClient(&redis.Options{
		Addr:     redishost + ":" + redisport,
		Password: redispasswd, // no password set
		DB:       dbname,      // use default DB
	})

	client.LRem("frontend:"+fullname, 0, "http://"+ipaddress+":"+strconv.Itoa(port))

	// if there is only one value and it most be the hostname, remove all
	if client.LLen("frontend:"+fullname).Val() == 1 {
		client.Del("frontend:" + fullname)
	}

	return nil
}
