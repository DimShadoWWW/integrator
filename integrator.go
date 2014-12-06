package main

import (
	"encoding/json"
	"flag"
	"github.com/DimShadoWWW/integrator/integratorlib"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/wsxiaoys/terminal/color"
	"net/http"
	"os"
)
import log "github.com/cihub/seelog"
import _ "github.com/rakyll/gometry/http"

type Configuration struct {
	Basedir string
	Address string
	Port    string
	ApiKey  string
}

var (

	// path to templates folder
	basedir string

	// address to access docker API
	address string

	// port to listen on
	port string

	client integratorlib.IntegratorStruct

	APIKey string
)

func main() {

	var logConfig = `
 	<seelog type="asyncloop" minlevel="info">
 	<outputs>
 	<file path="integrator.log" formatid="main"/>
 	</outputs>
 	<formats>
 	<format id="main" format="%Date %Time [%Level] %Msg%n"/>
 	</formats>
 	</seelog>`

	logger, err := log.LoggerFromConfigAsBytes([]byte(logConfig))

	if err != nil {
		color.Errorf("@bERROR: "+color.ResetCode, "Error during logger config load: ", err.Error())
	}

	log.ReplaceLogger(logger)

	defer log.Flush()

	defaultBasedir := "/home/core/integrator_units/"
	defaultAddress := "unix:///var/run/docker.sock"
	defaultPort := "8080"
	defaultAPIKey := "CHANGE_ME"

	if _, err := os.Stat("config.json"); err == nil {
		file, _ := os.Open("config.json")
		decoder := json.NewDecoder(file)
		configuration := Configuration{}
		err := decoder.Decode(&configuration)
		if err != nil {
			color.Errorf("@bERROR: "+color.ResetCode, err)
		}
		defaultBasedir = configuration.Basedir
		defaultAddress = configuration.Address
		defaultPort = configuration.Port
		defaultAPIKey = configuration.ApiKey
	}

	// parse command line flags
	flag.StringVar(&basedir, "basedir", defaultBasedir, "fullpath to folder with templates")
	flag.StringVar(&address, "address", defaultAddress, "docker address")
	flag.StringVar(&port, "port", defaultPort, "port to listen on")
	flag.StringVar(&APIKey, "apikey", defaultAPIKey, "ApiKey for access")
	flag.Parse()

	color.Println(basedir)
	color.Println(APIKey)
	color.Println(address)
	color.Println(port)

	client, err = integratorlib.NewIntegrator(address, APIKey, basedir)
	if err != nil {
		panic(err)
	}

	checkHeaderThenServe := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set some header.
			if r.Header.Get("API-Access") == APIKey {
				w.Header().Add("Keep-Alive", "300")
				// Serve with the actual handler.
				h.ServeHTTP(w, r)
			} else {
				// http.Error(c, "403 Forbidden - Access Denied", http.StatusForbidden)
				color.Errorf("@bERROR: " + color.ResetCode + " (403) accessing " + r.URL.Path[1:] + " from " + r.RemoteAddr)
			}
		}
	}

	r := mux.NewRouter()
	// r.Headers("API-Access", APIKey)
	r.HandleFunc("/api/status", client.StatusHandler)
	r.HandleFunc("/api/containers", client.StatusContainerHandler)
	r.HandleFunc("/api/containers/del/{id}", client.DelContainerHandler)
	r.HandleFunc("/api/containers/stop/{id}", client.StopContainerHandler)
	r.HandleFunc("/api/containers/start/{id}", client.StartContainerHandler)
	r.HandleFunc("/api/containers/clean", client.CleanContainersHandler)
	r.HandleFunc("/api/templates/list", client.ListTemplateHandler)
	r.HandleFunc("/api/templates/read/{id}", client.ReadTemplateHandler)
	r.HandleFunc("/api/templates/save/{id}", client.SaveTemplateHandler)
	r.HandleFunc("/api/templates/run/{id}", client.RunTemplateHandler)
	// r.HandleFunc("/api/containers/new", ReadTemplateHandler)
	r.HandleFunc("/api/clean", client.CleanHandler)
	r.HandleFunc("/api/images", client.StatusImageHandler)
	r.HandleFunc("/api/images/build/{name}", client.BuildImageHandler)
	r.HandleFunc("/api/images/del/{id}", client.DelImageHandler)
	r.HandleFunc("/api/images/clean", client.CleanImagesHandler)
	http.Handle("/api/", r)

	http.Handle("/", checkHeaderThenServe(http.FileServer(rice.MustFindBox("public").HTTPBox())))
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Error(err)
	}
}
