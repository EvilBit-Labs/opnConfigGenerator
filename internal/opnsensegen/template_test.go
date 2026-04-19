package opnsensegen_test

import (
	"bytes"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBaseConfig(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "opnsense", cfg.System.Hostname)
	assert.Equal(t, "localdomain", cfg.System.Domain)
}

func TestLoadBaseConfigMissingFile(t *testing.T) {
	t.Parallel()

	_, err := opnsensegen.LoadBaseConfig("does-not-exist.xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read base config")
}

func TestParseConfig(t *testing.T) {
	t.Parallel()

	xmlData := []byte(`<?xml version="1.0"?>
<opnsense>
  <version>1.0</version>
  <system>
    <hostname>test</hostname>
    <domain>test.local</domain>
  </system>
  <vlans/>
  <interfaces/>
  <dhcpd/>
  <filter/>
</opnsense>`)

	cfg, err := opnsensegen.ParseConfig(xmlData)
	require.NoError(t, err)
	assert.Equal(t, "test", cfg.System.Hostname)
	assert.Equal(t, "test.local", cfg.System.Domain)
}

func TestParseConfigInvalidXML(t *testing.T) {
	t.Parallel()

	_, err := opnsensegen.ParseConfig([]byte("not xml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse config XML")
}

func TestMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(cfg, &buf))

	output := buf.String()
	assert.Contains(t, output, "<?xml")
	assert.Contains(t, output, "<opnsense>")
	assert.Contains(t, output, "<hostname>opnsense</hostname>")
	assert.Contains(t, output, "</opnsense>")
}
