package client

import "testing"

func TestRegisterPlugin(t *testing.T) {
	id := "TestPluginId"

	t.Run("registers a plugin", func(t *testing.T) {
		if err := RegisterPlugin(id); err != nil {
			t.Errorf("RegisterPlugin() returned error")
		}
	})
}
