package integratorlib

import (
	"encoding/json"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func (intg IntegratorStruct) StatusHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		a := getIpAddress(r)
		rip, err := net.ResolveTCPAddr("tcp", a+":0")
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}

		color.Println("@bACCESS from: "+color.ResetCode, rip.IP)

		cont_all, err := intg.Client.GetContainers(true)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
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

		images, err := intg.Client.GetImages()
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
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
			intg.ReturnsEmpty(c, r)
			return
		}

		var r1 interface{}
		err = json.Unmarshal(result, &r1)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
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
