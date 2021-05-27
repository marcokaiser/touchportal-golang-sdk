package plugin

import (
	"context"
	"log"
	"sync"

	"go.acpr.dev/touchportal-golang-sdk/pkg/client"
)

type Plugin struct {
	Id                 string
	TouchPortalVersion string
	SdkVersion         int
	PluginVersion      int

	settings interface{}

	stopped chan bool
	client  *client.Client
}

// NewPlugin creates, initialises and returns a TouchPortal plugin instance
func NewPlugin(ctx context.Context, id string) *Plugin {
	p := &Plugin{
		Id:      id,
		stopped: make(chan bool),
		client:  client.NewClient(),
	}

	go func() {
		p.client.Run(ctx)
		p.stopped <- true
	}()

	<-p.client.Ready()

	return p
}

// Register asks the TouchPortal plugin instance to handle the registration process
// with TouchPortal. It ensures that any settings are synced to the SDK and registers
// a handler that allows the SDK to deal with shutdown requests.
func (p *Plugin) Register() error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	p.OnInfo(func(event client.InfoMessage) {
		if event.Settings != nil {
			p.client.Dispatch("settings", client.SettingsMessage{
				Message:   client.Message{Type: "settings"},
				RawValues: event.Settings,
			})
		}

		p.TouchPortalVersion = event.Version
		p.PluginVersion = event.PluginVersion
		p.SdkVersion = event.SdkVersion

		wg.Done()
	})

	p.OnClosePlugin(func(event client.ClosePluginMessage) {
		log.Println("touchportal requested plugin shutdown. quitting...")
		p.client.Close()
	})

	err := p.client.SendMessage(client.NewPairMessage(p.Id))
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}

func (p *Plugin) UpdateState(id string, value string) error {
	msg := client.NewStateUpdateMessage(id, value)

	return p.client.SendMessage(msg)
}

func (p *Plugin) Done() <-chan bool {
	return p.stopped
}
