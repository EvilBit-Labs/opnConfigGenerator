// Package faker produces realistic *model.CommonDevice values for the
// serializer pipeline. This package has no knowledge of XML or the opnsense
// schema; it only populates the opnDossier CommonDevice model.
package faker

import (
	"math/rand/v2"

	"github.com/brianvoe/gofakeit/v7"
)

// newRand builds a *rand.Rand and a *gofakeit.Faker sharing the same stream.
// See WithSeed for the seed == 0 sentinel semantics.
func newRand(seed int64) (*rand.Rand, *gofakeit.Faker) {
	var rng *rand.Rand
	if seed == 0 {
		//nolint:gosec // Fake data generation; not security-sensitive.
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	} else {
		//nolint:gosec // Fake data generation; not security-sensitive.
		rng = rand.New(rand.NewPCG(uint64(seed), 0))
	}
	return rng, gofakeit.NewFaker(rng, false)
}
