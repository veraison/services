## Veraison Services Configuration

Veraison services use [Viper](https://github.com/spf13/viper) for
configuration. Viper is great, but it can be a little too accepting. Config
loader implemented here adds a validation layer on top of viper that allows
pre-processing config collected by Viper before it is used. Full recognition
before processing.

### The General Idea

- Configuration for each component is defined by its own `struct`.
- The loader is given an instance of the `struct` to populate. Hard-coded
  defaults can be provided by setting field values of this instance.
- Viper is used to collect configuration for all components across various
  sources (files, environment variables, set by client code, etc.).
- Viper's `Sub()` is used to create a Viper instance with the sub-tree of
  collected configuration relevant to the component.
- The component-relevant viper is passed to the loader that uses it to populate
  the config `struct`. Loader does this by calling Viper's `AllSettings()` which
  returns a `map[string]interface{}`; the loader can also be passed such a map
  directly.
- If the `struct` defines a `Validate() error` method, the loader will then
  invoke that. This can be used to validate additional constraints on
  the settings that go beyond their basic type.
- The component then uses configuration from the `struct` without needing to
  worry about validating each individual setting.


### Validation

The loader enforces the following invariants:

- Setting values are of the correct type (this is ensured by unmarshalling into
  a `struct`).
- All settings have been set. There is no explicit concept of optional
  settings; settings become optional if a default is provided. This ensures
  that a component always has valid values to operate on.
- Everything collected by Viper has been been processed, there are no "extras".
  This helps catch typos and dead config that is not actually used to influence
  components. This is why `Sub()` should be used to extract the
  component-relevant sub-tree from Viper before giving to the loader. (note:
  this constraint can be relaxed by creating a non-exclusive loader with
  `NewNonExclusiveLoader()`. This may be useful for processing nested configs.)
- Additional constraints for a field can be specified using a `govalidator`
  tag. Valid values for the tag is listed in the [`govalidator
  README](https://github.com/asaskevich/govalidator/blob/master/README.md).
- Any additional, including inter-setting (i.e. co-constraints), constraints can be specified by
  implementing `Validate() error` for the config `struct`.

### Field names in configuration

The loader uses [mapstructure](https://github.com/mitchellh/mapstructure) to
unmarshal settings from a source `map` into the destination `struct`. Setting
names from configuration source are matched to `struct` field names via
`strings.EqualFold`. If setting names are different from field names (e.g. if
they contain hyphens or other special characters), it is possible to specify
the name using "mapstructure" tag. E.g.

```golang
type Config struct {
        HyphenatedSetting string `mapstructure:"hyphenated-setting"`
}
```

### Handling of Defaults

There are two ways to provide default values from the code, depending on
whether the default is being set by the component itself (e.g. vts server
setting default address it will listen on), or by another component using it
(e.g. polcli specifying defaults for the policy store).

If the defaults are set by the component itself, they maybe set simply by
specifying them as field values when instantiating the config struct.

If the defaults come from the instantiating code, rather than component itself,
they maybe set inside the Viper instance being passed to the component using
Viper's `SetDefault()`.


#### Zero-values in Defaults

It is not possible to distinguish between an unset `struct` field, and one that
has been set to that type's Zero value. (Analogously, when Get'ing an unset
setting in Viper, one gets the getter type's Zero value.) Since the goal is
robustness, we treat Zero value defaults as unset and raise an error if a value
for that field hasn't been provided in the input (note: zero values in the
input are OK, as their mere presence indicates that they have been explicitly
set).

If you wish to treat a field as defaulting to its type's zero value, rather
than as unset, you have to set `config:"zerodefault"` tag on that field:

```golang
type Config struct {
        MustBeSet string
        CanBeEmpty string `config:"zerodefault"`
        CanBeZero int `config:"zerodefault"`
}
```

### Example

```golang
import (
        "fmt"
        "log"
        "net/url"
)

type Config struct {
	Host       string
	Port       uint16
}

// Optional Validate()
func (o Config) Validate() error {
        if _, err := url.ParseRequestURI(c.Host); err != nil {
                return fmt.Errorf("bad host: %w", err)
        }

        return nil
}

func main() {
	var myConfig Config

	loader := NewLoader(&myConfig)

        // It is also possible to load directly from a Viper instance using
        // LoadFromViper().
	err := loader.LoadFromMap(map[string]interface{}{
		"host": "example.com",
		"port": 8443,
	})
	if err != nil {
            log.Fatal(err)
        }

        // myConfig.Host and myConfig.Port have now been set to valid values

}
```

