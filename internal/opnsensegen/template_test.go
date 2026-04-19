package opnsensegen_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
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

// countingWriter records the number of Write calls so we can verify the
// atomic-write contract: MarshalConfig buffers the full document in memory
// and performs exactly one Write when encode/stabilize succeed.
type countingWriter struct {
	writes int
	err    error
	buf    bytes.Buffer
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.err != nil {
		return 0, w.err
	}
	return w.buf.Write(p)
}

func TestMarshalConfigIsAtomicOnSuccess(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	w := &countingWriter{}
	require.NoError(t, opnsensegen.MarshalConfig(cfg, w))

	assert.Equal(t, 1, w.writes, "MarshalConfig must perform exactly one Write on success")
}

func TestMarshalConfigDoesNotWriteOnWriterFailure(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	sentinel := errors.New("disk full")
	w := &countingWriter{err: sentinel}
	err = opnsensegen.MarshalConfig(cfg, w)

	require.ErrorIs(t, err, sentinel)
	// The writer saw exactly one attempt — no header-first / body-second
	// partial write pattern.
	assert.Equal(t, 1, w.writes)
	assert.Zero(t, w.buf.Len(), "failing writer must not accumulate partial output")
}

// TestMarshalConfigSortsMapBackedSections exercises the token-stream
// stabilizer by constructing a document with children that would iterate
// in non-alphabetical order under Go's randomized map iteration, then
// asserts the output is ordered.
func TestMarshalConfigSortsMapBackedSections(t *testing.T) {
	t.Parallel()

	cfg := &opnsense.OpnSenseDocument{
		Version: "1.0",
		System:  opnsense.System{Hostname: "test", Domain: "test.local"},
		Interfaces: opnsense.Interfaces{
			Items: map[string]opnsense.Interface{
				"zeta":  {If: "igb0", Enable: "1"},
				"alpha": {If: "igb1", Enable: "1"},
				"mu":    {If: "igb2", Enable: "1"},
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(cfg, &buf))

	out := buf.String()
	iAlpha := strings.Index(out, "<alpha>")
	iMu := strings.Index(out, "<mu>")
	iZeta := strings.Index(out, "<zeta>")
	require.NotEqual(t, -1, iAlpha, "alpha must appear")
	require.NotEqual(t, -1, iMu, "mu must appear")
	require.NotEqual(t, -1, iZeta, "zeta must appear")
	assert.Less(t, iAlpha, iMu, "alpha must appear before mu")
	assert.Less(t, iMu, iZeta, "mu must appear before zeta")
}

func TestMarshalConfigHandlesEmptyMapBackedSections(t *testing.T) {
	t.Parallel()

	cfg := &opnsense.OpnSenseDocument{
		Version:    "1.0",
		System:     opnsense.System{Hostname: "test", Domain: "test.local"},
		Interfaces: opnsense.Interfaces{Items: map[string]opnsense.Interface{}},
		Dhcpd:      opnsense.Dhcpd{Items: map[string]opnsense.DhcpdInterface{}},
	}

	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(cfg, &buf))

	out := buf.String()
	// Empty map-backed sections must round-trip without error and must
	// still be well-formed XML.
	assert.Contains(t, out, "<opnsense>")
	assert.Contains(t, out, "</opnsense>")
}

// TestMarshalConfigByteStableMapIteration runs MarshalConfig 20 times on the
// same input. Go's map iteration is randomized per encode, so without the
// sort post-processor, iterations diverge. 20 is high enough to defeat
// randomization luck.
func TestMarshalConfigByteStableMapIteration(t *testing.T) {
	t.Parallel()

	cfg := &opnsense.OpnSenseDocument{
		Version: "1.0",
		System:  opnsense.System{Hostname: "test", Domain: "test.local"},
		Interfaces: opnsense.Interfaces{
			Items: map[string]opnsense.Interface{
				"wan":  {If: "igb0", Enable: "1", IPAddr: "dhcp"},
				"lan":  {If: "igb1", Enable: "1", IPAddr: "192.168.1.1", Subnet: "24"},
				"opt1": {If: "igb2", Enable: "1", IPAddr: "10.0.0.1", Subnet: "24"},
				"opt2": {If: "igb3", Enable: "1", IPAddr: "10.0.1.1", Subnet: "24"},
				"opt3": {If: "igb4", Enable: "1", IPAddr: "10.0.2.1", Subnet: "24"},
			},
		},
	}

	var first bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(cfg, &first))

	for i := range 20 {
		var next bytes.Buffer
		require.NoError(t, opnsensegen.MarshalConfig(cfg, &next))
		require.Equalf(t, first.Bytes(), next.Bytes(), "iteration %d diverged", i+1)
	}
}
