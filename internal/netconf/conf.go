package netconf

import (
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"
)

const netconfigTpl = `

config interface 'loopback'
	option device 'lo'
	option proto 'static'
	option ipaddr '127.0.0.1'
	option netmask '255.0.0.0'

config globals 'globals'
	option ula_prefix '{{- .IPv6Prefix -}}'

config interface 'lan'
	option device '{{- .Nic -}}'
	option proto 'static'
	option ipaddr '{{- .IPAddr -}}'
	option netmask '{{- .Netmask -}}'
	option ip6assign '60'
	option gateway '{{- .Gateway -}}'
	list dns '{{- .DNS -}}'

config device
	option name '{{- .Nic -}}'
`

const brNetconfigTpl = `
config interface 'loopback'
	option device 'lo'
	option proto 'static'
	option ipaddr '127.0.0.1'
	option netmask '255.0.0.0'

config globals 'globals'
	option ula_prefix '{{- .IPv6Prefix -}}'

config device
	option name 'br-lan'
	option type 'bridge'
	list ports '{{- .Nic -}}'

config interface 'lan'
	option device 'br-lan'
	option proto 'static'
	option ip6assign '60'
	option ipaddr '{{- .IPAddr -}}'
	option netmask '{{- .Netmask -}}'
	option gateway '{{- .Gateway -}}'
	list dns '{{- .DNS -}}'
`

type WrtxNetConfig struct {
	Nic        string
	IPAddr     string
	Netmask    string
	Gateway    string
	DNS        string
	IPv6Prefix string
}

func GenerateNetConfig(cfg WrtxNetConfig, br bool, output *os.File) error {
	sTpl := netconfigTpl
	if br {
		sTpl = brNetconfigTpl
	}
	tpl, err := template.New("").Parse(sTpl)
	if err != nil {
		return fmt.Errorf("failed to parse netconfig template: %v", err)
	}
	return tpl.Execute(output, cfg)
}

func GenerateRandIPv6Prefix() string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	parts := make([]interface{}, 8)
	for i := range parts {
		parts[i] = rand.Intn(16)
	}
	return fmt.Sprintf("fdf7:%x%x%x%x:%x%x%x%x::/48", parts...)
}

func NewWrtxNetConfig(nic, ipaddr, netmask, gateway, dns string) WrtxNetConfig {
	return WrtxNetConfig{
		Nic:        nic,
		IPAddr:     ipaddr,
		Netmask:    netmask,
		Gateway:    gateway,
		DNS:        dns,
		IPv6Prefix: GenerateRandIPv6Prefix(),
	}
}
