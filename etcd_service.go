package main

import (
	"encoding/base64"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/docker/libkv"
	kvstore "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

type EtcdRegistry struct {
	kv kvstore.Store
}

func (r *EtcdRegistry) initRegistry() {
	etcd.Register()

	kv, err := libkv.NewStore(kvstore.ETCD, []string{serverConfig.RegistryURL}, nil)
	if err != nil {
		log.Printf("cannot create etcd registry: %v", err)
		return
	}
	r.kv = kv

	return
}

func (r *EtcdRegistry) fetchServices() []*Service {
	var services []*Service
	kvs, err := r.kv.List(serverConfig.ServiceBaseURL)
	if err != nil {
		log.Printf("failed to list services %s: %v", serverConfig.ServiceBaseURL, err)
		return services
	}

	for _, value := range kvs {

		nodes, err := r.kv.List(value.Key)
		if err != nil {
			log.Printf("failed to list %s: %v", value.Key, err)
			continue
		}

		for _, n := range nodes {
			key := string(n.Key[:])
			i := strings.LastIndex(key, "/")
			serviceName := strings.TrimPrefix(key[0:i], serverConfig.ServiceBaseURL)
			var serviceAddr string
			fields := strings.Split(key, "/")
			if fields != nil && len(fields) > 1 {
				serviceAddr = fields[len(fields)-1]
			}
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

func (r *EtcdRegistry) deactivateService(name, address string) error {
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

func (r *EtcdRegistry) activateService(name, address string) error {
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

func (r *EtcdRegistry) updateMetadata(name, address string, metadata string) error {
	key := path.Join(serverConfig.ServiceBaseURL, name, address)
	err := r.kv.Put(key, []byte(metadata), &kvstore.WriteOptions{IsDir: false})
	return err
}
