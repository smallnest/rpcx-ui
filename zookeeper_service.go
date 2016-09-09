package main

import (
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type ZooKeeperRegistry struct {
}

func (r *ZooKeeperRegistry) fetchServices() []*Service {
	c, _, err := zk.Connect(strings.Split(serverConfig.RegistryURL, ","), 10*time.Second)
	if err != nil {
		panic(err)
	}

	exist, _, _ := c.Exists(serverConfig.ServiceBaseURL)
	if !exist {
		return []*Service{}
	}

	var services []*Service
	zkServices, _, _ := c.Children(serverConfig.ServiceBaseURL)
	for _, s := range zkServices {
		endpoints, _, _ := c.Children(serverConfig.ServiceBaseURL + "/" + s)
		for _, ep := range endpoints {
			bytes, _, _ := c.Get(serverConfig.ServiceBaseURL + "/" + s + "/" + ep)
			metadata := string(bytes)
			v, err := url.ParseQuery(metadata)
			state := "n/a"
			if err == nil && v.Get("state") != "" {
				state = v.Get("state")
			}

			id := base64.StdEncoding.EncodeToString([]byte(s + "@" + ep))
			service := &Service{Id: id, Name: s, Address: ep, Metadata: metadata, State: state}

			services = append(services, service)
		}
	}

	return services
}
