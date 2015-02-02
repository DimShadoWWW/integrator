package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/DimShadoWWW/integrator/integratorlib"
	"github.com/GeertJohan/go.rice"
	"github.com/gin-gonic/gin"
	"github.com/wsxiaoys/terminal/color"
	"net/http"
	"os"
	"path"
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

// checkHeaderThenServe := func(h http.Handler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// Set some header.
// 	}
// }

// Implements a basic Basic HTTP Authorization. It takes as argument a map[string]string where
// the key is the user name and the value is the password.
func AuthRequired(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Search user in the slice of allowed credentials
		if c.Request.Header.Get("API-Access") != APIKey {
			// c.Writer.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			c.Fail(401, errors.New("Unauthorized"))
		}
	}
}

func StaticRice(r *gin.Engine, p string, ric *rice.HTTPBox) {
	prefix := "/public/"
	p = path.Join(p, "/*filepath")
	fileServer := http.StripPrefix(prefix, http.FileServer(ric))
	r.GET(p, func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.HEAD(p, func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

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

	// r := mux.NewRouter()
	// // r.Headers("API-Access", APIKey)
	// r.HandleFunc("/api/status", client.StatusHandler)
	// r.HandleFunc("/api/containers", client.StatusContainerHandler)
	// r.HandleFunc("/api/containers/del/{id}", client.DelContainerHandler)
	// r.HandleFunc("/api/containers/stop/{id}", client.StopContainerHandler)
	// r.HandleFunc("/api/containers/start/{id}", client.StartContainerHandler)
	// r.HandleFunc("/api/containers/clean", client.CleanContainersHandler)
	// r.HandleFunc("/api/templates/list", client.ListTemplateHandler)
	// r.HandleFunc("/api/templates/read/{id}", client.ReadTemplateHandler)
	// r.HandleFunc("/api/templates/save/{id}", client.SaveTemplateHandler)
	// r.HandleFunc("/api/templates/run/{id}", client.RunTemplateHandler)
	// // r.HandleFunc("/api/containers/new", ReadTemplateHandler)
	// r.HandleFunc("/api/clean", client.CleanHandler)
	// r.HandleFunc("/api/images", client.StatusImageHandler)
	// r.HandleFunc("/api/images/build/{name}", client.BuildImageHandler)
	// r.HandleFunc("/api/images/del/{id}", client.DelImageHandler)
	// r.HandleFunc("/api/images/clean", client.CleanImagesHandler)
	// http.Handle("/api/", r)
	// http.Handle("/", checkHeaderThenServe(http.FileServer(rice.MustFindBox("public").HTTPBox())))
	// err = http.ListenAndServe(":"+port, nil)
	// if err != nil {
	// 	log.Error(err)
	// }

	// r := gin.Default()

	// // Simple group: v1
	// api := r.Group("/api")

	// {
	// 	api.GET("/status", client.StatusHandler)
	// 	api.GET("/containers/clean", client.CleanContainersHandler)
	// 	api.GET("/containers", client.StatusContainerHandler)
	// 	api.GET("/containers/:id/del", client.DelContainerHandler)
	// 	api.GET("/containers/:id/stop", client.StopContainerHandler)
	// 	api.GET("/containers/:id/start", client.StartContainerHandler)

	// 	api.GET("/templates/list", client.ListTemplateHandler)
	// 	api.GET("/templates/:id", client.client.ReadTemplateHandler)
	// 	api.GET("/templates/:id/save", client.SaveTemplateHandler)
	// 	api.GET("/templates/:id/run", client.RunTemplateHandler)
	// 	api.GET("/clean", client.CleanHandler)
	// 	api.GET("/images", client.StatusImageHandler)
	// 	// api.GET("/images/:name/build", client.BuildImageHandler)
	// 	api.GET("/images/del/:id", client.DelImageHandler)
	// 	api.GET("/images/clean", client.CleanImagesHandler)

	// 	// api.GET("/", client.)
	// 	// api.GET("/", client.)

	// 	// api.POST("/submit", submitEndpoint)
	// 	// api.POST("/read", readEndpoint)
	// }

	// Creates a router without any middleware by default
	r := gin.New()

	// Global middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Per route middlewares, you can add as many as you desire.
	// r.GET("/benchmark", MyBenchLogger(), benchEndpoint)

	// Authorization group
	// authorized := r.Group("/", AuthRequired())
	// exactly the same than:
	authorized := r.Group("/")
	// per group middlewares! in this case we use the custom created
	// AuthRequired() middleware just in the "authorized" group.
	authorized.Use(AuthRequired(APIKey))
	{

		r.GET("/", func(c *gin.Context) {
			c.Redirect(301, "/public/index.html")
		})
		// Simple group: v1
		api := authorized.Group("/api")
		{
			api.GET("/status", client.StatusHandler)
			api.GET("/containers/clean", client.CleanContainersHandler)
			api.GET("/containers/list", client.StatusContainerHandler)
			api.GET("/containers/del/:id", client.DelContainerHandler)
			api.GET("/containers/stop/:id", client.StopContainerHandler)
			api.GET("/containers/start/:id", client.StartContainerHandler)

			api.GET("/templates/list", client.ListTemplateHandler)
			api.GET("/templates/read/:id", client.ReadTemplateHandler)
			api.POST("/templates/save/:id", client.SaveTemplateHandler)
			api.GET("/templates/run/:id", client.RunTemplateHandler)
			api.GET("/templates/del/:id", client.DeleteTemplateHandler)
			api.GET("/clean", client.CleanHandler)
			api.GET("/images", client.StatusImageHandler)
			api.POST("/images/pull", client.PullImageHandler)
			api.GET("/images/del/:id", client.DelImageHandler)
			api.GET("/images/clean", client.CleanImagesHandler)

			// api.GET("/", client.)
			// api.POST("/submit", submitEndpoint)
		}
	}

	// http.Handle("/", checkHeaderThenServe(http.FileServer(rice.MustFindBox("public").HTTPBox())))
	StaticRice(r, "/public/", rice.MustFindBox("public").HTTPBox())

	// Listen and server on 0.0.0.0:8080
	r.Run(":" + port)
}
