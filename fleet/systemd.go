package fleet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	service = `[Unit]
Description={{.Description}}
After=docker.service
Requires=docker.service
{{range .Deps}}After={{.}}.service
Requires={{.}}.service{{end}}
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
TimeoutStartSec=0
ExecStart=/bin/bash -c '/usr/bin/docker start -a {{.Name|lower}} || /usr/bin/docker run --name {{.Name|lower}}{{if .Hostname}} -h {{.Hostname|lower}}.{{.Domain|lower}} {{end}}{{if .Privileged}} --privileged {{end}}{{if.Volumes}}{{range .Volumes}}{{.|volumeExpand}}{{end}}{{end}} {{if .Ports}}{{range .Ports}}{{.|portExpand}}{{end}}{{end}} {{if .Links}}{{range .Links}}{{.|linkExpand}}{{end}}{{end}} {{.ImageName}} {{.Command}}'
ExecStop=/bin/bash -c '/usr/bin/docker stop {{.Name|lower}};/usr/bin/docker rm {{.Name|lower}}'


{{if .Conflicts}}[X-Fleet]
{{range .Conflicts}}
X-Conflicts={{.}}.service{{end}}{{end}}
`

	dns_service = `[Unit]
Description={{.Description}} presence service
BindsTo={{.Name}}.service
After={{.Name}}.service
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
TimeoutStartSec=0
ExecStart=/usr/bin/etcdctl set /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}} '{ \"Host\": \"%H\", \"Port\": {{$httpport}}, \"Priority\": \"{{$priority}}\" }'
ExecStartPost=/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}} '{ \"Host\": \"%H\", \"Port\": {{$httpport}}, \"Priority\": \"{{$priority}}\" }'
ExecStop=/usr/bin/etcdctl rm /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}}
ExecStopPost=/usr/bin/etcdctl rm /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}}

[X-Fleet]
X-ConditionMachineOf={{.Name}}.service
`
	proxy_service = `[Unit]
Description={{.Description}} presence service
BindsTo={{.Name}}-dns.service
After={{.Name}}-dns.service

{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
TimeoutStartSec=0
ExecStart=/home/core/proxyctl -cmd=add -hostname={{$hostname}} -domain={{$domain}} -id={{$id}} -port={{$httpport}}
ExecStop=/home/core/proxyctl -cmd=del -hostname={{$hostname}} -domain={{$domain}} -id={{$id}} -port={{$httpport}}

[X-Fleet]
X-ConditionMachineOf={{.Name}}.service
`
	service_discovery = `
[Unit]
Description={{.Description}} presence service
BindsTo={{.Name}}.service
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
ExecStart=/bin/sh -c "while true; do {{if .HttpPort}}/usr/bin/etcdctl set /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}} '{ \"Host\": \"%H\", \"Port\": {{$httpport}}, \"Priority\": \"{{$priority}}\" }' --ttl 60;{{else}}{{if .Ports}}{{range .Ports}}/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name|lower}}/{{.HostPort}} '{ \"Host\": \"%H\", \"Port\": {{.HostPort}}, \"Priority\": \"{{$priority}}\" }' --ttl 60;{{end}}{{else}}/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name|lower}} '{ \"Host\": \"%H\", \"Priority\": \"{{$priority}}\" }' --ttl 60;{{end}}{{end}}sleep 45;done"
ExecStop=/bin/sh -c "{{if .HttpPort}}/usr/bin/etcdctl rm /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}};{{else}}{{if .Ports}}{{range .Ports}}/usr/bin/etcdctl rm /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name}}-{{$id}}/{{.HostPort}};{{end}}/usr/bin/etcdctl rmdir /services/{{$hostname|lower}}.{{$domain|lower}}/{{$name}}-{{$id}}{{else}}/usr/bin/etcdctl rm /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}}{{end}}{{end}}"

[X-Fleet]
X-ConditionMachineOf={{.Name}}.service
`
)

type Volume struct {
	Id           int64
	LocalDir     string
	ContainerDir string
}

type Port struct {
	Id            int64
	HostPort      string
	ContainerPort string
	Protocol      string
}

type Link struct {
	Id    int64
	Name  string
	Alias string
}
type Env struct {
	Id    int64
	Name  string
	Value string
}

type SystemdService struct {
	Id          int64
	Name        string
	Description string
	Command     string
	ImageName   string
	Hostname    string
	Domain      string
	Conflicts   []string
	Deps        []string
	Ports       []Port
	Volumes     []Volume
	Variables   []Env
	Links       []Link
	Privileged  bool
	Priority    int
	HttpPort    int
	Region      string
	// IncludeFleet bool
}

func VolumeExpander(args ...interface{}) string {
	line := ""
	for _, i := range args {
		j, _ := i.(Volume)

		if j.LocalDir != "" && j.ContainerDir != "" {
			line = line + fmt.Sprintf(" -v %s:%s", j.LocalDir, j.ContainerDir)
		}
	}

	return line
}

func PortExpander(args ...interface{}) string {
	line := ""
	for _, i := range args {
		j, _ := i.(Port)
		line = line + " -p "
		if j.HostPort != "" {
			line = line + j.HostPort + ":" + j.ContainerPort
		} else {
			line = line + j.ContainerPort
		}
		if j.Protocol != "" {
			line = line + "/" + j.Protocol
		}
	}

	return line
}

func LinkExpander(args ...interface{}) string {
	line := ""
	for _, i := range args {
		j, _ := i.(Link)

		if j.Name != "" && j.Alias != "" {
			line = line + fmt.Sprintf(" --link %s:%s", j.Name, j.Alias)
		}
	}

	return line
}

func VarExpander(args ...interface{}) string {
	line := ""
	for _, i := range args {
		j, _ := i.(Env)

		if j.Name != "" && j.Value != "" {
			line = line + fmt.Sprintf(" -e %s=\"%s\"", j.Name, j.Value)
		}
	}

	return line
}

func Lower(args ...interface{}) string {
	val, _ := args[0].(string)

	return strings.ToLower(val)
}

func Dns2Path(args ...interface{}) string {
	val, _ := args[0].(string)

	hostpath := strings.Split(strings.ToLower(val), ".")
	for i, j := 0, len(hostpath)-1; i < j; i, j = i+1, j-1 {
		hostpath[i], hostpath[j] = hostpath[j], hostpath[i]
	}
	path := strings.Join(hostpath, "/")

	return path
}

func CreateSystemdFiles(system SystemdService, outdir string) []string {

	t := template.New("Systemd service")
	// add our function
	t = t.Funcs(template.FuncMap{
		"volumeExpand": VolumeExpander,
		"portExpand":   PortExpander,
		"linkExpand":   LinkExpander,
		"varExpand":    VarExpander,
		"lower":        Lower,
		"dns2path":     Dns2Path,
	})
	t, err := t.Parse(service)
	checkError(err)

	fname := outdir + strings.Replace(system.Name, "{{ID}}", strconv.FormatInt(system.Id, 10), -1) + ".service"
	f, err := os.Create(fname)
	checkError(err)
	defer f.Close()
	err = t.Execute(io.Writer(f), system)
	checkError(err)
	service_files := []string{fname}

	// dns
	t = template.New("Systemd dns service")
	t = t.Funcs(template.FuncMap{
		"volumeExpand": VolumeExpander,
		"portExpand":   PortExpander,
		"linkExpand":   LinkExpander,
		"varExpand":    VarExpander,
		"lower":        Lower,
		"dns2path":     Dns2Path,
	})
	t, err = t.Parse(dns_service)
	checkError(err)

	fname = outdir + strings.Replace(system.Name, "{{ID}}", strconv.FormatInt(system.Id, 10), -1) + "-dns.service"
	f, err = os.Create(fname)
	checkError(err)
	defer f.Close()
	service_files = append(service_files, fname)

	err = t.Execute(io.Writer(f), system)
	checkError(err)
	fmt.Println(system.Name)

	// enable proxy only if httport > 0
	if system.HttpPort > 0 {
		t = template.New("Systemd proxy service")
		t = t.Funcs(template.FuncMap{
			"volumeExpand": VolumeExpander,
			"portExpand":   PortExpander,
			"linkExpand":   LinkExpander,
			"varExpand":    VarExpander,
			"lower":        Lower,
			"dns2path":     Dns2Path,
		})
		t, err = t.Parse(proxy_service)
		checkError(err)

		fname = outdir + strings.Replace(system.Name, "{{ID}}", strconv.FormatInt(system.Id, 10), -1) + "-proxy.service"
		f, err = os.Create(fname)
		checkError(err)
		defer f.Close()
		service_files = append(service_files, fname)

		err = t.Execute(io.Writer(f), system)
		checkError(err)
	}

	return service_files
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func (q *SystemdService) FromJSON(file string) error {

	//Reading JSON file
	J, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var data = &q
	//Umarshalling JSON into struct
	return json.Unmarshal(J, data)
}

func (s *SystemdService) ToJSON(fname string) error {

	ff, err := os.Create(fname)
	if err != nil {
		return err
	}
	t, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	n, err := io.WriteString(ff, string(t))
	if err != nil {
		fmt.Println(n)
		return err
	}

	ff.Close()

	return nil
}
