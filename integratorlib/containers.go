package integratorlib

import (
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"net/http"
	"strconv"
)

//Containers
func (intg IntegratorStruct) StatusContainerHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		cont_all, err := intg.Client.GetContainers(true)
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

func (intg IntegratorStruct) StartContainerHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.StartContainer(id, nil)
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

func (intg IntegratorStruct) RunContainerHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.StartContainer(id, nil)
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

func (intg IntegratorStruct) StopContainerHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}

		err := intg.Client.Client.StopContainer(id, 20)
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

func (intg IntegratorStruct) DelContainerHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		status := map[string]string{"status": "0"}
		err := intg.Client.Client.RemoveContainer(docker.RemoveContainerOptions{ID: id})
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

func (intg IntegratorStruct) CleanContainersHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		containers := intg.Client.CleanContainers()
		intg.Client.RemoveContainers(containers)

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
