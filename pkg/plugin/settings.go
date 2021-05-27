package plugin

import (
	"encoding/json"
	"log"
	"reflect"

	"go.acpr.dev/touchportal-golang-sdk/pkg/client"
)

type Settings interface {
	IsUpdated()
}

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
			log.Panicf("it is only possible to have settings that are strings or integers; field %s is of type %s\n", rvs.Type().Field(i).Name, kind)
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

		obj, ok := p.settings.(Settings)
		if ok {
			obj.IsUpdated()
		}
	})
}
