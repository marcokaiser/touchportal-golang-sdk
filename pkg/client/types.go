package client

import "encoding/json"

type Message struct {
	Type string `json:"type"`
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

type PairMessage struct {
	Message
	Id string `json:"id"`
}

type SettingsMessage struct {
	Message
	RawValues json.RawMessage `json:"values"`
	Values    map[string]interface{}
}
type StateUpdateMessage struct {
	Message
	Id    string `json:"id"`
	Value string `json:"value"`
}

func NewPairMessage(id string) *PairMessage {
	return &PairMessage{
		Message: Message{Type: "pair"},
		Id:      id,
	}
}

func NewStateUpdateMessage(id string, value string) *StateUpdateMessage {
	return &StateUpdateMessage{
		Message: Message{Type: "stateUpdate"},
		Id:      id,
		Value:   value,
	}
}
