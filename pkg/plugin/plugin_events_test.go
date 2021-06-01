package plugin

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.acpr.dev/touchportal-golang-sdk/pkg/client"
	"go.acpr.dev/touchportal-golang-sdk/pkg/mocks"
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
