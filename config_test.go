package main

import (
	"mazarin/config"
	"testing"
)

// This test is used to check that we can parse multiple ports and urls, it also verifies we don't edit anything else.
func TestParseMulti(t *testing.T) {
	input := []config.ProxyConfig{
		{ // Basic tcp
			Port:       ":25565",
			TargetAddr: "192.168.129.88:25565",
			Protocol:   "tcp",
		},
		{ // Basic web
			ListenUrl:  "test1.domain.com",
			Port:       ":443",
			TargetAddr: "192.168.129.88:80",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi port parsing tcp
			Ports:      []string{":25565", ":3000"},
			TargetAddr: "192.168.129.88:25565",
			Protocol:   "tcp",
		},
		{ // Multi port parsing web
			ListenUrl:  "test2.domain.com",
			Ports:      []string{":443", ":80"},
			TargetAddr: "192.168.129.88:80",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi url parsing web
			ListenUrls: []string{"test3.domain.com", "test4.domain.com"},
			Port:       ":80",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi both parsing web
			ListenUrls: []string{"test5.domain.com", "test6.domain.com"},
			Ports:      []string{":80", ":443"},
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
	}
	expected := []config.ProxyConfig{
		{ // Basic tcp
			Port:       ":25565",
			TargetAddr: "192.168.129.88:25565",
			Protocol:   "tcp",
		},
		{ // Basic web
			ListenUrl:  "test1.domain.com",
			Port:       ":443",
			TargetAddr: "192.168.129.88:80",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi port parsing tcp
			Port:       ":25565",
			TargetAddr: "192.168.129.88:25565",
			Protocol:   "tcp",
		},
		{
			Port:       ":3000",
			TargetAddr: "192.168.129.88:25565",
			Protocol:   "tcp",
		},
		{ // Multi port parsing web
			ListenUrl:  "test2.domain.com",
			Port:       ":443",
			TargetAddr: "192.168.129.88:80",
			Type:       "proxy",
			Protocol:   "web",
		},
		{
			ListenUrl:  "test2.domain.com",
			Port:       ":80",
			TargetAddr: "192.168.129.88:80",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi url parsing web
			ListenUrl:  "test3.domain.com",
			Port:       ":80",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{
			ListenUrl:  "test4.domain.com",
			Port:       ":80",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{ // Multi both parsing web
			ListenUrl:  "test5.domain.com",
			Port:       ":80",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{
			ListenUrl:  "test5.domain.com",
			Port:       ":443",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{
			ListenUrl:  "test6.domain.com",
			Port:       ":80",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
		{
			ListenUrl:  "test6.domain.com",
			Port:       ":443",
			TargetAddr: "192.168.129.88:443",
			Type:       "proxy",
			Protocol:   "web",
		},
	}
	output := config.ParseMulti(input)
	for i, result := range output {
		if result.ListenUrl != expected[i].ListenUrl {
			t.Errorf("[Config: %d] ListenUrl: got %v, want %v", i, result.ListenUrl, expected[i].ListenUrl)
		}
		if result.Port != expected[i].Port {
			t.Errorf("[Config: %d] Port: got %v, want %v", i, result.Port, expected[i].Port)
		}
		if result.TargetAddr != expected[i].TargetAddr {
			t.Errorf("[Config: %d] TargetAddr: got %v, want %v", i, result.TargetAddr, expected[i].TargetAddr)
		}
		if result.Type != expected[i].Type {
			t.Errorf("[Config: %d] Type: got %v, want %v", i, result.Type, expected[i].Type)
		}
		if result.Protocol != expected[i].Protocol {
			t.Errorf("[Config: %d] Protocol: got %v, want %v", i, result.Protocol, expected[i].Protocol)
		}
	}
}
