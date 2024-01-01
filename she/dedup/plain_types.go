package dedup

import (
	"math/rand"
)

const MATCH_THRESHOLD uint64 = 40

type PlainBio struct {
	BioMode string  // finger, iris
	Data    []int64 // data may be 1 byte, compatibility with lattigo
	Mask    []int64 // Mask is binary
	MaxVal  int64
	HasMask bool
}

func NewRandomPlainBio(bio *JanusParams) *PlainBio {
	fc := &PlainBio{
		Data:    make([]int64, bio.TemplateSize),
		MaxVal:  bio.SensorD,
		HasMask: bio.SensorHasMask,
	}
	for i := 0; i < bio.TemplateSize; i++ {
		fc.Data[i] = rand.Int63n(bio.SensorD)
	}

	if bio.SensorHasMask {
		fc.Mask = make([]int64, bio.TemplateSize)
		for i := 0; i < bio.TemplateSize; i++ {
			if rand.Float32() < 0.85 {
				fc.Mask[i] = 1
			}
		}
	}
	return fc
}

func (base *PlainBio) CreateFakeMatch(similarity float32) *PlainBio {
	dlen := len(base.Data)
	bio := &PlainBio{
		Data:    make([]int64, dlen),
		MaxVal:  base.MaxVal,
		HasMask: base.HasMask,
	}

	for i := 0; i < dlen; i++ {
		bio.Data[i] = base.Data[i]
		if rand.Float32() > similarity {
			bio.Data[i] = (1 - bio.Data[i] + bio.MaxVal) % bio.MaxVal
		}
	}

	if bio.HasMask {
		bio.Mask = make([]int64, dlen)
		for i := 0; i < dlen; i++ {
			if rand.Float32() < 0.85 {
				bio.Mask[i] = 1
			}
		}
	}
	return bio
}

func (base *PlainBio) ComputeDist(target *PlainBio) (dist int64) {
	if base.HasMask && target.HasMask {
		mask := MergeMask(base.Mask, target.Mask)
		return DistEuclidean(
			ApplyMask(base.Data, mask),
			ApplyMask(target.Data, mask),
		)
	} else {
		return DistEuclidean(base.Data, target.Data)
	}
}

func (base *PlainBio) Match(target *PlainBio, threshold int64) bool {
	return base.ComputeDist(target) <= threshold
}
