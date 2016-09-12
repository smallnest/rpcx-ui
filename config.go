package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func init() {
	loadConfig()
}

// Config parameters
var serverConfig = Configuration{}
var reg Registry

func loadConfig() {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	//fmt.Printf("Loaded Config: \n%s\n", string(file))
	json.Unmarshal(file, &serverConfig)
	fmt.Println("succeeded to read the config")

	if serverConfig.RegistryType == "zookeeper" {
		reg = &ZooKeeperRegistry{}
	} else if serverConfig.RegistryType == "etcd" {
		reg = &EtcdRegistry{}
	}

	if reg != nil {
		reg.initRegistry()
	}

}

// Configuration is configuration strcut refects the config.json
type Configuration struct {
	RegistryType   string `json:"registry_type"`
	RegistryURL    string `json:"registry_url"`
	ServiceBaseURL string `json:"service_base_url"`
	Host           string `json:"host,omitempty"`
	Port           int    `json:"port,omitempty"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
}
