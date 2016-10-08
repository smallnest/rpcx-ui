package main

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/coreos/etcd/client"
)

type EtcdRegistry struct {
	keysAPI client.KeysAPI
}

func (r *EtcdRegistry) initRegistry() {
	cli, err := client.New(client.Config{
		Endpoints: strings.Split(serverConfig.RegistryURL, ","),
		Transport: client.DefaultTransport,
	})

	if err != nil {
		panic(err)
	}
	r.keysAPI = client.NewKeysAPI(cli)

}
func (r *EtcdRegistry) fetchServices() []*Service {

	var services []*Service
	resp, err := r.keysAPI.Get(context.TODO(), serverConfig.ServiceBaseURL, &client.GetOptions{
		Recursive: true,
		Sort:      true,
	})
	if err == nil && resp.Node != nil {
		if len(resp.Node.Nodes) > 0 {
			for _, n := range resp.Node.Nodes {
				for _, ep := range n.Nodes {

					serviceName := strings.TrimPrefix(n.Key, serverConfig.ServiceBaseURL+"/")
					serviceAddr := strings.TrimPrefix(ep.Key, n.Key+"/")
					v, err := url.ParseQuery(ep.Value)
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
					service := &Service{ID: id, Name: serviceName, Address: serviceAddr, Metadata: ep.Value, State: state, Group: group}
					services = append(services, service)
				}
			}

		}

	}

	return services
}

func (r *EtcdRegistry) deactivateService(name, address string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	resp, err := r.keysAPI.Get(context.TODO(), key, &client.GetOptions{
		Recursive: false,
	})

	if err != nil {
		return err
	}

	metadata := resp.Node.Value
	v, err := url.ParseQuery(metadata)
	v.Set("state", "inactive")

	_, err = r.keysAPI.Set(context.TODO(), key, v.Encode(), &client.SetOptions{
		PrevExist: client.PrevIgnore,
	})

	return err
}

func (r *EtcdRegistry) activateService(name, address string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	resp, err := r.keysAPI.Get(context.TODO(), key, &client.GetOptions{
		Recursive: false,
	})

	if err != nil {
		return err
	}

	metadata := resp.Node.Value
	v, err := url.ParseQuery(metadata)
	v.Set("state", "active")

	_, err = r.keysAPI.Set(context.TODO(), key, v.Encode(), &client.SetOptions{
		PrevExist: client.PrevIgnore,
	})

	return err
}

func (r *EtcdRegistry) updateMetadata(name, address string, metadata string) error {
	key := serverConfig.ServiceBaseURL + "/" + name + "/" + address

	_, err := r.keysAPI.Set(context.TODO(), key, metadata, &client.SetOptions{
		PrevExist: client.PrevIgnore,
	})

	return err
}
