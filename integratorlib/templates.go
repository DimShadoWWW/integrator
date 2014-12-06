package integratorlib

import (
	"encoding/json"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Containers
func (intg IntegratorStruct) ListTemplateHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		fileInfo, err := os.Stat(intg.Basedir)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}
		if fileInfo.IsDir() {
			var files []string
			files_in_dir, _ := ioutil.ReadDir(intg.Basedir)
			for _, f := range files_in_dir {
				if f.Mode().IsRegular() && filepath.Ext(f.Name()) == ".json" {
					files = append(files, strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())))
				}
			}

			result, err := json.Marshal(files)
			if err != nil {
				color.Errorf("@bERROR: "+color.ResetCode, err)
				intg.ReturnsEmpty(c, r)
				return
			}
			c.Header().Set("Content-Length", strconv.Itoa(len(result)))
			c.Header().Set("Content-Type", "application/json")
			io.WriteString(c, string(result))
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//Containers
func (intg IntegratorStruct) ReadTemplateHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]

		filename := intg.Basedir + "/" + id + ".json"

		fileInfo, err := os.Stat(filename)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			content := "{}"
			c.Header().Set("Content-Length", strconv.Itoa(len(content)))
			c.Header().Set("Content-Type", "application/json")
			io.WriteString(c, string(content))
			return
		}
		if fileInfo.Mode().IsRegular() {
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				color.Errorf("@bERROR: "+color.ResetCode, err)
				intg.ReturnsEmpty(c, r)
				return
			} else {
				c.Header().Set("Content-Length", strconv.Itoa(len(content)))
				c.Header().Set("Content-Type", "application/json")
				io.WriteString(c, string(content))
			}
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

//Containers
func (intg IntegratorStruct) SaveTemplateHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		color.Println("@r SAVING" + color.ResetCode)
		vars := mux.Vars(r)
		id := vars["id"]

		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			color.Errorf("@bERROR: Failed to read body on @rSaveTemplateHandler "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}
		var services fleet.SystemdServiceList
		err = json.Unmarshal(content, &services)
		if err != nil {
			color.Errorf("@bERROR: Failed to unmarshal SystemdServiceList on @rSaveTemplateHandler "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}

		services_str, err := json.MarshalIndent(services, "", "  ")
		if err != nil {
			color.Println("jerr:", err.Error())
		}
		filename := intg.Basedir + "/" + id + ".json"
		color.Println("@r: writing " + filename + color.ResetCode)
		f, err := os.Create(filename)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}
		_, err = io.WriteString(f, string(services_str))
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			intg.ReturnsEmpty(c, r)
			return
		}
		f.Close()

		c.Header().Set("Content-Length", strconv.Itoa(2))
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, "OK")

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}

func (intg IntegratorStruct) RunTemplateHandler(c http.ResponseWriter, r *http.Request) {
	if intg.Checkaccess(r.Header.Get("API-Access")) {
		vars := mux.Vars(r)
		id := vars["id"]
		serviceid := vars["serviceid"]
		filename := intg.Basedir + "/" + id + ".json"

		fileInfo, err := os.Stat(filename)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			content := "{}"
			c.Header().Set("Content-Length", strconv.Itoa(len(content)))
			c.Header().Set("Content-Type", "application/json")
			io.WriteString(c, string(content))
			return
		}
		if fileInfo.Mode().IsRegular() {
			color.Println("@bStarting: "+color.ResetCode, id)
			if filename != "" {

				rand.Seed(time.Now().UnixNano())
				var id int64
				var err error

				myServices := fleet.SystemdServiceList{}

				file, _ := os.Open(filename)

				err = myServices.FromJSON(file)
				if err != nil {
					panic(err)
				}
				if myServices.Instances == 0 {
					myServices.Instances = 1
				}

				for inst := 0; inst < myServices.Instances; inst++ {
					color.Println("Instance", inst)

					if serviceid != "" {
						id, err = strconv.ParseInt(serviceid, 10, 64)
						if err != nil {
							color.Println("Fatal error ", err.Error())
							for i := 0; i < 10; i++ {
								id = rand.Int63() + 1
							}
						}
					} else {
						for i := 0; i < 10; i++ {
							id = rand.Int63() + 1
						}
					}

					for _, serv := range myServices.Services {
						serv.Id = id

						service_files := fleet.CreateSystemdFiles(serv, "./")

						color.Println("DEPLOY")
						for _, s := range service_files {
							err = fleet.Deploy(s, "")
							if err != nil {
								color.Println(err)
							} else {
								os.Remove("service_files")
							}
						}

					}
				}
				content := "OK"
				c.Header().Set("Content-Length", strconv.Itoa(len(content)))
				c.Header().Set("Content-Type", "application/json")
				io.WriteString(c, string(content))
			}
		}
		color.Errorf("@bERROR: "+color.ResetCode, err)
		c.Header().Set("Content-Length", err.Error())
		c.Header().Set("Content-Type", "application/json")
		io.WriteString(c, err.Error())

	} else {
		http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
	}
}
