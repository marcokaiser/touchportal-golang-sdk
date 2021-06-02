package plugin

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.acpr.dev/touchportal-golang-sdk/client"
	"go.acpr.dev/touchportal-golang-sdk/mocks"
)

func TestNewPluginWithClient(t *testing.T) {
	t.Run("it can create a test plugin with custom client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		ctx := context.Background()

		mc.EXPECT().Run(ctx)

		ready := make(chan bool)
		go func() {
			// we need to wait till marking the client ready so the
			// goroutine has a chance to process the Run expectation
			time.Sleep(time.Microsecond * 100)
			ready <- true
		}()
		mc.EXPECT().Ready().Return(ready)

		sut := NewPluginWithClient(ctx, mc, "test")

		if sut.Id != "test" {
			t.Fail()
		}
	})
}

func TestPlugin_UpdateState(t *testing.T) {
	t.Run("it can update the plugins state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		mc.
			EXPECT().
			SendMessage(
				client.NewStateUpdateMessage("testState", "testStateValue"),
			).
			Return(nil)

		p := &Plugin{
			Id:     "test",
			client: mc,
		}

		p.UpdateState("testState", "testStateValue")
	})
}

func TestPlugin_Done(t *testing.T) {
	t.Run("it provides a way to know the plugin is done running", func(t *testing.T) {
		p := &Plugin{
			done: make(chan bool),
		}

		wait := make(chan bool)
		go func() {
			defer close(wait)

			// the actual function call we're testing
			<-p.Done()
		}()

		// something changes the stop status inside the plugin
		p.done <- true

		select {
		case <-wait:
		case <-time.After(100 * time.Millisecond):
			t.Fail()
		}
	})
}

func TestPlugin_Register(t *testing.T) {
	type fields struct {
		Id                 string
		TouchPortalVersion string
		SdkVersion         int
		PluginVersion      int

		Client PluginClient
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "it successfully registers a plugin instance",
			fields: fields{
				Id:                 "test",
				TouchPortalVersion: "version",
				SdkVersion:         1,
				PluginVersion:      1,
				Client:             registrationSuccessMocks(t, "test", "version", 1, 1),
			},
			wantErr: false,
		},
		{
			name: "it handles the failure to send a registration request",
			fields: fields{
				Id:     "test",
				Client: registrationFailureMocks(t, "test"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				Id:     tt.fields.Id,
				client: tt.fields.Client,
			}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()

				if err := p.Register(); (err != nil) != tt.wantErr {
					t.Errorf("Plugin.Register() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			wg.Wait()
		})
	}
}

func TestPlugin_infoReceivedHandler(t *testing.T) {
	t.Run("it provides a handler to deal with info messages from the client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		p := &Plugin{
			Id:     "test",
			client: mc,
		}

		m := client.InfoMessage{
			Message:       client.Message{Type: client.MessageTypeInfo},
			Version:       "version",
			PluginVersion: 1,
			SdkVersion:    1,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		sut := p.infoReceivedHandler(&wg)

		sut(m)

		if p.TouchPortalVersion != m.Version {
			t.Errorf("Failed to set TouchPortal version want: %s got: %s", m.Version, p.TouchPortalVersion)
		}
		if p.PluginVersion != m.PluginVersion {
			t.Errorf("Failed to set plugin version want: %d got: %d", m.PluginVersion, p.PluginVersion)
		}
		if p.SdkVersion != m.SdkVersion {
			t.Errorf("Failed to set sdk version want: %d got: %d", m.SdkVersion, p.SdkVersion)
		}
	})
}

func TestPlugin_infoReceivedHandler_withSettings(t *testing.T) {
	t.Run("it appropriately fires a dispatch when an info message contains a settings key", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		messageType, _ := client.ClientMessageTypeString("settings")
		sType := reflect.TypeOf((*client.SettingsMessage)(nil)).Elem()
		mc.EXPECT().Dispatch(messageType, gomock.AssignableToTypeOf(sType))

		p := &Plugin{
			Id:     "test",
			client: mc,
		}

		settings := []byte(`[{"Host": "localhost"},{"Port": "1234"}]`)

		m := client.InfoMessage{
			Message:       client.Message{Type: client.MessageTypeInfo},
			Version:       "version",
			PluginVersion: 1,
			SdkVersion:    1,
			Settings:      settings,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		sut := p.infoReceivedHandler(&wg)

		sut(m)
	})
}

func TestPlugin_closePluginHandler(t *testing.T) {
	t.Run("it provides a handler to deal with plugin shutdown requests", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mc := mocks.NewMockPluginClient(ctrl)

		mc.EXPECT().Close()

		p := &Plugin{
			Id:     "test",
			client: mc,
		}

		m := client.ClosePluginMessage{
			Message:  client.Message{Type: client.MessageTypeClosePlugin},
			PluginId: "test",
		}

		sut := p.closePluginReceivedHandler()

		sut(m)
	})
}

func registrationFailureMocks(t *testing.T, id string) PluginClient {
	ctrl := gomock.NewController(t)
	mc := mocks.NewMockPluginClient(ctrl)

	// register should setup info handler to set registration info
	messageType, _ := client.ClientMessageTypeString("info")
	mc.EXPECT().AddMessageHandler(messageType, gomock.Any())

	// register should add a closePlugin handler to handle shutdowns
	messageType, _ = client.ClientMessageTypeString("closePlugin")
	mc.EXPECT().AddMessageHandler(messageType, gomock.Any())

	pairMessage := client.NewPairMessage(id)
	mc.
		EXPECT().
		SendMessage(pairMessage).
		Return(errors.New("failed to send message"))

	return mc
}

func registrationSuccessMocks(t *testing.T, id string, version string, pluginVersion int, sdkVersion int) PluginClient {
	ctrl := gomock.NewController(t)
	mc := mocks.NewMockPluginClient(ctrl)

	infoReceived := make(chan bool)

	// register should setup info handler to set registration info
	messageType, _ := client.ClientMessageTypeString("info")
	mc.
		EXPECT().
		AddMessageHandler(
			messageType,
			gomock.Any(),
		).
		DoAndReturn(func(msgType client.ClientMessageType, handler func(e interface{})) {
			// by running in a goroutine and blocking on a channel that we later close
			// we can mock the receipt of a message over the client socket
			go func() {
				<-infoReceived

				m := client.InfoMessage{
					Message:       client.Message{Type: client.MessageTypeInfo},
					Version:       version,
					PluginVersion: pluginVersion,
					SdkVersion:    sdkVersion,
				}

				handler(m)
			}()
		})

	// register should add a closePlugin handler to handle shutdowns
	messageType, _ = client.ClientMessageTypeString("closePlugin")
	mc.EXPECT().AddMessageHandler(messageType, gomock.Any())

	pairMessage := client.NewPairMessage(id)
	mc.
		EXPECT().
		SendMessage(pairMessage).
		DoAndReturn(func(m interface{}) error {
			if !reflect.DeepEqual(pairMessage, m) {
				msg := reflect.TypeOf(m).Elem().String()
				t.Errorf("incorrect pairing message sent want: &{client.pairMessage} got: ${%s}", msg)
			}

			// this mocks the fact TouchPortal has responded to our registration
			close(infoReceived)

			return nil
		})

	return mc
}
