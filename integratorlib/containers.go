package integratorlib

import (
	"encoding/json"
	"errors"
	"github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"strconv"
)

//Containers
func (intg IntegratorStruct) StatusContainerHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		cont_all, err := intg.Client.GetContainers(true)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}

		result, err := json.Marshal(cont_all)
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

func (intg IntegratorStruct) StartContainerHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.StartContainer(id, nil)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

func (intg IntegratorStruct) RunContainerHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.StartContainer(id, nil)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

func (intg IntegratorStruct) StopContainerHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		status := map[string]string{"status": "0"}

		err := intg.Client.Client.StopContainer(id, 20)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

func (intg IntegratorStruct) DelContainerHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

func (intg IntegratorStruct) CleanContainersHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		containers := intg.Client.CleanContainers()
		intg.Client.RemoveContainers(containers)

		status := map[string]int{"status": 0}
		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}
