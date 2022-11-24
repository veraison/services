package plugin

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
)

var defaultLoader *Loader

type unknownPluginErr struct {
	Name string
}

func (o unknownPluginErr) Error() string {
	return fmt.Sprintf("unknown plugin: %s", o.Name)
}

type LoaderConfig struct {
	Directory string `mapstructure:"dir"`
}

type Loader struct {
	Location string

	logger                *zap.SugaredLogger
	loaded                map[string]IPluginContext
	pluginMap             map[string]plugin.Plugin
	registeredPluginTypes map[string]string
}

func NewLoader(logger *zap.SugaredLogger) *Loader {
	return &Loader{logger: logger}
}

func CreateLoader(cfg map[string]interface{}, logger *zap.SugaredLogger) (*Loader, error) {
	pm := NewLoader(logger)
	err := pm.Init(cfg)
	return pm, err
}

func (o *Loader) Init(m map[string]interface{}) error {
	o.pluginMap = make(map[string]plugin.Plugin)
	o.loaded = make(map[string]IPluginContext)
	o.registeredPluginTypes = make(map[string]string)

	var cfg LoaderConfig
	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromMap(m); err != nil {
		return err
	}

	o.Location = cfg.Directory

	return nil
}

func (o *Loader) Close() {
	for _, plugin := range o.loaded {
		plugin.Close()
	}
}

func (o *Loader) GetRegisteredMediaTypes() []string {
	var mediaTypes []string

	for mt := range o.loaded {
		mediaTypes = append(mediaTypes, mt)
	}

	return mediaTypes
}

func (o *Loader) GetRegisteredMediaTypesByName(name string) []string {
	var mediaTypes []string

	for mt, pc := range o.loaded {
		if pc.GetTypeName() == name {
			mediaTypes = append(mediaTypes, mt)
		}
	}

	return mediaTypes
}

func Init(m map[string]interface{}) error {
	return defaultLoader.Init(m)
}

func Register[I IPluggable](name string, ch RPCChannel[I]) error {
	return RegisterUsing(defaultLoader, name, ch)
}

func RegisterUsing[I IPluggable](man *Loader, name string, ch RPCChannel[I]) error {
	if _, ok := man.pluginMap[name]; ok {
		return fmt.Errorf("plugin for %q is already registred", name)
	}

	if err := registerRPCChannel(name, ch); err != nil {
		return err
	}

	man.pluginMap[name] = &Plugin[I]{Name: name}
	man.registeredPluginTypes[getTypeName[I]()] = name

	return nil
}

func Discover[I IPluggable]() error {
	return DiscoverUsing[I](defaultLoader)
}

func DiscoverUsing[I IPluggable](o *Loader) error {
	if o.Location == "" {
		return errors.New("plugin manager has not been initialized")
	}

	o.logger.Debugw("discovering plugins", "location", o.Location)
	pluginPaths, err := plugin.Discover("*.plugin", o.Location)
	if err != nil {
		return err
	}

	for _, path := range pluginPaths {
		pluginContext, err := createPluginContext[I](o, path, o.logger)
		if err != nil {
			var upErr unknownPluginErr
			if errors.As(err, &upErr) {
				o.logger.Debugw("plugin not found", "name", upErr.Name, "path", path)
				continue
			} else {
				return err
			}
		}

		for _, mediaType := range pluginContext.SupportedMediaTypes {
			if existing, ok := o.loaded[mediaType]; ok {
				return fmt.Errorf(
					"plugins %q and %q both provides support for %q",
					existing.GetName(),
					pluginContext.GetName(),
					mediaType,
				)
			}
			o.loaded[mediaType] = pluginContext
		}
	}

	return nil
}

func getTypeName[I IPluggable]() string {
	return reflect.TypeOf((*I)(nil)).Elem().Name()
}

func GetHandleByMediaType[I IPluggable](mediaType string) (I, error) {
	return GetHandleByMediaTypeUsing[I](defaultLoader, mediaType)
}

func GetHandleByMediaTypeUsing[I IPluggable](ldr *Loader, mediaType string) (I, error) {
	plugged, ok := ldr.loaded[mediaType].(*PluginContext[I])
	if !ok {
		iface := getTypeName[I]()
		return *new(I), fmt.Errorf("plugin providing %q with interface %s not found", mediaType, iface)
	}

	return plugged.Handle, nil
}

func GetHandleByNameUsing[I IPluggable](ldr *Loader, mediaType string) (I, error) {
	plugged, ok := ldr.loaded[mediaType].(*PluginContext[I])
	if !ok {
		iface := getTypeName[I]()
		return *new(I), fmt.Errorf("plugin providing %q with interface %s not found", mediaType, iface)
	}

	return plugged.Handle, nil
}

// IPluggable respresents a "pluggable" point within Veraison services. It is
// the common interfaces shared by all Veraison plugins loaded through this
// framework.
type IPluggable interface {
	// GetName returns a strin containing the the name of the
	// implementation of this IPluggable interface. It is the plugin name.
	GetName() string
	// GetSupportedMediaTypes returns a []string containing the media types
	// this plugin is capable of handling.
	GetSupportedMediaTypes() []string
}

// IPluginContext is the common interace for handling all PluginContext[I] type
// instances of the generic PluginContext[].
type IPluginContext interface {
	GetName() string
	GetTypeName() string
	Close()
}

// PluginConntext is a generic for handling Veraison services plugins. It is
// parameterised on the IPluggale interface it handles.
type PluginContext[I IPluggable] struct {
	// Path to the exectable binary containing the plugin implementation
	Path string
	// Name of this plugin
	Name string
	// SupportedMediaTypes are the types of input this plugin can process.
	// This is is the method by which a plugin is selected.
	SupportedMediaTypes []string
	// Handle is actual RPC interface to the plugin implementation.
	Handle I

	// go-plugin client
	client *plugin.Client
}

func (o PluginContext[I]) GetName() string {
	return o.Name
}

func (o PluginContext[I]) GetTypeName() string {
	return getTypeName[I]()
}

func (o PluginContext[I]) Close() {
	if o.client != nil {
		o.client.Kill()
	}
}

func createPluginContext[I IPluggable](
	man *Loader,
	path string,
	logger *zap.SugaredLogger,
) (*PluginContext[I], error) {
	client := plugin.NewClient(
		&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         man.pluginMap,
			Cmd:             exec.Command(path),
			Logger:          log.NewInternalLogger(logger),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	typeName := getTypeName[I]()
	name, ok := man.registeredPluginTypes[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type %q", name)
	}

	protocolClient, err := rpcClient.Dispense(name)
	if err != nil {
		client.Kill()
		if strings.Contains(err.Error(), "unknown plugin") {
			return nil, unknownPluginErr{Name: name}
		}
		return nil, fmt.Errorf("unable to dispense plugin %s: %w", path, err)
	}

	handle, ok := protocolClient.(I)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf(
			"protocol client %T does not provide an implementation of %s",
			protocolClient,
			getTypeName[I](),
		)
	}

	return &PluginContext[I]{
		Path:                path,
		Name:                handle.GetName(),
		SupportedMediaTypes: handle.GetSupportedMediaTypes(),
		Handle:              handle,
		client:              client,
	}, nil
}

func init() {
	defaultLoader = NewLoader(log.Named("plugin"))
}
