// Package plugin provides an event based sdk of the TouchPortal api.
package plugin

import (
	"context"
	"log"
	"sync"

	"go.acpr.dev/touchportal-golang-sdk/pkg/client"
)

//go:generate mockgen -destination=../mocks/plugin_mocks.go -package=mocks . PluginClient
type PluginClient interface {
	AddMessageHandler(client.ClientMessageType, func(e interface{}))
	Close()
	Dispatch(client.ClientMessageType, interface{})
	Ready() <-chan bool
	Run(context.Context)
	SendMessage(interface{}) error
}

type Plugin struct {
	Id                 string
	TouchPortalVersion string
	SdkVersion         int
	PluginVersion      int

	settings interface{}

	done   chan bool
	client PluginClient
}

// NewPlugin creates, initialises and returns a TouchPortal plugin instance
func NewPlugin(ctx context.Context, id string) *Plugin {
	return NewPluginWithClient(ctx, client.NewClient(), id)
}

// NewPluginWithClient creates, initialises and returns a TouchPortal plugin instance allowing
// the usage of a custom client instance
func NewPluginWithClient(ctx context.Context, cli PluginClient, id string) *Plugin {
	p := &Plugin{
		Id:     id,
		done:   make(chan bool),
		client: cli,
	}

	go func() {
		p.client.Run(ctx)
		p.done <- true
	}()

	// wait until client is ready to be used
	<-p.client.Ready()

	return p
}

// Register asks the TouchPortal plugin instance to handle the registration process
// with TouchPortal. It ensures that any settings are synced to the SDK and registers
// a handler that allows the SDK to deal with shutdown requests.
func (p *Plugin) Register() error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	p.OnInfo(p.infoReceivedHandler(&wg))

	p.OnClosePlugin(p.closePluginReceivedHandler())

	err := p.client.SendMessage(client.NewPairMessage(p.Id))
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}

// UpdateState allows you to send state update messages to TouchPortal
func (p *Plugin) UpdateState(id string, value string) error {
	msg := client.NewStateUpdateMessage(id, value)

	return p.client.SendMessage(msg)
}

// Done provides an unbuffered, blocking, channel that can be used to verify
// that the Plugin has finished it's run and cleaned up used resources.
func (p *Plugin) Done() <-chan bool {
	return p.done
}

func (p *Plugin) infoReceivedHandler(wg *sync.WaitGroup) func(event client.InfoMessage) {
	return func(event client.InfoMessage) {
		if event.Settings != nil {
			p.client.Dispatch(client.MessageTypeSettings, client.SettingsMessage{
				Message:   client.Message{Type: client.MessageTypeSettings},
				RawValues: event.Settings,
			})
		}

		p.TouchPortalVersion = event.Version
		p.PluginVersion = event.PluginVersion
		p.SdkVersion = event.SdkVersion

		wg.Done()
	}
}

func (p *Plugin) closePluginReceivedHandler() func(event client.ClosePluginMessage) {
	return func(event client.ClosePluginMessage) {
		log.Println("touchportal requested plugin shutdown. quitting...")
		p.client.Close()
	}
}
