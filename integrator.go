package main

import (
	"encoding/json"
	"flag"
	"github.com/DimShadoWWW/integrator/dockerlib"
	// "github.com/DimShadoWWW/integrator/etcdlib"
	"github.com/GeertJohan/go.rice"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)
import log "github.com/cihub/seelog"

type Configuration struct {
	Address string
	Port    string
	ApiKey  string
}

var (
	// address to access docker API
	address string

	// port to listen on
	port string

	client dockerlib.Lib

	APIKey string
)

func main() {

	var logConfig = `
 	<seelog type="asyncloop" minlevel="info">
 	<outputs>
 	<file path="/home/core/integrator.log" formatid="main"/>
 	</outputs>
 	<formats>
 	<format id="main" format="%Date %Time [%Level] %Msg%n"/>
 	</formats>
 	</seelog>`

	logger, err := log.LoggerFromConfigAsBytes([]byte(logConfig))

	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, "Error during logger config load: ", err.Error())
	}

	log.ReplaceLogger(logger)

	defer log.Flush()

	defaultAddress := "unix:///var/run/docker.sock"
	defaultPort := "8080"
	defaultAPIKey := "CHANGE_ME"

	if _, err := os.Stat("config.json"); err == nil {
		file, _ := os.Open("config.json")
		decoder := json.NewDecoder(file)
		configuration := Configuration{}
		err := decoder.Decode(&configuration)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		defaultAddress = configuration.Address
		defaultPort = configuration.Port
		defaultAPIKey = configuration.ApiKey
	}

	// parse command line flags
	flag.StringVar(&address, "address", defaultAddress, "docker address")
	flag.StringVar(&port, "port", defaultPort, "docker address")
	flag.StringVar(&APIKey, "apikey", defaultAPIKey, "docker address")
	flag.Parse()

	color.Println(APIKey)
	color.Println(address)
	color.Println(port)

	client = dockerlib.NewDockerLib(address)

	checkHeaderThenServe := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set some header.
			if r.Header.Get("API-Access") == APIKey {
				w.Header().Add("Keep-Alive", "300")
				// Serve with the actual handler.
				h.ServeHTTP(w, r)
			} else {
				// http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
				color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
			}
		}
	}

	r := mux.NewRouter()
	// r.Headers("API-Access", APIKey)
	r.HandleFunc("/api/status", StatusHandler)
	r.HandleFunc("/api/containers", StatusContainerHandler)
	r.HandleFunc("/api/containers/del/{id}", DelContainerHandler)
	r.HandleFunc("/api/containers/stop/{id}", StopContainerHandler)
	r.HandleFunc("/api/containers/start/{id}", StartContainerHandler)
	r.HandleFunc("/api/containers/clean", CleanContainersHandler)
	r.HandleFunc("/api/containers/new", RunContainerHandler)
	r.HandleFunc("/api/clean", CleanHandler)
	r.HandleFunc("/api/images", StatusImageHandler)
	r.HandleFunc("/api/images/build/{name}", BuildImageHandler)
	r.HandleFunc("/api/images/del/{id}", DelImageHandler)
	r.HandleFunc("/api/images/clean", CleanImagesHandler)
	http.Handle("/api/", r)

	http.Handle("/", checkHeaderThenServe(http.FileServer(rice.MustFindBox("public").HTTPBox())))
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Error(err)
	}
}

func newClient() *etcd.Client {
	etcdclient := etcd.NewClient([]string{"https://127.0.0.1:4001"})
	etcdclient.SyncCluster()
	return etcdclient
}

type ContainerStatus struct {
	Containers []dockerlib.APIContainers
	Status     map[string]int
}

type ImagesStatus struct {
	Images []dockerlib.APIImages
	Status map[string]int
}

type Status struct {
	Containers ContainerStatus
	Images     ImagesStatus
}

func checkaccess(a string) bool {
	if a == APIKey {
		return true
	}
	return false
}

func StatusHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		a := getIpAddress(r)
		rip, err := net.ResolveTCPAddr("tcp", a+":0")
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		color.Println("@bACCESS from: "+color.ResetCode, rip.IP)

		cont_all, err := client.GetContainers(true)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		var results Status
		up := 0
		failed := 0
		down := 0
		unknown := 0
		var contstatus []dockerlib.APIContainers
		for _, c := range cont_all {
			switch strings.Split(c.Status, " ")[0] {
			case "Up":
				up = up + 1
			case "Exit", "Exited":
				if strings.Split(c.Status, " ")[1] == "0" {
					down = down + 1
				} else {
					failed = failed + 1
				}
			default:
				unknown = unknown + 1
			}
			contstatus = append(contstatus, c)
		}

		images, err := client.GetImages()
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		good := 0
		temp := 0
		var imgstatus []dockerlib.APIImages
		for _, img := range images {
			if img.RepoTags[0] != "<none>:<none>" {
				good = good + 1
			} else {
				temp = temp + 1
			}
			imgstatus = append(imgstatus, img)
		}
		results = Status{
			Containers: ContainerStatus{
				Containers: contstatus,
				Status:     map[string]int{"up": up, "down": down, "failed": failed, "unknown": unknown},
			},
			Images: ImagesStatus{
				Images: imgstatus,
				Status: map[string]int{"good": good, "temp": temp},
			},
		}

		result, err := json.Marshal(results)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		var r1 interface{}
		err = json.Unmarshal(result, &r1)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//Containers
func StatusContainerHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		cont_all, err := client.GetContainers(true)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		result, err := json.Marshal(cont_all)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func StartContainerHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := client.Client.StartContainer(id, nil)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func RunContainerHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := client.Client.StartContainer(id, nil)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func StopContainerHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}

		err := client.Client.StopContainer(id, 20)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func DelContainerHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := client.Client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func CleanContainersHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		containers := client.CleanContainers()
		client.RemoveContainers(containers)

		status := map[string]int{"status": 0}
		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//Images
func StatusImageHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		cont_all, err := client.GetImages()
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		result, err := json.Marshal(cont_all)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func BuildImageHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["name"]

		status := map[string]string{}
		id, err := client.BuildImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		} else {
			status = map[string]string{"status": "0", "id": id}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func DelImageHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := client.Client.RemoveImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func CleanImagesHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		images := client.CleanImages()
		client.RemoveImages(images)

		status := map[string]int{"status": 0}
		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//General functions
func CleanHandler(c http.ResponseWriter, r *http.Request) {
	if checkaccess(r.Header.Get("API-Access")) {
		containers := client.CleanContainers()
		client.RemoveContainers(containers)
		images := client.CleanImages()
		client.RemoveImages(images)

		status := map[string]int{"status": 0}
		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getIpAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIp
}
