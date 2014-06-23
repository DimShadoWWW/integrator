package fleet

import (
	"encoding/json"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

const (
	service = `[Unit]
Description={{.Description}}
After=docker.service
Requires=docker.service{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
{{range $index, $element := .Deps}}After={{replaceId $element $id}}.service
Requires={{replaceId $element $id}}.service{{end}}

[Service]
TimeoutStartSec=0
ExecStart=/bin/bash -c '/usr/bin/docker start -a {{replaceId .Name .Id|lower}} || /usr/bin/docker run --name {{replaceId .Name .Id|lower}}{{if .Hostname}} -h {{.Hostname|lower}}.{{.Domain|lower}} {{end}}{{if .Privileged}} --privileged {{end}}{{if.Volumes}}{{range .Volumes}}{{.|volumeExpand}}{{end}}{{end}} {{if .Ports}}{{range .Ports}}{{.|portExpand}}{{end}}{{end}} {{if .Links}}{{range $index, $element := .Links}}{{linkExpand $element $id}}{{end}}{{end}} {{.ImageName}} {{.Command}}'
ExecStop=/bin/bash -c '/usr/bin/docker stop {{replaceId .Name .Id|lower}};/usr/bin/docker rm {{replaceId .Name .Id|lower}}'


{{if .Conflicts}}[X-Fleet]
{{range $index, $element := .Conflicts}}
X-Conflicts={{replaceId $element $id}}.service{{end}}{{end}}
`

	dns_service = `[Unit]
Description={{.Description}} DNS service
BindsTo={{replaceId .Name .Id}}.service
After={{replaceId .Name .Id}}.service
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
TimeoutStartSec=0
ExecStart=/home/core/dnsctl --hostname={{.Hostname|lower}} --domain={{.Domain|lower}} --region={{.Region|lower}} --id={{.Id}} --port={{.HttpPort}} --priority={{.Priority}} add
ExecStartPost=/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}} '{ \"Host\": \"%H\", \"Port\": {{$httpport}}, \"Priority\": \"{{$priority}}\" }'
ExecStop=/home/core/dnsctl --hostname={{.Hostname|lower}} --domain={{.Domain|lower}} --region={{.Region|lower}} --id={{.Id}} --port={{.HttpPort}} --priority={{.Priority}} del
ExecStopPost=/usr/bin/etcdctl rm /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}}

[X-Fleet]
X-ConditionMachineOf={{replaceId .Name .Id}}.service
`
	proxy_service = `[Unit]
Description={{.Description}} Proxy service
BindsTo={{replaceId .Name .Id}}.service
After={{replaceId .Name .Id}}.service
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
TimeoutStartSec=0
ExecStart=/home/core/proxyctl --id={{$id}} --hostname={{$hostname|lower}} --domain={{$domain|lower}} --region={{$region|lower}} --port={{$httpport}} add
ExecStop=/home/core/proxyctl --id={{$id}} --hostname={{$hostname|lower}} --domain={{$domain|lower}} --region={{$region|lower}} --port={{$httpport}} del

[X-Fleet]
X-ConditionMachineOf={{replaceId .Name .Id}}.service
`
	service_discovery = `
[Unit]
Description={{.Description}} presence service
BindsTo={{replaceId .Name .Id}}.service
{{$id := .Id}}{{$name := .Name}}{{$hostname := .Hostname}}{{$domain := .Domain}}{{$region := .Region}}{{$priority := .Priority}}{{$httpport := .HttpPort}}
[Service]
ExecStart=/bin/sh -c "while true; do {{if .HttpPort}}/usr/bin/etcdctl set /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}} '{ \"Host\": \"%H\", \"Port\": {{$httpport}}, \"Priority\": \"{{$priority}}\" }' --ttl 60;{{else}}{{if .Ports}}{{range .Ports}}/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name|lower}}/{{.HostPort}} '{ \"Host\": \"%H\", \"Port\": {{.HostPort}}, \"Priority\": \"{{$priority}}\" }' --ttl 60;{{end}}{{else}}/usr/bin/etcdctl set /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name|lower}} '{ \"Host\": \"%H\", \"Priority\": \"{{$priority}}\" }' --ttl 60;{{end}}{{end}}sleep 45;done"
ExecStop=/bin/sh -c "{{if .HttpPort}}/usr/bin/etcdctl rm /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}};{{else}}{{if .Ports}}{{range .Ports}}/usr/bin/etcdctl rm /services/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$name}}-{{$id}}/{{.HostPort}};{{end}}/usr/bin/etcdctl rmdir /services/{{$hostname|lower}}.{{$domain|lower}}/{{$name}}-{{$id}}{{else}}/usr/bin/etcdctl rm /skydns/{{ printf "%s.%s.%s" $region $hostname $domain |dns2path}}/{{$id}}{{end}}{{end}}"

[X-Fleet]
X-ConditionMachineOf={{replaceId .Name .Id}}.service
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

type SystemdServiceList struct {
	Services []SystemdService
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

func LinkExpander(l Link, id int64) string {
	line := ""

	if l.Name != "" && l.Alias != "" {
		line = line + fmt.Sprintf(" --link %s:%s", strings.Replace(l.Name, "_ID", "-"+strconv.FormatInt(id, 10), -1), l.Alias)
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

func ReplaceId(str string, id int64) string {
	return strings.ToLower(strings.Replace(str, "_ID", "-"+strconv.FormatInt(id, 10), -1))
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

	color.Println("@bCreating systemd unit file for service: "+color.ResetCode, system.Name)
	t := template.New("Systemd service")
	// add our function
	t = t.Funcs(template.FuncMap{
		"volumeExpand": VolumeExpander,
		"portExpand":   PortExpander,
		"linkExpand":   LinkExpander,
		"varExpand":    VarExpander,
		"lower":        Lower,
		"replaceId":    ReplaceId,
		"dns2path":     Dns2Path,
	})
	t, err := t.Parse(service)
	checkError(err)

	fname, err := generateUnitName(system, "")
	checkError(err)
	f, err := os.Create(outdir + fname)
	checkError(err)
	defer f.Close()
	err = t.Execute(io.Writer(f), system)
	checkError(err)
	service_files := []string{outdir + fname}

	// dns
	color.Println("@bCreating systemd unit file for dns control of service: "+color.ResetCode, system.Name)
	t = template.New("Systemd dns service")
	t = t.Funcs(template.FuncMap{
		"volumeExpand": VolumeExpander,
		"portExpand":   PortExpander,
		"linkExpand":   LinkExpander,
		"varExpand":    VarExpander,
		"lower":        Lower,
		"replaceId":    ReplaceId,
		"dns2path":     Dns2Path,
	})
	t, err = t.Parse(dns_service)
	checkError(err)

	fname, err = generateUnitName(system, "dns")
	checkError(err)
	f, err = os.Create(outdir + fname)
	checkError(err)
	defer f.Close()
	service_files = append(service_files, outdir+fname)

	err = t.Execute(io.Writer(f), system)
	checkError(err)
	fmt.Println(system.Name)

	// enable proxy only if httport > 0
	if system.HttpPort > 0 {
		color.Println("@bCreating systemd unit file for inverse proxy of service: "+color.ResetCode, system.Name)
		t = template.New("Systemd proxy service")
		t = t.Funcs(template.FuncMap{
			"volumeExpand": VolumeExpander,
			"portExpand":   PortExpander,
			"linkExpand":   LinkExpander,
			"varExpand":    VarExpander,
			"lower":        Lower,
			"replaceId":    ReplaceId,
			"dns2path":     Dns2Path,
		})
		t, err = t.Parse(proxy_service)
		checkError(err)

		fname, err = generateUnitName(system, "proxy")
		checkError(err)
		f, err = os.Create(outdir + fname)
		checkError(err)
		defer f.Close()
		service_files = append(service_files, outdir+fname)

		err = t.Execute(io.Writer(f), system)
		checkError(err)
	}

	return service_files
}

func checkError(err error) {
	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func (q *SystemdServiceList) FromJSON(file io.Reader) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	var data = &q
	//Umarshalling JSON into struct
	return json.NewDecoder(file).Decode(data) //.Unmarshal(file.Fd(), data)
}

func generateUnitName(system SystemdService, suffix string) (fname string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			fname = ""
			err = r.(error)
		}
	}()

	if suffix != "" {
		suffix = "-" + suffix
	}

	matched, err := regexp.MatchString("_ID", system.Name)
	if matched {
		fname = strings.Replace(system.Name, "_ID", suffix+"-"+strconv.FormatInt(system.Id, 10), -1) + ".service"
	} else {
		fname = system.Name + suffix + ".service"
	}

	return fname, nil
}

func (s *SystemdServiceList) ToJSON(fname string) error {

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
