package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

var (
	g_servers   []Server
	g_proxy     *httputil.ReverseProxy
	g_json_file string
	g_sigs      chan os.Signal
	session     int64
	g_cfg       Config
)

func main() {
	listen := flag.String("listen", ":9000", "http server listen address")
	filename := flag.String("f", "servers.json", "users config file")
	flag.Parse()

	g_cfg = Config{}
	g_json_file = *filename

	load_cfg()
	init_signals()

	http.HandleFunc("/", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
