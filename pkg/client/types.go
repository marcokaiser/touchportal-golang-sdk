//go:generate enumer -type=ClientMessageType -json -transform=lower-camel -output types_string.go  -trimprefix MessageType

package client

import "encoding/json"

type ClientMessageType int

const (
	MessageTypeAction ClientMessageType = iota
	MessageTypeClosePlugin
	MessageTypeInfo
	MessageTypePair
	MessageTypeSettings
	MessageTypeStateUpdate
)

type Message struct {
	Type ClientMessageType `json:"type"`
}

type ActionMessage struct {
	Message
	PluginId string          `json:"pluginId"`
	ActionId string          `json:"actionId"`
	Data     json.RawMessage `json:"data"`
}

type ClosePluginMessage struct {
	Message
	PluginId string `json:"pluginId"`
}

type InfoMessage struct {
	Message
	Version       string          `json:"tpVersionString"`
	VersionCode   int             `json:"tpVersionCode"`
	SdkVersion    int             `json:"sdkVersion"`
	PluginVersion int             `json:"pluginVersion"`
	Settings      json.RawMessage `json:"settings"`
}

type pairMessage struct {
	Message
	Id string `json:"id"`
}

type SettingsMessage struct {
	Message
	RawValues json.RawMessage `json:"values"`
	Values    map[string]interface{}
}

type stateUpdateMessage struct {
	Message
	Id    string `json:"id"`
	Value string `json:"value"`
}

// NewPairMessage provides a ready to go client.pairMessage that can be sent to
// TouchPortal as a part of the plugin registration flow.
func NewPairMessage(id string) *pairMessage {
	return &pairMessage{
		Message: Message{Type: MessageTypePair},
		Id:      id,
	}
}

// NewStateUpdateMessage provides a ready to go client.stateUpdateMessage that can be sent to
// TouchPortal.
func NewStateUpdateMessage(id string, value string) *stateUpdateMessage {
	return &stateUpdateMessage{
		Message: Message{Type: MessageTypeStateUpdate},
		Id:      id,
		Value:   value,
	}
}
