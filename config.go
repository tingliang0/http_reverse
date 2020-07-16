package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func load_cfg() {
	if g_json_file == "" {
		return
	}
	jsonFile, err := os.Open(g_json_file)
	defer jsonFile.Close()
	if err != nil {
		logError(err)
		return
	}
	// parse json
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logError(err)
		return
	}

	var cfg Config
	err = json.Unmarshal([]byte(byteValue), &cfg)
	if err != nil {
		logError(err)
	}

	g_cfg = cfg
	log.Printf("reload cfg: %+v", g_cfg)
}
