package plugin

import (
	"encoding/json"
	"log"

	"go.acpr.dev/touchportal-golang-sdk/client"
)

//go:generate enumer -type=pluginEvent -json -transform=lower-camel -output plugin_events_string.go  -trimprefix event
type pluginEvent int

const (
	eventAction pluginEvent = iota
	eventClosePlugin
	eventInfo
	eventSettings
)

func (p *Plugin) on(event pluginEvent, handler func(event interface{})) {
	t, err := client.ClientMessageTypeString(event.String())
	if err != nil {
		log.Fatalf("unable to create event type, %v", err)
	}

	p.client.AddMessageHandler(t, handler)
}

// OnAction allows the registration of an event handler to the "action" TouchPortal message.
// The matching of the actionId parameter to the one sent by TouchPortal is handled for you
// and your passed handler function will only be executed if it matches.
func (p *Plugin) OnAction(handler func(event client.ActionMessage), actionId string) {
	p.on(eventAction, p.onActionHandler(handler, actionId))
}

// OnClosePlugin allows the registration of an event handler to the "closePlugin" TouchPortal
// message. A default handler is already in place to close down the plugin itself but you
// may wish to add an additional hook so you can carry out other shutdown tasks.
func (p *Plugin) OnClosePlugin(handler func(event client.ClosePluginMessage)) {
	p.on(eventClosePlugin, p.onClosePluginHandler(handler))
}

// OnInfo allows the registration of an event handler to the "info" TouchPortal message.
// As the "info" message is only sent as a part of the registration process it is necessary
// to register any custom handlers before plugin.Register function is called.
func (p *Plugin) OnInfo(handler func(event client.InfoMessage)) {
	p.on(eventInfo, func(e interface{}) {
		handler(e.(client.InfoMessage))
	})
}

// onSettings sets up the necessary processing to turn a message containing settings
// into a data structure that can be packed into a user supplied struct.
func (p *Plugin) onSettings(handler func(event client.SettingsMessage)) {
	p.on(eventSettings, func(e interface{}) {
		msg, ok := e.(client.SettingsMessage)
		if !ok {
			log.Printf("failed to assert event is SettingsMessage: %+v", msg)
			return
		}

		settings := new([]map[string]interface{})
		err := json.Unmarshal(msg.RawValues, settings)
		if err != nil {
			log.Printf("failed to unmarshal settings from raw data: %v", err)
			return
		}

		flat := make(map[string]interface{}, len(*settings))
		for _, setting := range *settings {
			for k, v := range setting {
				flat[k] = v
			}
		}
		msg.Values = flat

		handler(msg)
	})
}

func (p *Plugin) onActionHandler(handler func(event client.ActionMessage), actionId string) func(e interface{}) {
	return func(e interface{}) {
		action, ok := e.(client.ActionMessage)
		if !ok {
			return
		}

		if action.PluginId == p.Id && action.ActionId == actionId {
			handler(action)
		}
	}
}

func (p *Plugin) onClosePluginHandler(handler func(event client.ClosePluginMessage)) func(e interface{}) {
	return func(e interface{}) {
		action, ok := e.(client.ClosePluginMessage)
		if !ok {
			return
		}

		if action.PluginId == p.Id {
			handler(e.(client.ClosePluginMessage))
		}
	}
}
