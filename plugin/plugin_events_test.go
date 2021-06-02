package plugin

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.acpr.dev/touchportal-golang-sdk/client"
	"go.acpr.dev/touchportal-golang-sdk/mocks"
)

func TestActionEnums(t *testing.T) {
	for _, enum := range pluginEventValues() {
		t.Run("its events match existing client messages", func(t *testing.T) {
			_, err := client.ClientMessageTypeString(enum.String())
			if err != nil {
				t.Errorf("action type of %v not found as client message type", enum)
			}
		})
	}
}

func TestPlugin_on(t *testing.T) {
	t.Run("it registers a handler for an event with the client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		messageType, _ := client.ClientMessageTypeString("info")
		messageHandler := func(event interface{}) {}
		mc.EXPECT().AddMessageHandler(messageType, gomock.AssignableToTypeOf(reflect.TypeOf(messageHandler)))

		p := &Plugin{
			client: mc,
		}

		p.on(eventInfo, func(e interface{}) {})
	})
}

func TestPlugin_onActionHandler(t *testing.T) {
	t.Run("it correctly validates and acts upon appropriate client messages", func(t *testing.T) {
		msg := client.ActionMessage{
			PluginId: "testPlugin",
			ActionId: "test",
		}

		var called bool = false
		handler := func(event client.ActionMessage) {
			called = true
		}
		actionId := "test"

		p := &Plugin{
			Id: "testPlugin",
		}

		p.onActionHandler(handler, actionId)(msg)

		if !called {
			t.Error("handler function not called despite good data")
		}
	})
}

func TestPlugin_onClosePluginHandler(t *testing.T) {
	t.Run("it correctly validates and acts upon appropriate client messages", func(t *testing.T) {
		msg := client.ClosePluginMessage{
			PluginId: "testPlugin",
		}

		var called bool = false
		handler := func(event client.ClosePluginMessage) {
			called = true
		}

		p := &Plugin{
			Id: "testPlugin",
		}

		p.onClosePluginHandler(handler)(msg)

		if !called {
			t.Error("handler function not called despite good data")
		}
	})
}
