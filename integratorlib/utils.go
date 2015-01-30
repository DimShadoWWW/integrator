package integratorlib

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net/http"
	"strconv"
	"strings"
)

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

func (intg IntegratorStruct) Checkaccess(a string) bool {
	if a == intg.APIKey {
		return true
	}
	return false
}

// 403
func (intg IntegratorStruct) ReturnsEmpty(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		c.Writer.Header().Set("Content-Length", strconv.Itoa(4))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, "{}")
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

//General functions
func (intg IntegratorStruct) CleanHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		containers := intg.Client.CleanContainers()
		intg.Client.RemoveContainers(containers)
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
