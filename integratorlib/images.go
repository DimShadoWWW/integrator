package integratorlib

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"strconv"
)

//Images
func (intg IntegratorStruct) StatusImageHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		cont_all, err := intg.Client.GetImages()
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

func (intg IntegratorStruct) BuildImageHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("name")

		status := map[string]string{}
		id, err := intg.Client.BuildImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		} else {
			status = map[string]string{"status": "0", "id": id}
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

func (intg IntegratorStruct) DelImageHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.RemoveImage(id)
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

func (intg IntegratorStruct) CleanImagesHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		images := intg.Client.CleanImages()
		intg.Client.RemoveImages(images)

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
