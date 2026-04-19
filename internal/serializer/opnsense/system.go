package opnsense

import (
	"strings"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// SerializeSystem maps model.System onto opnsense.System. Multi-value DNS
// and NTP server lists collapse to space-separated strings because OPNsense
// stores them that way at the XML level.
//
// WebGUI.Protocol and SSH.Group are set to schema-required defaults so the
// emitted config validates under opnsense.System's struct tags.
func SerializeSystem(sys model.System) opnsense.System {
	return opnsense.System{
		Hostname:    sys.Hostname,
		Domain:      sys.Domain,
		Timezone:    sys.Timezone,
		Language:    sys.Language,
		DNSServer:   strings.Join(sys.DNSServers, " "),
		TimeServers: strings.Join(sys.TimeServers, " "),
		WebGUI: opnsense.WebGUIConfig{
			Protocol: "https",
		},
		SSH: opnsense.SSHConfig{
			Group: "wheel",
		},
	}
}
