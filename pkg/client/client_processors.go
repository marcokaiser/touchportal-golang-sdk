package client

import (
	"encoding/json"
)

func (c *Client) SetMessageProcessor(msgType ClientMessageType, processor func(msg json.RawMessage) (interface{}, error)) {
	c.processors[msgType] = processor
}

func (c *Client) registerDefaultMessageProcessors() {
	c.SetMessageProcessor(MessageTypeAction, actionMessageProcessor)
	c.SetMessageProcessor(MessageTypeClosePlugin, closePluginProcessor)
	c.SetMessageProcessor(MessageTypeInfo, infoMessageProcessor)
	c.SetMessageProcessor(MessageTypeSettings, settingsMessageProcessor)
}

func actionMessageProcessor(msg json.RawMessage) (interface{}, error) {
	var pm ActionMessage
	err := json.Unmarshal(msg, &pm)

	return pm, err
}

func closePluginProcessor(msg json.RawMessage) (interface{}, error) {
	var pm ClosePluginMessage
	err := json.Unmarshal(msg, &pm)

	return pm, err
}

func infoMessageProcessor(msg json.RawMessage) (interface{}, error) {
	var pm InfoMessage
	err := json.Unmarshal(msg, &pm)

	return pm, err
}

func settingsMessageProcessor(msg json.RawMessage) (interface{}, error) {
	var pm SettingsMessage
	err := json.Unmarshal(msg, &pm)

	return pm, err
}
