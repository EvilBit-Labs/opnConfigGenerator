package faker

import (
	"strings"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/brianvoe/gofakeit/v7"
)

// fakerTimezones is a curated list; gofakeit does not ship a timezone picker
// that emits Region/City strings valid for OPNsense's schema.
var fakerTimezones = []string{
	"America/Denver",
	"America/Los_Angeles",
	"America/New_York",
	"Europe/London",
	"Europe/Berlin",
	"UTC",
}

// fakeSystem populates a model.System with schema-valid, realistic values.
//
// Hostname is derived from a domain-style token with dots collapsed to
// hyphens because the OPNsense schema validates System.Hostname with the
// `validate:"hostname"` tag (RFC 1123 label rules). Domain is a lowercased
// FQDN to satisfy `validate:"fqdn"`.
func fakeSystem(f *gofakeit.Faker) model.System {
	host := strings.ToLower(f.DomainName())
	host = strings.ReplaceAll(host, ".", "-")

	domain := strings.ToLower(f.DomainName())

	tz := fakerTimezones[f.IntRange(0, len(fakerTimezones)-1)]

	return model.System{
		Hostname:    host,
		Domain:      domain,
		Timezone:    tz,
		Language:    "en_US",
		DNSServers:  []string{"1.1.1.1", "9.9.9.9"},
		TimeServers: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
	}
}
