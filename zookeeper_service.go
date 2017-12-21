package main

import (
	"encoding/base64"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/docker/libkv"
	kvstore "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
)

type ZooKeeperRegistry struct {
	kv kvstore.Store
}

func (r *ZooKeeperRegistry) initRegistry() {
	zookeeper.Register()

	if strings.HasPrefix(serverConfig.ServiceBaseURL, "/") {
		serverConfig.ServiceBaseURL = serverConfig.ServiceBaseURL[1:]
	}

	if strings.HasSuffix(serverConfig.ServiceBaseURL, "/") {
		serverConfig.ServiceBaseURL = serverConfig.ServiceBaseURL[0 : len(serverConfig.ServiceBaseURL)-1]
	}

	kv, err := libkv.NewStore(kvstore.ZK, []string{serverConfig.RegistryURL}, nil)
	if err != nil {
		log.Printf("cannot create etcd registry: %v", err)
		return
	}
	r.kv = kv

	return
}

func (r *ZooKeeperRegistry) fetchServices() []*Service {
	var services []*Service

	kvs, err := r.kv.List(serverConfig.ServiceBaseURL)
	if err != nil {
		log.Printf("failed to list services %s: %v", serverConfig.ServiceBaseURL, err)
		return services
	}

	for _, value := range kvs {
		serviceName := value.Key

		nodes, err := r.kv.List(serverConfig.ServiceBaseURL + "/" + value.Key)
		if err != nil {
			log.Printf("failed to list  %s: %v", serverConfig.ServiceBaseURL+"/"+value.Key, err)
			continue
		}

		for _, n := range nodes {
			var serviceAddr = n.Key

			v, err := url.ParseQuery(string(n.Value[:]))
			if err != nil {
				log.Println("etcd value parse failed. error: ", err.Error())
				continue
			}
			state := "n/a"
			group := ""
			if err == nil {
				state = v.Get("state")
				if state == "" {
					state = "active"
				}
				group = v.Get("group")
			}
			id := base64.StdEncoding.EncodeToString([]byte(serviceName + "@" + serviceAddr))
			service := &Service{ID: id, Name: serviceName, Address: serviceAddr, Metadata: string(n.Value[:]), State: state, Group: group}
			services = append(services, service)
		}

	}

	return services
}

func (r *ZooKeeperRegistry) deactivateService(name, address string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)

	kv, err := r.kv.Get(key)

	if err != nil {
		return err
	}

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("etcd value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "inactive")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &kvstore.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("etcd set failed, err : ", err.Error())
	}

	return err
}

func (r *ZooKeeperRegistry) activateService(name, address string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)
	kv, err := r.kv.Get(key)

	v, err := url.ParseQuery(string(kv.Value[:]))
	if err != nil {
		log.Println("etcd value parse failed. err ", err.Error())
		return err
	}
	v.Set("state", "active")
	err = r.kv.Put(kv.Key, []byte(v.Encode()), &kvstore.WriteOptions{IsDir: false})
	if err != nil {
		log.Println("etcdv3 put failed. err: ", err.Error())
	}

	return err
}

func (r *ZooKeeperRegistry) updateMetadata(name, address string, metadata string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)
	err := r.kv.Put(key, []byte(metadata), &kvstore.WriteOptions{IsDir: false})
	return err
}
