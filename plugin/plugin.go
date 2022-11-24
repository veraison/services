package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/services/log"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERAISON_PLUGIN",
	MagicCookieValue: "VERAISON",
}

type Plugin[I IPluggable] struct {
	Name string
	Impl I
}

func (p *Plugin[I]) Server(*plugin.MuxBroker) (interface{}, error) {
	log.Infof("[CHECK] getting server for %s: %v", p.Name, p.Impl)
	return GetRPCServer(p.Name, p.Impl), nil
}

func (p *Plugin[I]) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return GetRPCClient(p.Name, p.Impl, c), nil
}

var pluginMap = map[string]plugin.Plugin{}

func RegisterImplementation[I IPluggable](name string, i I, ch RPCChannel[I]) {
	pluginMap[name] = &Plugin[I]{
		Name: name,
		Impl: i,
	}

	registerRPCChannel(name, ch)

}

func Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
