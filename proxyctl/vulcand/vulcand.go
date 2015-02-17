package vulcand

import (
	"fmt"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/mailgun/go-etcd/etcd"
	"github.com/mailgun/vulcand/api"
	"github.com/mailgun/vulcand/engine"
	"github.com/mailgun/vulcand/engine/etcdng"
	"github.com/mailgun/vulcand/plugin"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	ipaddress        string
	toclean_etcdpath []string
	container_name   string
	external_domain  = "production"
)

type Command struct {
	vulcanUrl string
	client    *api.Client
	registry  *plugin.Registry
}

type VulcandHostJSON struct {
	Host string ` json:"host"`
	Port int    `json:"port"`
}

func VulcandHostAdd(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string, ttl time.Duration) error {

	// Docker internal ip address
	fmt.Println("Adding internal PROXY entry")
	if service.Id == 0 {
		container_name = service.Hostname
	} else {
		container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)
	}

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain
	route := "Host(`" + fullname + "`) && Path(`/`)"

	rand.Seed(time.Now().UnixNano())

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	for i := 0; i < 31; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			// retry until this
			if i == 30 {
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	r := plugin.NewRegistry()
	mengine, err := etcdng.New(etcdnodes, "vulcand", r, etcdng.Options{EtcdConsistency: etcd.STRONG_CONSISTENCY})
	client := api.NewClient(strings.Replace(etcdnodes[0], ":4001", ":8182", -1), r)

	var backendId, serverId string
	var frontend *engine.Frontend
	var backend *engine.Backend
	// , upstreamId string
	// var server engine.Server

	frontends, err := mengine.GetFrontends()
	if err != nil {
		return err
	}

	for _, fe := range frontends {
		if fe.Route == route {
			bk := engine.BackendKey{Id: fe.BackendId}
			b, err := client.GetBackend(bk)
			if err != nil {
				return err
			}
			srvs, err := client.GetServers(b.GetUniqueId())
			if err != nil {
				return err
			}

			for _, se := range srvs {
				if se.URL == "http://"+ipaddress+":"+strconv.Itoa(port) {
					// server was already added, return with no error
					return nil
				}
			}
			backendId = b.GetId()
			frontend = &engine.Frontend{
				Id:        fe.Id,
				BackendId: fe.BackendId,
				Route:     fe.Route,
				Type:      fe.Type,
				Settings:  fe.Settings,
			}
		}
	}

	if backendId == "" {
		settings, err := getBackendSettings()
		if err != nil {
			return err
		}

		backendId = genId()
		backend, err = engine.NewHTTPBackend(backendId, settings)
		if err != nil {
			return err
		}
	}

	if serverId == "" {
		serverId = genId()
		server, err := engine.NewServer(serverId, "http://"+ipaddress+":"+strconv.Itoa(port))
		if err != nil {
			return err
		}
		if err := client.UpsertServer(engine.BackendKey{Id: backend.GetId()}, *server, ttl); err != nil {
			return err
		}

	}

	if frontend == nil {
		settings, err := getFrontendSettings()
		if err != nil {
			return err
		}

		frontend, err = engine.NewHTTPFrontend(genId(), backend.GetId(), route, settings)
		if err != nil {
			return err
		}
		if err := client.UpsertFrontend(*frontend, ttl); err != nil {
			return err
		}
	}

	return nil
}

func VulcandHostDel(etcdnodes []string, dockeruri string, service fleet.SystemdService, port int, path string) error {

	fullname := service.Hostname + "." + external_domain + "." + service.Region + "." + service.Domain
	route := "Host(`" + fullname + "`) && Path(`/`)"

	// Docker internal ip address
	fmt.Println("Remove internal PROXY entry")
	if service.Id == 0 {
		container_name = service.Hostname
	} else {
		container_name = service.Hostname + "-" + strconv.FormatInt(service.Id, 10)
	}

	dockerclient, err := dockerlib.NewDockerLib(dockeruri)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		ipaddress, err = dockerclient.GetContainerIpaddress(container_name)
		if err != nil {
			if i == 9 {
				fmt.Println("Failed to connect to docker")
				return err
			}
		} else {
			break
		}
		// wait 10 seconds to try again
		time.Sleep(10 * time.Second)
	}

	r := plugin.NewRegistry()
	mengine, err := etcdng.New(etcdnodes, "vulcand", r, etcdng.Options{EtcdConsistency: etcd.STRONG_CONSISTENCY})
	client := api.NewClient(strings.Replace(etcdnodes[0], ":4001", ":8182", -1), r)

	var frontend *engine.Frontend
	var backend *engine.Backend
	var server *engine.Server

	frontends, err := mengine.GetFrontends()
	if err != nil {
		return err
	}

	for _, fe := range frontends {
		if fe.Route == route {
			bk := engine.BackendKey{Id: fe.BackendId}
			b, err := client.GetBackend(bk)
			if err != nil {
				return err
			}
			srvs, err := client.GetServers(b.GetUniqueId())
			if err != nil {
				return err
			}

			for _, se := range srvs {
				if se.URL == "http://"+ipaddress+":"+strconv.Itoa(port) {
					backend = &engine.Backend{
						Id:       b.Id,
						Type:     b.Type,
						Stats:    b.Stats,
						Settings: b.Settings,
					}
					server = &engine.Server{
						Id:    se.Id,
						URL:   se.URL,
						Stats: se.Stats,
					}
					frontend = &engine.Frontend{
						Id:        fe.Id,
						BackendId: fe.BackendId,
						Route:     fe.Route,
						Type:      fe.Type,
						Settings:  fe.Settings,
					}
					break
					break
				}
			}
		}
	}

	if backend != nil {
		if server != nil {
			sk := engine.ServerKey{BackendKey: engine.BackendKey{Id: backend.GetId()}, Id: server.GetId()}
			err = client.DeleteServer(sk)
			if err != nil {
				return err
			} else {
				log.Println("server deleted")
			}
		}

		srvs, err := client.GetServers(backend.GetUniqueId())
		if err != nil {
			return err
		}

		if len(srvs) == 0 {
			err = client.DeleteBackend(engine.BackendKey{Id: backend.GetId()})
			if err != nil {
				return err
			} else {
				log.Println("backend empty, deleted")
				err := client.DeleteFrontend(engine.FrontendKey{Id: frontend.GetId()})
				if err != nil {
					return err
				} else {
					log.Println("frontend deleted")
				}
			}
		} else {
			log.Println("backend not empty")
		}
	}

	return nil
}

func genId() string {
	var id int64
	for i := 0; i < 10; i++ {
		id = rand.Int63() + 1
	}
	return string(strconv.FormatInt(id, 10))
}

func getBackendSettings() (engine.HTTPBackendSettings, error) {
	s := engine.HTTPBackendSettings{}

	// readTimeout
	s.Timeouts.Read = "3"
	// dialTimeout
	s.Timeouts.Dial = "3"
	// handshakeTimeout
	s.Timeouts.TLSHandshake = "3"

	// keepAlivePeriod
	s.KeepAlive.Period = "10"
	// maxIdleConns
	s.KeepAlive.MaxIdleConnsPerHost = 20

	// tlsSettings, err := getTLSSettings(c)
	// if err != nil {
	// 	return s, err
	// }
	// s.TLS = tlsSettings
	return s, nil
}

func getFrontendSettings() (engine.HTTPFrontendSettings, error) {
	s := engine.HTTPFrontendSettings{}

	// maxMemBodyKB "maximum request size to cache in memory, in KB"
	// s.Limits.MaxMemBodyBytes = int64(c.Int("maxMemBodyKB") * 1024)
	// maxBodyKB "maximum request size to allow for a frontend, in KB"
	// s.Limits.MaxBodyBytes = int64(c.Int("maxBodyKB") * 1024)

	// failoverPredicate "predicate that defines cases when failover is allowed"
	// s.FailoverPredicate = c.String("failoverPredicate")
	// forwardHost "hostname to set when forwarding a request"
	// s.Hostname = c.String("forwardHost")
	// trustForwardHeader "allows copying X-Forwarded-For header value from the original request"
	s.TrustForwardHeader = false

	return s, nil
}
