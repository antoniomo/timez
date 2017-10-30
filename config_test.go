package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	z := &mockZone{}
	z.On("Zone").Return(time.UTC)

	f := bytes.NewBuffer([]byte(`
default: Pacific/Auckland
freya: UTC
thor: asgard
`))

	var x = make(map[string]string)
	for k, v := range tzAlias {
		x[k] = v
	}
	x["default"] = "Pacific/Auckland"
	x["freya"] = "UTC"
	x["thor"] = "asgard"

	expectedConfig := config{
		localTZ: time.UTC,
		aliases: x,
	}
	actualConfig := mustLoadConfig(z, f)
	assert.Equal(t, expectedConfig, actualConfig)
}
