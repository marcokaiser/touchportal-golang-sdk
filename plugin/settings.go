package plugin

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/marcokaiser/touchportal-golang-sdk/client"
)

// SettingsUpdated implemented on the struct you use to power your plugins settings
// will allow you to be made aware when the settings have been updated by TouchPortal.
type SettingsUpdated interface {
	IsUpdated()
}

// Settings allows you to provide a reference to a struct that will be populated by TouchPortal
// when the plugin registers itself or a settings update occurs.
//
// It works in a similar way to standard json Marshal/Unmarshal and is even driven by
// the same struct tags.
//
// Currently only string and int types are supported
//
//	type settings struct {
//	    Host string `json:"Host"`
//	    Port int    `json:"Port,string"`
//	}
//
//	func main() {
//	    p := NewPlugin(...)
//	    s := &settings{}
//
//	    p.Settings(s)
//	    p.Register()
//	    // p will now contain any settings that TouchPortal returned
//	}
//
// Interestingly it's important to note that TouchPortal string encodes both string and integer values
// so when Unmarshaling to an int you will need to ensure you mark it as ",string" as shown above.
func (p *Plugin) Settings(s interface{}) {
	rv := reflect.ValueOf(s)
	if rv.IsNil() || rv.Kind() != reflect.Ptr || rv.Elem().Type().Kind() != reflect.Struct {
		log.Panicf("please pass a struct ptr to the plugin.Settings function; %s passed\n", rv.Kind())
	}

	rvs := reflect.ValueOf(s).Elem()
	for i := 0; i < rvs.NumField(); i++ {
		field := rvs.Field(i)

		kind := field.Type().Kind()
		if kind != reflect.String && kind != reflect.Int {
			log.Panicf(
				"it is only possible to have settings that are strings or integers; field %s is of type %s\n",
				rvs.Type().Field(i).Name,
				kind)
		}
	}

	p.settings = s

	p.onSettings(func(event client.SettingsMessage) {
		// not sure I like doing it like this but it seems the quickest
		// turn the settings back into json
		enc, err := json.Marshal(event.Values)
		if err != nil {
			log.Panicf("failed to marshal settings back into json: %v\n", err)
		}

		// write the settings to the settings object we were given
		err = json.Unmarshal(enc, p.settings)
		if err != nil {
			log.Panicf("failed to write settings to given settings struct: %v\n", err)
		}

		obj, ok := p.settings.(SettingsUpdated)
		if ok {
			obj.IsUpdated()
		}
	})
}
