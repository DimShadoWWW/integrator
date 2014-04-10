package main

import (
	// "errors"
	// "encoding/json"
	// "fmt"
	// "github.com/dotcloud/docker/utils"
	"github.com/deckarep/golang-set"
	"github.com/dotcloud/docker/engine"
	"github.com/dotcloud/docker/nat"
	"github.com/dotcloud/docker/utils"
	"github.com/fsouza/go-dockerclient"
	"github.com/stevedomin/termtable"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
	// "net/url"
)

var _ = log.Print // for debugging, remove
type Error string

func (e Error) Error() string {
	return string(e)
}

type Lib struct {
	address string
	cfg     []string
	client  *docker.Client
	// pids    PidLib
	localip string
}

func NewDockerLib(address string) Lib {
	c, err := docker.NewClient(address)
	if err != nil {
		panic(err)
	}
	ifs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	var ip string
	ip = "aa"
	for _, v := range ifs {
		if v.Name != "lo" {
			addrs, _ := v.Addrs()
			if len(addrs) == 0 {
				continue
			}
			ip = addrs[0].String()
		}
	}
	return Lib{address: address, client: c, localip: ip}
}

func (l *Lib) Start(svcName string) error {
	// image := l.cfg[svcName].Image
	// ports := l.cfg[svcName].Ports
	// env := l.cfg[svcName].Env

	// if l.pids.hasPid(svcName) {
	// 	return errors.New("Service " + svcName + " already running")
	// }

	// // Start Dependency Containers
	// if err := l.startDeps(svcName); err != nil {
	// 	return err
	// }

	// // Create Container
	// config := docker.Config{
	// 	Image: image,
	// 	Env:   l.getEnv(env),
	// }
	// opts := docker.CreateContainerOptions{Config: &config}

	// container, err := l.client.CreateContainer(opts)
	// if err != nil {
	// 	return err
	// }

	// // Start Container
	// links, err := l.getLinks(svcName)
	// if err != nil {
	// 	return err
	// }
	// hostConfig := docker.HostConfig{
	// 	PortBindings: l.getPortBindings(ports),
	// 	Links:        links,
	// }
	// err = l.client.StartContainer(container.ID, &hostConfig)
	// if err != nil {
	// 	return err
	// }

	// for a, b := range hostConfig.PortBindings[docker.Port(80)] {
	// 	fmt.Printf("%v\t", a)
	// 	fmt.Printf("%v\n", b)
	// }
	// // l.redis.ReddisAdd(env["HOST"], l.localip+":"+string(hostConfig.PortBindings[docker.Port(80)][].HostPort)

	// if err = l.pids.setPid(svcName, container.ID); err != nil {
	// 	return err
	// }
	return nil
}

func (l *Lib) Stop(svcName string) error {
	// if !l.pids.hasPid(svcName) {
	// 	return errors.New("Service not running")
	// }

	// id, err := l.pids.getPid(svcName)
	// if err != nil {
	// 	return err
	// }

	// if err = l.client.StopContainer(id, 5); err != nil {
	// 	return err
	// }

	// if err = l.pids.removePid(svcName); err != nil {
	// 	return err
	// }
	return nil
}

func (l *Lib) startDeps(svcName string) error {
	// deps := l.cfg[svcName].Deps

	// for _, svcName := range deps {
	// 	if !l.pids.hasPid(svcName) {

	// 		if err := l.Start(svcName); err != nil {
	// 			return err
	// 		}
	// 		fmt.Println("Dep " + svcName + " started")
	// 	}
	// }
	return nil
}

func (l *Lib) getContainerName(svcName string) (string, error) {
	c, err := docker.NewClient(l.address)
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

func (l *Lib) getPortBindings(ports map[string]string) map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for internal, external := range ports {
		portBinding := []docker.PortBinding{}
		if external != "" {
			portBinding = []docker.PortBinding{docker.PortBinding{HostIp: "0.0.0.0", HostPort: external}}
		}
		port := docker.Port(nat.NewPort("tcp", internal))

		portBindings[port] = portBinding
	}

	return portBindings
}

func (l *Lib) getEnv(env map[string]string) []string {
	envFlat := make([]string, 0, 10)
	for k, v := range env {
		envFlat = append(envFlat, k+"="+v)
	}
	return envFlat
}

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

func (l *Lib) getContainerID(name string) (string, error) {

	conts, _ := l.client.ListContainers(docker.ListContainersOptions{})
	for _, cont := range conts {
		for _, n := range cont.Names {
			if n == name {
				return cont.ID, nil
			}
		}
	}
	return "", Error("Container not found")
}

// This work with api verion < v1.7 and > v1.9
type APIImages struct {
	ID          string   `json:"Id"`
	RepoTags    []string `json:",omitempty"`
	Created     int64
	Size        int64
	VirtualSize int64
	ParentId    string `json:",omitempty"`
	Repository  string `json:",omitempty"`
	Tag         string `json:",omitempty"`
}

type APIPort struct {
	PrivatePort int64
	PublicPort  int64
	Type        string
	IP          string
}

type APIContainers struct {
	ID         string `json:"Id"`
	Image      string
	Command    string
	Created    int64
	Status     string
	Ports      []APIPort
	SizeRw     int64
	SizeRootFs int64
	Names      string
}

func (l *Lib) ListImages() {
	imgs, _ := l.client.ListImages(false)
	for _, img := range imgs {
		// fmt.Println(len(img.RepoTags[0]))
		// fmt.Print("%v\n\n"+ img)
		// fmt.Print("%v\n\n"+ img.RepoTags)
		if img.RepoTags[0] != "<none>:<none>" {
			color.Println("@rID: "+color.ResetCode+" "+color.ResetCode+" ", img.ID)
			color.Println("@rRepoTags: "+color.ResetCode+" ", img.RepoTags)
			// color.Println("@rCreated: "+ img.Created)
			// color.Println("@rSize: "+ img.Size)
			// color.Println("@rVirtualSize: "+ img.VirtualSize/(1000*1000)+ "MB")
			// color.Println("@rParentId: "+ img.ParentId)
			// color.Println("@rRepository: "+ img.Repository)
		}
	}
}

func (l *Lib) RemoveContainers(ids []string) error {
	for _, id := range ids {
		color.Println("@bREMOVING: "+color.ResetCode, id)
		err := l.client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			return err
		}
	}
	color.Println("Done")
	return nil
}

func (l *Lib) RemoveImages(ids []string) error {
	for _, id := range ids {
		color.Println("@bREMOVING: "+color.ResetCode, id)
		err := l.client.RemoveImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			return err
		}
	}
	color.Println("Done")
	return nil
}
func (l *Lib) GetContainers(all bool) ([]APIContainers, error) {
	query := "0"
	if all {
		query = "1"
	}

	req, err := http.NewRequest("GET", "/containers/json?all="+query, nil)

	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mine")
	var resp *http.Response
	dial, err := net.Dial("unix", "/var/run/docker.sock")
	// "unix", "/var/run/docker.sock")
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	resp, err = clientconn.Do(req)
	// fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}

	outs := engine.NewTable("Created", 0)
	if _, err := outs.ReadListFrom(body); err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
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

func (l *Lib) GetImages() ([]APIImages, error) {

	req, err := http.NewRequest("GET", "/images/json", nil)

	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mine")
	var resp *http.Response
	dial, err := net.Dial("unix", "/var/run/docker.sock")
	// "unix", "/var/run/docker.sock")
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	resp, err = clientconn.Do(req)
	// fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
		return nil, err
	}

	outs := engine.NewTable("Created", 0)
	if _, err := outs.ReadListFrom(body); err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
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
		// ParentId    string `json:",omitempty"`
		// Repository  string `json:",omitempty"`
		// Tag         string `json:",omitempty"`
		c.ID = out.Get("Id")
		c.ParentId = out.Get("ParentId")
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

func (l *Lib) Status() {
	color.Println("Docker status:")
	// list all containers
	cont_run, err := l.GetContainers(false)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
	}
	cont_all, err := l.GetContainers(true)
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
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

	color.Println("@bContainers"+color.ResetCode+":", len(all),
		"\t@rRunning"+color.ResetCode+":", len(running),
		"@yStopped"+color.ResetCode+":", len(all.Difference(running)))

	t := termtable.NewTable(nil, nil)
	t.SetHeader([]string{"ID", "Image", "status"})
	for _, c := range cont_all {
		t.AddRow([]string{c.ID, c.Image, c.Status})
	}
	color.Println(t.Render())

	imgs_all, err := l.GetImages()
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
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
		t1.AddRow([]string{i.ID, i.RepoTags[0], utils.HumanDuration(time.Now().UTC().Sub(time.Unix(i.Created, 0)))})
	}
	color.Println(t1.Render())
}

func (l *Lib) CleanContainers() []string {
	// list all containers
	cont_run, err := l.GetContainers(false)
	if err != nil {
		color.Errorf("@rERROR: "+color.ResetCode, err)
	}
	cont_all, err := l.GetContainers(true)
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
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

	color.Println("@bContainers"+color.ResetCode+":", len(all),
		"\t@rRunning"+color.ResetCode+":", len(running),
		"@yStopped"+color.ResetCode+":", len(all.Difference(running)))
	var ids []string
	for id, _ := range all.Difference(running) {
		ids = append(ids, id.(string))
	}
	return ids
}

func (l *Lib) CleanImages() []string {

	imgs_all, err := l.GetImages()
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, err)
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
