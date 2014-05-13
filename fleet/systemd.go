package fleet

import (
	"fmt"
	"io"
	"os"
	"text/template"
	// "reflect"
)

const (
	service = `
[Unit]
Description={{.Description}}
After=docker.service
Requires=docker.service
{{range .Deps}}
After={{.}}.service
Requires={{.}}.service
{{end}}

[Service]
ExecStart=/usr/bin/docker run --name {{.Name}}-{{.Id}} {{if .Hostname}}-h {{.Hostname}}.{{.Domain}}{{end}} {{if .Privileged}}--privileged{{end}} {{if.Volumes}}{{range .Volumes}}{{.|volumeExpand}}{{end}}{{end}} {{if .Ports}}{{range .Ports}}{{.|portExpand}}{{end}}{{end}} {{if .Links}}{{range .Links}}{{.|linkExpand}}{{end}}{{end}} {{.ImageName}} {{.Command}}
ExecStop=/usr/bin/docker stop {{.Name}}

{{if .Conflicts}}
[X-Fleet]
{{range .Conflicts}}
X-Conflicts={{.}}.service
{{end}}
{{end}}
`

	service_discovery = `
[Unit]
Description={{.Description}} presence service
BindsTo={{.Name}}.service

{{$id := .Id}}
{{$name := .Name}}
{{$hostname := .Hostname}}
{{$domain := .Domain}}
[Service]
ExecStart=/bin/sh -c "while true; do {{if .Ports}}{{range .Ports}}etcdctl set /services/{{$hostname}}.{{$domain}}/{{$name}}-{{$id}} '{ \"Host\": \"%H\", \"Port\": {{.HostPort}}, \"Priority\": \"{{.Priority}}\" }' --ttl 60;{{end}}{{end}}sleep 45;done"
ExecStop=/usr/bin/etcdctl rm /services/{{$hostname}}.{{$domain}}/{{$name}}-{{$id}}

[X-Fleet]
X-ConditionMachineOf={{.Name}}.service
`
)

type Volume struct {
	LocalDir     string
	ContainerDir string
}

type Port struct {
	HostPort      string
	ContainerPort string
	Protocol      string
}

type Link struct {
	Name  string
	Alias string
}
type Env struct {
	Name  string
	Value string
}

type SystemdService struct {
	Id          string
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
	// IncludeFleet bool
}

func VolumeExpander(args ...interface{}) string {
	line := ""
	for _, i := range args {
		j, _ := i.(Volume)

		// ok := false
		// var i Volume
		// if len(args) == 1 {
		// 	i, ok = args[0].(Volume)
		// }
		// if !ok {
		// 	fmt.Println("failed")
		// 	fmt.Sprint(args...)
		// }

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

func CreateSystemdFiles(system SystemdService, outdir string) {

	t := template.New("Systemd service")
	// add our function
	t = t.Funcs(template.FuncMap{"volumeExpand": VolumeExpander, "portExpand": PortExpander, "linkExpand": LinkExpander, "varExpand": VarExpander})

	t, err := t.Parse(service)
	checkError(err)

	f, err := os.Create(outdir + system.Name + "-" + system.Id + ".service")
	checkError(err)
	defer f.Close()

	fmt.Printf("%s\n", system)
	err = t.Execute(io.Writer(f), system)
	checkError(err)
	fmt.Printf("%s\n", system.Volumes[0])

	t = template.New("Systemd discovery service")
	t, err = t.Parse(service_discovery)
	checkError(err)

	f, err = os.Create(outdir + system.Name + "-" + system.Id + "-discovery.service")
	checkError(err)
	defer f.Close()

	err = t.Execute(io.Writer(f), system)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
