package integratorlib

import (
	"encoding/json"
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
func (intg IntegratorStruct) ReturnsEmpty(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		c.Header().Set("Content-Length", strconv.Itoa(4))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, "{}")
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//General functions
func (intg IntegratorStruct) CleanHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		containers := intg.Client.CleanContainers()
		intg.Client.RemoveContainers(containers)
		images := intg.Client.CleanImages()
		intg.Client.RemoveImages(images)

		status := map[string]int{"status": 0}
		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}
		c.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, string(result))
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}
