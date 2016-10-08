package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type ZooKeeperRegistry struct {
	conn *zk.Conn
}

func (r *ZooKeeperRegistry) initRegistry() {
	c, _, err := zk.Connect(strings.Split(serverConfig.RegistryURL, ","), 10*time.Second)
	r.conn = c
	if err != nil {
		panic(err)
	}

}
func (r *ZooKeeperRegistry) fetchServices() []*Service {
	c := r.conn
	if c == nil {
		r.initRegistry()
		c = r.conn
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
			bytes, _, err := c.Get(serverConfig.ServiceBaseURL + "/" + s + "/" + ep)
			if err != nil {
				fmt.Println(err)
				continue
			}

			metadata := string(bytes)
			if metadata == "0" {
				continue
			}

			v, err := url.ParseQuery(metadata)
			state := "n/a"
			group := ""
			if err == nil {
				state = v.Get("state")
				if state == "" {
					state = "active"
				}

				group = v.Get("group")
			}

			id := base64.StdEncoding.EncodeToString([]byte(s + "@" + ep))
			service := &Service{ID: id, Name: s, Address: ep, Metadata: metadata, State: state, Group: group}

			services = append(services, service)
		}
	}

	return services
}

func (r *ZooKeeperRegistry) deactivateService(name, address string) error {
	c := r.conn
	if c == nil {
		r.initRegistry()
		c = r.conn
	}

	exist, _, _ := c.Exists(serverConfig.ServiceBaseURL)
	if !exist {
		return errors.New("base path: " + serverConfig.ServiceBaseURL + " doesn't exist")
	}

	bytes, stat, err := c.Get(serverConfig.ServiceBaseURL + "/" + name + "/" + address)
	if err != nil {
		return err
	}

	metadata := string(bytes)
	v, err := url.ParseQuery(metadata)
	v.Set("state", "inactive")
	bytes = []byte(v.Encode())
	_, err = c.Set(serverConfig.ServiceBaseURL+"/"+name+"/"+address, bytes, stat.Version)
	return err
}

func (r *ZooKeeperRegistry) activateService(name, address string) error {
	c := r.conn
	if c == nil {
		r.initRegistry()
		c = r.conn
	}

	exist, _, _ := c.Exists(serverConfig.ServiceBaseURL)
	if !exist {
		return errors.New("base path: " + serverConfig.ServiceBaseURL + " doesn't exist")
	}

	bytes, stat, _ := c.Get(serverConfig.ServiceBaseURL + "/" + name + "/" + address)
	metadata := string(bytes)
	v, err := url.ParseQuery(metadata)
	v.Set("state", "active")
	bytes = []byte(v.Encode())
	_, err = c.Set(serverConfig.ServiceBaseURL+"/"+name+"/"+address, bytes, stat.Version)
	return err
}

func (r *ZooKeeperRegistry) updateMetadata(name, address string, metadata string) error {
	c := r.conn
	if c == nil {
		r.initRegistry()
		c = r.conn
	}

	exist, _, _ := c.Exists(serverConfig.ServiceBaseURL)
	if !exist {
		return errors.New("base path: " + serverConfig.ServiceBaseURL + " doesn't exist")
	}

	bytes, stat, _ := c.Get(serverConfig.ServiceBaseURL + "/" + name + "/" + address)
	bytes = []byte(metadata)
	_, err := c.Set(serverConfig.ServiceBaseURL+"/"+name+"/"+address, bytes, stat.Version)
	return err
}
