// Code generated by "enumer -type=pluginEvent -json -transform=lower-camel -output plugin_events_string.go -trimprefix event"; DO NOT EDIT.

//
package plugin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const _pluginEventName = "actionclosePlugininfosettings"

var _pluginEventIndex = [...]uint8{0, 6, 17, 21, 29}

func (i pluginEvent) String() string {
	if i < 0 || i >= pluginEvent(len(_pluginEventIndex)-1) {
		return fmt.Sprintf("pluginEvent(%d)", i)
	}
	return _pluginEventName[_pluginEventIndex[i]:_pluginEventIndex[i+1]]
}

var _pluginEventValues = []pluginEvent{0, 1, 2, 3}

var _pluginEventNames = []string{"action", "closePlugin", "info", "settings"}

var _pluginEventNameToValueMap = map[string]pluginEvent{
	_pluginEventName[0:6]:   0,
	_pluginEventName[6:17]:  1,
	_pluginEventName[17:21]: 2,
	_pluginEventName[21:29]: 3,
}

// pluginEventString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func pluginEventString(s string) (pluginEvent, error) {

	if val, ok := _pluginEventNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to pluginEvent values", s)
}

func ParsepluginEvent(s string) (pluginEvent, error) {
	return pluginEventString(s)
}

// pluginEventValues returns all values of the enum
func pluginEventValues() []pluginEvent {
	return _pluginEventValues
}

func pluginEventNames() []string {
	return _pluginEventNames
}

// IsApluginEvent returns "true" if the value is listed in the enum definition. "false" otherwise
func (i pluginEvent) IsApluginEvent() bool {
	for _, v := range _pluginEventValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for pluginEvent
func (i pluginEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for pluginEvent
func (i *pluginEvent) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("pluginEvent should be a string, got %s", data)
	}

	var err error
	*i, err = pluginEventString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for pluginEvent
func (i pluginEvent) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for pluginEvent
func (i *pluginEvent) UnmarshalText(text []byte) error {
	var err error
	*i, err = pluginEventString(string(text))
	return err
}

// MarshalYAML implements a YAML Marshaler for pluginEvent
func (i pluginEvent) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// UnmarshalYAML implements a YAML Unmarshaler for pluginEvent
func (i *pluginEvent) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var err error
	*i, err = pluginEventString(s)
	return err
}

func (i pluginEvent) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *pluginEvent) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("value is not a byte slice")
		}

		str = string(bytes[:])
	}

	val, err := pluginEventString(str)
	if err != nil {
		return err
	}

	*i = val
	return nil
}
