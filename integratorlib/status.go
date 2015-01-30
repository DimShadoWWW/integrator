package integratorlib

import (
	"encoding/json"
	"errors"
	"github.com/DimShadoWWW/integrator/dockerlib"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net"
	"strconv"
	"strings"
)

func (intg IntegratorStruct) StatusHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		a := getIpAddress(c.Request)
		rip, err := net.ResolveTCPAddr("tcp", a+":0")
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}

		color.Println("@bACCESS from: "+color.ResetCode, rip.IP)

		cont_all, err := intg.Client.GetContainers(true)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
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
			c.JSON(500, gin.H{})
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
			c.JSON(500, gin.H{})
			return
		}

		var r1 interface{}
		err = json.Unmarshal(result, &r1)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}
