//go:generate mockgen -source=plugin.go -destination=mocks/plugin.go -mock_names pluginClient=MockPluginClient pluginClient

package plugin

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.acpr.dev/touchportal-golang-sdk/client"
	. "go.acpr.dev/touchportal-golang-sdk/plugin/mocks"
)

func TestNewPluginWithClient(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	ctx := context.Background()

	mc.EXPECT().Run(ctx)

	ready := make(chan bool)
	go func(ready chan bool) {
		// we need to wait till marking the client ready so the
		// goroutine has a chance to process the Run expectation
		time.Sleep(time.Microsecond * 100)
		ready <- true
	}(ready)
	mc.EXPECT().Ready().Return(ready)

	sut := NewPluginWithClient(ctx, mc, "test")
	assert.Equal(t, "test", sut.ID)
}

func TestPlugin_UpdateState(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	mc.
		EXPECT().
		SendMessage(
			client.NewStateUpdateMessage("testState", "testStateValue"),
		).
		Return(nil)

	p := &Plugin{
		ID:     "test",
		client: mc,
	}

	err := p.UpdateState("testState", "testStateValue")
	assert.Nil(t, err, "failed to update state err: %v", err)
}

func TestPlugin_Done(t *testing.T) {
	t.Parallel()

	p := &Plugin{
		done: make(chan bool),
	}

	wait := make(chan bool)
	go func(wait chan bool) {
		defer close(wait)

		// the actual function call we're testing
		<-p.Done()
	}(wait)

	// something changes the stop status inside the plugin
	p.done <- true

	select {
	case <-wait:
	case <-time.After(100 * time.Millisecond):
		t.Error("plugin not stopped before timeout")
	}
}

func TestPlugin_Register(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID                 string
		TouchPortalVersion string
		SdkVersion         int
		PluginVersion      int

		Client pluginClient
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "it successfully registers a plugin instance",
			fields: fields{
				ID:                 "test",
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
				ID:     "test",
				Client: registrationFailureMocks(t, "test"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &Plugin{
				ID:     tt.fields.ID,
				client: tt.fields.Client,
			}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := p.Register()
				assert.Equal(t, tt.wantErr, (err != nil), "plugin register failed to match expected result")
			}()
			wg.Wait()
		})
	}
}

func TestPlugin_infoReceivedHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	p := &Plugin{
		ID:     "test",
		client: mc,
	}

	m := client.InfoMessage{
		Message:       client.Message{Type: client.MessageTypeInfo},
		Version:       "version",
		PluginVersion: 1,
		SdkVersion:    3,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	sut := p.infoReceivedHandler(&wg)

	sut(m)

	assert.Equal(t, m.Version, p.TouchPortalVersion,
		"Failed to set TouchPortal version want %s got %s",
		m.Version,
		p.TouchPortalVersion,
	)

	assert.Equal(t, m.PluginVersion, p.PluginVersion,
		"Failed to set plugin version want %d got %d",
		m.PluginVersion,
		p.PluginVersion,
	)

	assert.Equal(t, m.SdkVersion, p.SdkVersion,
		"Failed to set sdk version want %d got %d",
		m.SdkVersion,
		p.SdkVersion,
	)
}

func TestPlugin_infoReceivedHandler_withSettings(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	messageType, _ := client.ClientMessageTypeString("settings")
	sType := reflect.TypeOf((*client.SettingsMessage)(nil)).Elem()
	mc.EXPECT().Dispatch(messageType, gomock.AssignableToTypeOf(sType))

	p := &Plugin{
		ID:     "test",
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
}

func TestPlugin_closePluginHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

	mc.EXPECT().Close()

	p := &Plugin{
		ID:     "test",
		client: mc,
	}

	m := client.ClosePluginMessage{
		Message:  client.Message{Type: client.MessageTypeClosePlugin},
		PluginID: "test",
	}

	sut := p.closePluginReceivedHandler()

	sut(m)
}

func registrationFailureMocks(t *testing.T, id string) pluginClient {
	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

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

func registrationSuccessMocks(t *testing.T, id string, version string, pluginVersion int, sdkVersion int) pluginClient {
	ctrl := gomock.NewController(t)
	mc := NewMockPluginClient(ctrl)

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
			assert.IsType(t, pairMessage, m, "incorrect messsage type sent for pairing")

			// this mocks the fact TouchPortal has responded to our registration
			close(infoReceived)

			return nil
		})

	return mc
}
