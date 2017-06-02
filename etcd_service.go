package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	requestTimeOut = 10 * time.Second
)

type EtcdRegistry struct {
	Cli *clientv3.Client
}

func (r *EtcdRegistry) initRegistry() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(serverConfig.RegistryURL, ","),
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(err)
	}
	r.Cli = cli
	return
}

func (r *EtcdRegistry) fetchServices() []*Service {
	var services []*Service
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeOut)
	defer cancel()
	resp, err := r.Cli.Get(ctx, serverConfig.ServiceBaseURL, clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		log.Println("fetchServices error: ", err.Error())
		return nil
	}
	for _, value := range resp.Kvs {
		fmt.Println("etcd v3 values: %v", value)
		key := string(value.Key[:])
		serviceName := strings.TrimPrefix(key, serverConfig.ServiceBaseURL+"/")
		var serviceAddr string
		fields := strings.Split(key, "/")
		if fields != nil && len(fields) > 1 {
			serviceAddr = fields[len(fields)-1]
		}
		v, err := url.ParseQuery(string(value.Value[:]))
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
		service := &Service{ID: id, Name: serviceName, Address: serviceAddr, Metadata: string(value.Value[:]), State: state, Group: group}
		services = append(services, service)
	}

	return services
}

func (r *EtcdRegistry) deactivateService(name, address string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	resp, err := r.Cli.Get(context.Background(), key)

	if err != nil {
		return err
	}

	for _, value := range resp.Kvs {
		v, err := url.ParseQuery(string(value.Value[:]))
		if err != nil {
			log.Println("etcd value parse failed. err ", err.Error())
			continue
		}
		v.Set("state", "inactive")
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeOut)
		defer cancel()
		_, err = r.Cli.Put(ctx, key, v.Encode())
		if err != nil {
			log.Println("etcd set failed, err : ", err.Error())
		}
	}

	return err
}

func (r *EtcdRegistry) activateService(name, address string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeOut)
	defer cancel()
	resp, err := r.Cli.Get(ctx, key)
	if err != nil {
		return err
	}

	for _, value := range resp.Kvs {
		v, err := url.ParseQuery(string(value.Value[:]))
		if err != nil {
			log.Println("etcd value parse failed. err ", err.Error())
			continue
		}
		v.Set("state", "active")
		_, err = r.Cli.Put(ctx, key, v.Encode())
		if err != nil {
			log.Println("etcdv3 put failed. err: ", err.Error())
		}
	}

	return err
}

func (r *EtcdRegistry) updateMetadata(name, address string, metadata string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeOut)
	defer cancel()
	_, err := r.Cli.Put(ctx, key, metadata)
	return err
}
