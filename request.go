package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
)

type Content struct {
	ServerId string `json:"server_id"`
}

type Config struct {
	AllowIps []string `json:"allow_ips"`
	Servers  []Server
}

type Server struct {
	ServerId string  `json:"server_id"`
	Proxy    url.URL `json:"proxy"`
}

func logError(err error) {
	log.Printf("[%+v] ERR: %+v", session, err)
}

func logRequest(req *http.Request) {
	log.Printf("[%+v] REQ: %+v -> %+v%+v %+v", session, req.RemoteAddr, req.Host, req.URL, req.Method)
}

func logRedirect(server_id string, url *url.URL) {
	log.Printf("[%+v] SUC: %+v -> %+v\n", session, server_id, url)
}

func is_allow_ip(remoteIp string) bool {
	ip := strings.Split(remoteIp, ":")[0]

	allow_ips := g_cfg.AllowIps
	for i := 0; i < len(allow_ips); i++ {
		if allow_ips[i] == ip {
			return true
		}
	}
	return false
}

func getMatchUrl(server_id string) (error, url.URL) {
	servers := g_cfg.Servers
	for i := 0; i < len(servers); i++ {
		url := servers[i].Proxy
		if server_id == servers[i].ServerId {
			return nil, url
		}
	}
	return errors.New(fmt.Sprintf("%+v not match url", server_id)), url.URL{}
}

func getServerIdFromRequest(req *http.Request) (error, string) {

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err, ""
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// decode body
	var content Content
	err = json.Unmarshal(body, &content)
	if err != nil {
		return err, ""
	}

	if content.ServerId == "" {
		return errors.New("server_id is empty"), ""
	}

	return nil, content.ServerId
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	atomic.AddInt64(&session, 1)
	logRequest(req)

	if !is_allow_ip(req.RemoteAddr) {
		logError(errors.New(fmt.Sprintf("not allow ip %+v", req.RemoteAddr)))
		res.WriteHeader(403)
		return
	}

	err, server_id := getServerIdFromRequest(req)
	if err != nil {
		logError(err)
		res.WriteHeader(403)
		return
	}

	err, url := getMatchUrl(server_id)
	if err != nil {
		logError(err)
		res.WriteHeader(403)
		return
	}

	// proxy
	proxy := httputil.NewSingleHostReverseProxy(&url)
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Host = url.Host

	logRedirect(server_id, req.URL)

	proxy.ServeHTTP(res, req)
}
