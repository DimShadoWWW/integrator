package main

import (
	"encoding/json"
	"flag"
	"github.com/GeertJohan/go.rice"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	// "regexp"
	// "os"
	// "text/template"
)

var (
	// address to access docker API
	address string

	// port to listen on
	port string

	client Lib
)

func main() {
	// parse command line flags
	flag.StringVar(&address, "address", "unix:///var/run/docker.sock", "docker address")
	flag.StringVar(&port, "port", "8080", "docker address")
	flag.Parse()

	// flag.Usage = func() {
	// 	fmt.Printf("Usage: %s [options] command\n", os.Args[0])
	// 	flag.PrintDefaults()
	// 	fmt.Println("")
	// 	fmt.Println("Commands: ")
	// 	fmt.Println("  clearall, ca: Remove non running containers and old images")
	// 	fmt.Println("  clearcontainers, cc: Remove non running containers")
	// 	fmt.Println("  clearimages, ci: Remove nontagged images")
	// 	fmt.Println("  status, st: Docker status")
	// }
	client = NewDockerLib(address)
	// switch flag.Arg(0) {
	// case "clearall", "ca":
	// 	containers := client.CleanContainers()
	// 	client.RemoveContainers(containers)
	// 	images := client.CleanImages()
	// 	client.RemoveImages(images)
	// case "clearcontainers", "cc":
	// 	containers := client.CleanContainers()
	// 	client.RemoveContainers(containers)
	// case "clearimages", "ci":
	// 	images := client.CleanImages()
	// 	client.RemoveImages(images)
	// case "status", "st":
	// 	client.Status()
	// default:
	// 	flag.Usage()
	// }
	//

	r := mux.NewRouter()
	r.HandleFunc("/api/containers", ContainerStatusHandler)
	r.HandleFunc("/api/status", StatusHandler)
	r.HandleFunc("/api/containers/del/{id}", DelContainerHandler)
	r.HandleFunc("/api/containers/stop/{id}", StopContainerHandler)
	r.HandleFunc("/api/containers/start/{id}", StartContainerHandler)
	r.HandleFunc("/api/clean", CleanHandler)
	r.HandleFunc("/api/containers/clean", CleanContainersHandler)
	r.HandleFunc("/api/images/clean", CleanImagesHandler)
	http.Handle("/api/", r)

	http.Handle("/", http.FileServer(rice.MustFindBox("public").HTTPBox()))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

type ContainerStatus struct {
	Containers []APIContainers
	Status     map[string]int
}

type ImagesStatus struct {
	Images []APIImages
	Status map[string]int
}

type Status struct {
	Containers ContainerStatus
	Images     ImagesStatus
}

func StatusHandler(c http.ResponseWriter, r *http.Request) {
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
	var contstatus []APIContainers
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
	var imgstatus []APIImages
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
}

func ContainerStatusHandler(c http.ResponseWriter, r *http.Request) {
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
}

func StartContainerHandler(c http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status := map[string]string{"status": "0"}
	err := client.client.StartContainer(id, nil)
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
}

func StopContainerHandler(c http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status := map[string]string{"status": "0"}

	err := client.client.StopContainer(id, 20)
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
}

func DelContainerHandler(c http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status := map[string]string{"status": "0"}
	err := client.client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
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
}

func CleanHandler(c http.ResponseWriter, r *http.Request) {
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
}

func CleanContainersHandler(c http.ResponseWriter, r *http.Request) {
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
}

func CleanImagesHandler(c http.ResponseWriter, r *http.Request) {
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
