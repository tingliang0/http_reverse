package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	ServerId string  `json:"server_id"`
	Proxy    url.URL `json:"proxy"`
}

type Content struct {
	ServerId string `json:"server_id"`
}

var (
	cfg         string
	g_servers   []Server
	g_proxy     *httputil.ReverseProxy
	g_json_file string
	g_sigs      chan os.Signal
)

func get_match_url(server_id string) (error, url.URL) {
	for i := 0; i < len(g_servers); i++ {
		url := g_servers[i].Proxy
		if server_id == g_servers[i].ServerId {
			return nil, url
		}
	}
	return errors.New("url not match"), url.URL{}
}

func create_reverse_proxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)
		var content Content
		err := json.Unmarshal(body, &content)
		if err != nil {
			fmt.Println(err)
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			req.URL.Scheme = "http"
			req.URL.Host = req.URL.Host
			req.URL.Path = req.URL.Path
		} else {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			err, url := get_match_url(content.ServerId)
			if err != nil {
				log.Println("not support server_id: %+v\n", content.ServerId)
				req.URL.Scheme = "http"
				req.URL.Host = req.URL.Host
				req.URL.Path = req.URL.Path
			} else {
				req.URL.Scheme = url.Scheme
				req.URL.Host = url.Host
				req.URL.Path = req.URL.Path
				log.Printf("%+v -> %+v", content.ServerId, req.URL)
			}
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

func init_servers(filename string) []url.URL {
	// open json file
	fd, err := os.Open(cfg)
	defer fd.Close()
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()
	if err != nil {
		panic(err.Error())
	}

	// parse json
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var servers []Server
	err = json.Unmarshal([]byte(byteValue), &servers)
	if err != nil {
		panic(err.Error())

	}
	g_servers = servers

	var urls []url.URL = make([]url.URL, len(g_servers))

	for i := 0; i < len(g_servers); i++ {
		url := g_servers[i].Proxy
		urls[i] = url
	}

	return urls
}

func reload_servers() []url.URL {
	before := len(g_servers)
	urls := init_servers(g_json_file)
	after := len(g_servers)
	log.Printf("reload servers, before %+v, after %+v\n", before, after)
	return urls
}

func init_signals() {
	g_sigs := make(chan os.Signal, 1)
	signal.Notify(g_sigs, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-g_sigs
			log.Printf("receive sig: %+v\n", sig)
			reload_servers()
		}
	}()
}

func main() {
	listen := flag.String("listen", ":9000", "http server listen address")
	filename := flag.String("f", "servers.json", "users config file")
	flag.Parse()

	g_json_file = *filename
	init_servers(g_json_file)
	init_signals()

	g_proxy := create_reverse_proxy()
	log.Fatal(http.ListenAndServe(*listen, g_proxy))
}
