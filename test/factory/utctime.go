package factory

import (
	"math/rand"

	random "github.com/brianvoe/gofakeit/v5"
	"github.com/crypto-com/chain-indexing/external/utctime"
)

func addUTCTimeFuncLookup() {
	random.AddFuncLookup("utctime", random.Info{
		Category:    "custom",
		Description: "Random time.Time",
		Example:     "0",
		Output:      "utctime.UTCTime",
		Call: func(m *map[string][]string, info *random.Info) (interface{}, error) {
			return RandomUTCTime(), nil
		},
	})
}

func RandomUTCTime() utctime.UTCTime {
	// nolint:gosec
	return utctime.FromUnixNano(rand.Int63())
}
