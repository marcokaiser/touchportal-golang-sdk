package client

import "encoding/json"

type Message struct {
	Type string `json:"type"`
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

func NewPairMessage(id string) *PairMessage {
	m := &PairMessage{}
	m.Type = "pair"
	m.Id = id

	return m
}
