package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Progressive struct {
	Url string `json:"url"`
	Height int32 `json:"height"`
}

type File struct {
	Progressives []Progressive `json:"progressive"`
}

type Request struct {
	Files File
}

type Config struct {
	Req Request `json:"request"`
}

// usage:
//	./vimeo [VIDEO_URL][VIDEO_ID]
//
func main () {
	var video string = os.Args [1]

	if !strings.HasPrefix (video, "http") {
		video = "https://player.vimeo.com/video/" + video
	}

	if !strings.HasSuffix (video, "/") {
		video += "/"
	}

	fmt.Printf ("Downloading video %s ...please wait\n", video)

	video += "config"

	var uri string = pickBestQuality (getVideoConfig (video))
	var filename string = buildFileName (uri)
	var file *os.File = createFile (filename)
	var client *http.Client = createHttpClient ()

	downloadFile (file, client, uri)

}

// get config settings of the media, needed to figure out
// best quality available
func getVideoConfig (uri string) *Config {
	//uri := "https://player.vimeo.com/video/354915591/config"
	//url := "https://player.vimeo.com/video/354915352/config"

	resp, err := http.Get (uri)
	if err != nil {
		panic (err)
	}

	defer resp.Body.Close ()

	html, err := ioutil.ReadAll (resp.Body)
	if err != nil {
		panic (err)
	}

	var config Config
	err = json.Unmarshal (html, &config)
	if err != nil {
		panic (err)
	}

	return &config
}

// select best quality video based on file's height,
// once it is found, return the url corresponding
//
func pickBestQuality (config *Config) string {
	var foo int32 = -1
	var uri string

	for _, p := range config.Req.Files.Progressives {
		if p.Height > foo {
			foo = p.Height
			uri = p.Url
		}
	}

	return uri
}

// extract filename from url
func buildFileName (uri string) string {
	fileUrl, err := url.Parse (uri)
	if err != nil {
		panic (err)
	}

	path := fileUrl.Path
	segments := strings.Split (path, "/")
	filename := segments [len (segments) - 1]

	return filename
}

// creates file in OS filesystem
func createFile (filename string) *os.File {
	file, err := os.Create (filename)
	if err != nil {
		panic (err)
	}

	return file
}

// creates http client in order to download the video
func createHttpClient () *http.Client {
	client := http.Client {
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.URL.Opaque = req.URL.Path
			return nil
		},
	}

	return &client
}

// do the download
func downloadFile (file *os.File, client *http.Client, uri string) {
	resp, err := client.Get (uri)
	if err != nil {
		panic (err)
	}
	defer resp.Body.Close ()

	size, err := io.Copy (file, resp.Body)
	if err != nil {
		panic (err)
	}
	defer file.Close ()

	fmt.Printf ("%d bytes downloaded\n", size)
}
