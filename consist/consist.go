package consist

import (
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
)

func init() {
	cfg := consistent.Config{
		PartitionCount:    7,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            hasher{},
	}
	Consist = consistent.New(nil, cfg)
}

type Member string

func (m Member) String() string {
	return string(m)
}

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

var Consist *consistent.Consistent
