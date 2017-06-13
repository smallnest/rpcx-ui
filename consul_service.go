package main

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/api"
)

type ConsulRegistry struct {
	consulAddress string
	consulConfig  *api.Config
	client        *api.Client
}

func (r *ConsulRegistry) initRegistry() {
	if r.consulConfig == nil {
		r.consulConfig = api.DefaultConfig()
		r.consulConfig.Address = r.consulAddress
	}
	r.client, _ = api.NewClient(r.consulConfig)

}
func (r *ConsulRegistry) fetchServices() []*Service {
	var services []*Service

	agent := r.client.Agent()
	ass, err := agent.Services()
	if err != nil {
		return nil
	}

	for as, v := range ass {
		if strings.Contains(v.Address, "@") {

			var meta string
			if len(v.Tags) > 0 {
				meta = v.Tags[0]
			}

			metaData, err := url.ParseQuery(meta)
			state := "n/a"
			group := ""
			if err == nil {
				state = metaData.Get("state")
				if state == "" {
					state = "active"
				}
				group = metaData.Get("group")
			}

			id := base64.StdEncoding.EncodeToString([]byte(as))
			s := &Service{
				ID:       id,
				Name:     v.Service,
				Address:  v.Address,
				State:    state,
				Metadata: meta,
				Group:    group,
			}

			services = append(services, s)
		}
	}

	return services
}

func (r *ConsulRegistry) deactivateService(name, address string) error {
	return errors.New("unsupport action")
}

func (r *ConsulRegistry) activateService(name, address string) error {
	// agent := r.client.Agent()
	// ass, err := agent.Services()
	// if err != nil {
	// 	return nil
	// }

	// var s *api.AgentService
	// for _, v := range ass {
	// 	if v.Service == name && v.Address == address {
	// 		s = v
	// 		break
	// 	}
	// }

	// metadata := ""
	// if len(s.Tags) > 0 {
	// 	metadata = s.Tags[0]
	// }
	// v, err := url.ParseQuery(metadata)
	// v.Set("state", "inactive")
	// s.Tags[0] = v.Encode()

	// // how to set Check?
	// agent.ServiceRegister()

	return errors.New("unsupport action")
}

func (r *ConsulRegistry) updateMetadata(name, address string, metadata string) error {
	return errors.New("unsupport action")
}
