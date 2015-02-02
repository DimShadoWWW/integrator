package integratorlib

import (
	"encoding/json"
	"errors"
	"github.com/DimShadoWWW/integrator/fleet"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Containers
func (intg IntegratorStruct) ListTemplateHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		fileInfo, err := os.Stat(intg.Basedir)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
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
				c.JSON(500, gin.H{})
				return
			}
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
			c.Writer.Header().Set("Content-Type", "application/json")
			io.WriteString(c.Writer, string(result))
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

//Containers
func (intg IntegratorStruct) ReadTemplateHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")

		filename := intg.Basedir + "/" + id + ".json"

		fileInfo, err := os.Stat(filename)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			content := "{}"
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
			c.Writer.Header().Set("Content-Type", "application/json")
			io.WriteString(c.Writer, string(content))
			return
		}
		if fileInfo.Mode().IsRegular() {
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				color.Errorf("@bERROR: "+color.ResetCode, err)
				c.JSON(500, gin.H{})
				return
			} else {
				c.Writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
				c.Writer.Header().Set("Content-Type", "application/json")
				io.WriteString(c.Writer, string(content))
			}
		}

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

//Containers
func (intg IntegratorStruct) DeleteTemplateHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")
		status := map[string]string{"status": "0", "error": ""}

		filename := intg.Basedir + "/" + id + ".json"

		fileInfo, err := os.Stat(filename)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		} else {
			if fileInfo.Mode().IsRegular() {
				err = os.Remove(filename)
				if err != nil {
					color.Errorf("@rERROR: "+color.ResetCode, err)
					status = map[string]string{"status": "1", "error": err.Error()}
				}
			}
		}

		result, err := json.Marshal(status)
		if err != nil {
			color.Errorf("@rERROR: "+color.ResetCode, err)
			status = map[string]string{"status": "1", "error": err.Error()}
		} else {
			status = map[string]string{"status": "0", "error": ""}
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, string(result))
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

//Containers
func (intg IntegratorStruct) SaveTemplateHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {
		color.Println("@r SAVING" + color.ResetCode)

		id := c.Params.ByName("id")

		content, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			color.Errorf("@bERROR: Failed to read body on @rSaveTemplateHandler "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		var services fleet.SystemdServiceList
		err = json.Unmarshal(content, &services)
		if err != nil {
			color.Errorf("@bERROR: Failed to unmarshal SystemdServiceList on @rSaveTemplateHandler "+color.ResetCode, err)
			c.JSON(500, gin.H{})
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
			c.JSON(500, gin.H{})
			return
		}
		_, err = io.WriteString(f, string(services_str))
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.JSON(500, gin.H{})
			return
		}
		f.Close()

		c.Writer.Header().Set("Content-Length", strconv.Itoa(2))
		c.Writer.Header().Set("Content-Type", "application/json")
		io.WriteString(c.Writer, "OK")

		//rr, err := json.NewEncoder(c.ResponseWriter).Encode(m)
	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}

func (intg IntegratorStruct) RunTemplateHandler(c *gin.Context) {
	if intg.Checkaccess(c.Request.Header.Get("API-Access")) {

		id := c.Params.ByName("id")
		serviceid := c.Params.ByName("serviceid")
		filename := intg.Basedir + "/" + id + ".json"

		fileInfo, err := os.Stat(filename)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			content := "{}"
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
			c.Writer.Header().Set("Content-Type", "application/json")
			io.WriteString(c.Writer, string(content))
			return
		}
		if fileInfo.Mode().IsRegular() && filename != "" {
			color.Println("@bStarting: "+color.ResetCode, id)

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

			content := "Loaded: ["
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
					for ks, s := range service_files {
						if ks != 0 {
							content = content + ", "
						}
						err = fleet.Deploy(s, "")
						if err != nil {
							color.Println(err)
						} else {
							os.Remove(s)
						}
						content = content + s
					}

				}
			}
			content = content + "]"
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
			c.Writer.Header().Set("Content-Type", "application/json")
			io.WriteString(c.Writer, string(content))

		} else {
			color.Errorf("@bERROR: "+color.ResetCode, err)
			c.Writer.Header().Set("Content-Length", err.Error())
			c.Writer.Header().Set("Content-Type", "application/json")
			io.WriteString(c.Writer, err.Error())
		}

	} else {
		c.Fail(401, errors.New("Unauthorized"))
		color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + c.Request.URL.Path[1:] + " from " + c.Request.RemoteAddr)
	}
}
