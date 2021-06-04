package plugin

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.acpr.dev/touchportal-golang-sdk/client"
	. "go.acpr.dev/touchportal-golang-sdk/plugin/mocks"
)

func TestActionEnums(t *testing.T) {
	t.Parallel()

	for _, enum := range pluginEventValues() {
		enum := enum
		t.Run(enum.String()+" matches existing client messages", func(t *testing.T) {
			t.Parallel()

			_, err := client.ClientMessageTypeString(enum.String())
			if err != nil {
				t.Errorf("action type of %v not found as client message type", enum)
			}
		})
	}
}

func TestPlugin_on(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	messageType, _ := client.ClientMessageTypeString("info")
	messageHandler := func(event interface{}) {}
	mc.EXPECT().AddMessageHandler(messageType, gomock.AssignableToTypeOf(reflect.TypeOf(messageHandler)))

	p := &Plugin{
		client: mc,
	}

	p.on(eventInfo, func(e interface{}) {})
}

func TestPlugin_onActionHandler(t *testing.T) {
	t.Parallel()

	msg := client.ActionMessage{
		PluginID: "testPlugin",
		ActionID: "test",
	}

	var called bool = false

	handler := func(event client.ActionMessage) {
		called = true
	}
	actionID := "test"

	p := &Plugin{
		ID: "testPlugin",
	}

	p.onActionHandler(handler, actionID)(msg)

	if !called {
		t.Error("handler function not called despite good data")
	}
}

func TestPlugin_onClosePluginHandler(t *testing.T) {
	t.Parallel()

	msg := client.ClosePluginMessage{
		PluginID: "testPlugin",
	}

	var called bool = false

	handler := func(event client.ClosePluginMessage) {
		called = true
	}

	p := &Plugin{
		ID: "testPlugin",
	}

	p.onClosePluginHandler(handler)(msg)

	if !called {
		t.Error("handler function not called despite good data")
	}
}
