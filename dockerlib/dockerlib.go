package dockerlib

import (
	"github.com/deckarep/golang-set"
	"github.com/dotcloud/docker/engine"
	"github.com/dotcloud/docker/nat"
	"github.com/dotcloud/docker/pkg/units"
	"github.com/fsouza/go-dockerclient"
	"github.com/stevedomin/termtable"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)
import log "github.com/cihub/seelog"

type Error string // Error type

func (e Error) Error() string {
	return string(e)
}

// Docker lib struct
type Lib struct {
	Address string
	Cfg     []string
	Client  *docker.Client
	// pids    PidLib
	Localip string
}

// This work with api verion < v1.7 and > v1.9
type APIImages struct {
	ID          string   `json:"Id"`
	RepoTags    []string `json:",omitempty"`
	Created     int64
	Size        int64
	VirtualSize int64
	ParentID    string `json:",omitempty"`
	Repository  string `json:",omitempty"`
	Tag         string `json:",omitempty"`
}

// dockey port struct
type APIPort struct {
	PrivatePort int64
	PublicPort  int64
	Type        string
	IP          string
}

// container struct
type APIContainers struct {
	ID         string    `json:"ID"  binding:"required`
	Image      string    `json:"Image"  binding:"required"`
	Command    string    `json:"Command"`
	Created    int64     `json:"Created"`
	Status     string    `json:"Status"`
	Ports      []APIPort `json:"Ports"`
	SizeRw     int64     `json:"SizeRw"`
	SizeRootFs int64     `json:"SizeRootFs"`
	Names      string    `json:"Names"`
	DNS        string    `json:"Dns"`
	DNSSearch  string    `json:"DnsSearch"`
	Env        []string  `json:"Env"`
	Links      []string  `json:"Links"`
	Volume     []string  `json:"Volume"`
	Detach     bool      `json:"Detach"`
	User       string    `json:"User"`
	Workdir    string    `json:"Workdir"`
	Hostname   string    `json:"Hostname"`
	Privileged bool      `json:"Privileged"`
}

// Get a DockerLib struct
func NewDockerLib(address string) (Lib, error) {
	c, err := docker.NewClient(address)
	if err != nil {
		return Lib{}, err
	}

	ip, err := GetLocalIp("8.8.8.8:53")
	if err != nil {
		return Lib{}, err
	}

	return Lib{Address: address, Client: c, Localip: ip}, nil
}

// Get container name
func (l *Lib) getContainerName(svcName string) (string, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return "", err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return "", err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return "", err
	}
	return container.Name, nil
}

// Get container ports
func (l *Lib) GetContainerPorts(svcName string) (map[docker.Port][]docker.PortBinding, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return nil, err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return nil, err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return nil, err
	}

	return container.NetworkSettings.Ports, nil
}

// Get container tcp port
func (l *Lib) GetContainerTcpPort(svcName string, port int) (string, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return "", err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return "", err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return "", err
	}

	for k, p := range container.NetworkSettings.Ports {
		if k.Proto() == "tcp" {
			cport, err := strconv.Atoi(k.Port())
			if err != nil {
				return "", err
			}
			if cport == port {
				return p[0].HostPort, nil
			}

		}
	}

	return "0", Error("Tcp port not found in container")
}

// Check if container has this port open
func (l *Lib) GetContainerCheckOpenPort(svcName string, port int) (bool, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return false, err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return false, err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return false, err
	}

	for k, _ := range container.NetworkSettings.Ports {
		cport, err := strconv.Atoi(k.Port())
		if err != nil {
			return false, err
		}
		if cport == port {
			return true, nil
		}
	}

	return false, Error("Port not open in container")
}

// Get container udp port
func (l *Lib) GetContainerUdpPort(svcName string, port int) (string, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return "", err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return "", err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return "", err
	}

	for k, p := range container.NetworkSettings.Ports {
		if k.Proto() == "udp" {
			cport, err := strconv.Atoi(k.Port())
			if err != nil {
				return "", err
			}
			if cport == port {
				return p[0].HostPort, nil
			}

		}
	}

	return "", Error("Udp port not found in container")
}

// Get container's ip address in docker0
func (l *Lib) GetContainerIpaddress(svcName string) (string, error) {
	c, err := docker.NewClient(l.Address)
	if err != nil {
		return "", err
	}
	id, err := l.getContainerID(svcName)
	if err != nil {
		return "", err
	}
	container, err := c.InspectContainer(id)
	if err != nil {
		return "", err
	}

	return container.NetworkSettings.IPAddress, nil
}

// Get container's port bindings
func (l *Lib) getPortBindings(ports map[string]string) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for internal, external := range ports {
		portBinding := []docker.PortBinding{}
		if external != "" {
			portBinding = []docker.PortBinding{docker.PortBinding{HostIP: "0.0.0.0", HostPort: external}}
		}
		port := docker.Port(nat.NewPort("tcp", internal))

		portBindings[port] = portBinding
	}

	return portBindings
}

// Get container environmet var
func (l *Lib) getEnv(env map[string]string) []string {
	envFlat := make([]string, 0, 10)
	for k, v := range env {
		envFlat = append(envFlat, k+"="+v)
	}
	return envFlat
}

// Get container's links
func (l *Lib) getLinks(svcName string) ([]string, error) {
	// deps := l.cfg[svcName].Deps
	// links := make([]string, 0, 10)
	// for _, svcName := range deps {
	// 	if !l.pids.hasPid(svcName) {
	// 		return links, errors.New("Dep not running: " + svcName)
	// 	}
	// 	name, err := l.getContainerName(svcName)
	// 	if err != nil {
	// 		return links, err
	// 	}
	// 	links = append(links, name+":"+svcName)
	// }
	// return links, nil
	return nil, nil
}

// Get container id
func (l *Lib) getContainerID(name string) (string, error) {

	conts, _ := l.Client.ListContainers(docker.ListContainersOptions{})
	for _, cont := range conts {
		for _, n := range cont.Names {
			if n == "/"+name {
				return cont.ID, nil
			}
		}
	}
	return "", Error("Container not found")
}

// List local docker images
func (l *Lib) ListImages() {
	imgs, _ := l.Client.ListImages(docker.ListImagesOptions{})
	for _, img := range imgs {
		if img.RepoTags[0] != "<none>:<none>" {
			color.Println("@rID: "+color.ResetCode+" "+color.ResetCode+" ", img.ID)
			color.Println("@rRepoTags: "+color.ResetCode+" ", img.RepoTags)
			// color.Println("@rCreated: "+ img.Created)
			// color.Println("@rSize: "+ img.Size)
			// color.Println("@rVirtualSize: "+ img.VirtualSize/(1000*1000)+ "MB")
			// color.Println("@rParentID: "+ img.ParentID)
			// color.Println("@rRepository: "+ img.Repository)
		}
	}
}

// Pull docker image
func (l *Lib) PullImage(name string) error {
	imageData := strings.Split(name, ":")
	name = imageData[0]
	tag := "latest"
	if len(imageData) > 1 {
		tag = imageData[1]
	}
	return l.Client.PullImage(docker.PullImageOptions{Repository: name, Tag: tag}, docker.AuthConfiguration{})
}

// Remove docker container
func (l *Lib) RemoveContainers(ids []string) error {
	for _, id := range ids {
		color.Println("@bREMOVING: "+color.ResetCode, id)
		err := l.Client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			log.Error(err)
			return err
		}
	}
	color.Println("Done")
	return nil
}

// start a container
func (l *Lib) StartContainer(id string) error {
	color.Println("@bREMOVING: "+color.ResetCode, id)
	err := l.Client.StartContainer(id, nil)
	//.RemoveContainer(docker.RemoveContainerOptions{ID: id})
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return err
	}
	color.Println("Done")
	return nil
}

// Remove docker images
func (l *Lib) RemoveImages(ids []string) error {
	for _, id := range ids {
		color.Println("@bREMOVING: "+color.ResetCode, id)
		err := l.Client.RemoveImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			log.Error(err)
			return err
		}
	}
	color.Println("Done")
	return nil
}

// Get docker containers
func (l *Lib) GetContainers(all bool) ([]APIContainers, error) {
	query := "0"
	if all {
		query = "1"
	}

	req, err := http.NewRequest("GET", "/containers/json?all="+query, nil)

	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mine")
	var resp *http.Response
	dial, err := net.Dial("unix", "/var/run/docker.sock")
	// "unix", "/var/run/docker.sock")
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	resp, err = clientconn.Do(req)
	// fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}

	outs := engine.NewTable("Created", 0)
	if _, err := outs.ReadListFrom(body); err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	var containers []APIContainers
	for _, out := range outs.Data {
		var c APIContainers
		c.ID = out.Get("Id")
		for _, n := range out.GetList("Names") {
			if strings.Index(strings.Replace(n, "/", "", 1), "/") == -1 {
				c.Names = strings.Replace(n, "/", "", 1)
				break
			}
		}
		c.Image = out.Get("Image")
		c.Command = out.Get("Command")
		c.Created = out.GetInt64("Created")
		c.Status = out.Get("Status")
		c.SizeRw = out.GetInt64("SizeRw")
		c.SizeRootFs = out.GetInt64("SizeRootFs")

		ports := engine.NewTable("", 0)
		ports.ReadListFrom([]byte(out.Get("Ports")))
		ports.SetKey("PublicPort")
		ports.Sort()
		for _, port := range ports.Data {
			var p APIPort
			p.IP = port.Get("IP")
			p.PrivatePort = port.GetInt64("PrivatePort")
			p.PublicPort = port.GetInt64("PublicPort")
			p.Type = port.Get("Type")
			c.Ports = append(c.Ports, p)
		}

		containers = append(containers, c)
	}
	return containers, nil
}

// Get local docker images
func (l *Lib) GetImages() ([]APIImages, error) {

	req, err := http.NewRequest("GET", "/images/json", nil)

	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mine")
	var resp *http.Response
	dial, err := net.Dial("unix", "/var/run/docker.sock")
	// "unix", "/var/run/docker.sock")
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	resp, err = clientconn.Do(req)
	// fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}

	outs := engine.NewTable("Created", 0)
	if _, err := outs.ReadListFrom(body); err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
		return nil, err
	}
	var images []APIImages
	for _, out := range outs.Data {
		var c APIImages

		// ID          string   `json:"Id"`
		// RepoTags    []string `json:",omitempty"`
		// Created     int64
		// Size        int64
		// VirtualSize int64
		// ParentID    string `json:",omitempty"`
		// Repository  string `json:",omitempty"`
		// Tag         string `json:",omitempty"`
		c.ID = out.Get("Id")
		c.ParentID = out.Get("ParentID")
		c.Repository = out.Get("Repository")
		c.Tag = out.Get("Tag")
		c.RepoTags = out.GetList("RepoTags")
		c.Created = out.GetInt64("Created")
		c.Size = out.GetInt64("Size")
		c.VirtualSize = out.GetInt64("VirtualSize")

		images = append(images, c)
	}
	return images, nil
}

// Get status of all docker containers and images
func (l *Lib) Status() {
	color.Println("Docker status:")
	// list all containers
	cont_run, err := l.GetContainers(false)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
	}
	cont_all, err := l.GetContainers(true)
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
		log.Error(err)
	}

	// // list running containers
	running := mapset.NewSet()
	for _, cont := range cont_run {
		// color.Println("@rID: "+color.ResetCode, cont.ID)
		running.Add(cont.ID)
	}
	all := mapset.NewSet()
	for _, cont := range cont_all {
		// color.Println("@bID: "+color.ResetCode, cont.ID)
		all.Add(cont.ID)
	}

	color.Println("@bContainers"+color.ResetCode+":", all.Cardinality(),
		"\t@rRunning"+color.ResetCode+":", running.Cardinality(),
		"@yStopped"+color.ResetCode+":", all.Difference(running).Cardinality())

	t := termtable.NewTable(nil, nil)
	t.SetHeader([]string{"ID", "Image", "status"})
	for _, c := range cont_all {
		t.AddRow([]string{c.ID, c.Image, c.Status})
	}
	color.Println(t.Render())

	imgs_all, err := l.GetImages()
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
		log.Error(err)
	}

	imgs_good := 0
	imgs_none := 0
	for _, img := range imgs_all {
		if img.RepoTags[0] != "<none>:<none>" {
			imgs_good = imgs_good + 1
		} else {
			imgs_none = imgs_none + 1
		}
	}

	color.Println("@bImages"+color.ResetCode+":", len(imgs_all),
		"\t@rGood"+color.ResetCode+":", imgs_good,
		"@yTrash"+color.ResetCode+":", imgs_none)

	t1 := termtable.NewTable(nil, nil)
	t1.SetHeader([]string{"ID", "Repository:Tag", "Created"})
	for _, i := range imgs_all {
		t1.AddRow([]string{i.ID, i.RepoTags[0], units.HumanDuration(time.Now().UTC().Sub(time.Unix(i.Created, 0)))})
	}
	color.Println(t1.Render())
}

// Remove failed containers
func (l *Lib) CleanContainers() []string {
	// list all containers
	cont_run, err := l.GetContainers(false)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		log.Error(err)
	}
	cont_all, err := l.GetContainers(true)
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
		log.Error(err)
	}

	// // list running containers
	running := mapset.NewSet()
	for _, cont := range cont_run {
		// color.Println("@rID: "+color.ResetCode, cont.ID)
		running.Add(cont.ID)
	}
	all := mapset.NewSet()
	for _, cont := range cont_all {
		// color.Println("@bID: "+color.ResetCode, cont.ID)
		all.Add(cont.ID)
	}

	color.Println("@bContainers"+color.ResetCode+":", all.Cardinality(),
		"\t@rRunning"+color.ResetCode+":", running.Cardinality(),
		"@yStopped"+color.ResetCode+":", all.Difference(running).Cardinality())
	var ids []string
	for id := range all.Difference(running).Iter() {
		ids = append(ids, id.(string))
	}
	return ids
}

// Remove docker images without tag (headless images or garbage from build process)
func (l *Lib) CleanImages() []string {

	imgs_all, err := l.GetImages()
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
		log.Error(err)
	}

	var ids []string
	for _, img := range imgs_all {
		if img.RepoTags[0] == "<none>:<none>" {
			ids = append(ids, img.ID)
		}
	}

	color.Println("To remove: @r", len(ids), color.ResetCode, " images")
	return ids
}

// build a docker image
func (l *Lib) BuildImage(name string) (string, error) {
	// imgs_all, err := l.Client.BuildImage(opts)
	// if err != nil {
	// 	color.Errorf("@bERROR: "+color.ResetCode, err)
	//  log.Error(err)
	// }

	// var ids []string
	// for _, img := range imgs_all {
	// 	if img.RepoTags[0] == "<none>:<none>" {
	// 		ids = append(ids, img.ID)
	// 	}
	// }

	// color.Println("To remove: @r", len(ids), color.ResetCode, " images")
	return "", nil
}

// Get server's external ip address
func GetLocalIp(server string) (string, error) {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return "", err
	}

	// conn.LocalAddr().String() returns ip_address:port
	return net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0]).String(), nil
}
