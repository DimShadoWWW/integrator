package dockerio

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	APIVERSION        = 1.9
	DEFAULTHTTPHOST   = "127.0.0.1"
	DEFAULTHTTPPORT   = 4243
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
)

var docker DockerAccess

func doGet(url string) string {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("The calculated length is:", len(string(contents)), "for the url:", url)
		return contents
		// fmt.Println(" ", response.StatusCode)
		// hdr := response.Header
		// for key, value := range hdr {
		// 	fmt.Println(" ", key, ":", value)
		// }
	}
}

func getImagesJSON() string {
	return doGet("http://" + DEFAULTHTTPHOST + ":" + DEFAULTHTTPPORT + "/images/json")
}
