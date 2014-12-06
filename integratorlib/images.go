package integratorlib

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net/http"
	"strconv"
)

//Images
func (intg IntegratorStruct) StatusImageHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		cont_all, err := intg.Client.GetImages()
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}

		result, err := json.Marshal(cont_all)
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

func (intg IntegratorStruct) BuildImageHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["name"]

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

func (intg IntegratorStruct) DelImageHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.RemoveImage(id)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		}

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

func (intg IntegratorStruct) CleanImagesHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
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
