package netconf

import (
	"fmt"
	"os"
	"text/template"
)

const netconfigTpl = `

config interface 'loopback'
	option device 'lo'
	option proto 'static'
	option ipaddr '127.0.0.1'
	option netmask '255.0.0.0'

config globals 'globals'
	option ula_prefix 'fdc6:225:4c9f::/48'

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

type WrtxNetConfig struct {
	Nic     string
	IPAddr  string
	Netmask string
	Gateway string
	DNS     string
}

func GenerateNetConfig(cfg WrtxNetConfig, output *os.File) error {
	tpl, err := template.New("").Parse(netconfigTpl)
	if err != nil {
		return fmt.Errorf("failed to parse netconfig template: %v", err)
	}
	return tpl.Execute(output, cfg)
}

func NewWrtxNetConfig(nic, ipaddr, netmask, gateway, dns string) WrtxNetConfig {
	return WrtxNetConfig{
		Nic:     nic,
		IPAddr:  ipaddr,
		Netmask: netmask,
		Gateway: gateway,
		DNS:     dns,
	}
}
