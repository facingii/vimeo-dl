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

const (
	VimeoPlayerBaseUrl    = "https://player.vimeo.com/video/"
	DownLoadInProgress    = "Downloading video %s...please wait\n"
	UsageMessage          = "You must specify video URL or VIDEO ID\n\nUsage:\n\tvimeo [VIDEO_URL][VIDEO_ID]\n\n"
	DownloadFinishMessage = "%d bytes downloaded\n"
	HttpString            = "http"
	ConfigString          = "config"
	SlashString           = "/"
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
	if len (os.Args) < 2 {
		fmt.Printf (UsageMessage)
		os.Exit (0)
	}

	var video string = os.Args [1]

	if !strings.HasPrefix (video, HttpString) {
		video = VimeoPlayerBaseUrl + video
	}

	if !strings.HasSuffix (video, SlashString) {
		video += SlashString
	}

	fmt.Printf (DownLoadInProgress, video [:len (video) - 1])

	video += ConfigString

	var uri string = pickBestQuality (getVideoConfig (video))
	var filename string = buildFileName (uri)
	var file *os.File = createFile (filename)
	var client *http.Client = createHttpClient ()

	var bytesDownloaded int64 = downloadFile (file, client, uri)
	fmt.Printf (DownloadFinishMessage, bytesDownloaded)
}

// get config settings of the media, needed to figure out
// best quality available
func getVideoConfig (uri string) *Config {
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
	segments := strings.Split (path, SlashString)
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
func downloadFile (file *os.File, client *http.Client, uri string) int64 {
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

	return size
}
