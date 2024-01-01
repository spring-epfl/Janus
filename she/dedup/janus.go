package dedup

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type JanusParams struct {
	// Split each template between $CtxPerTemplate ciphertexts where each ciphertext contains
	// $SlotsPerCtx slots of the template
	// The template size must be equal to $CtxPerTemplate * $SlotsPerCtx
	// The number of slots in a ciphertext must be a power of 2
	CtxPerTemplate int
	SlotsPerCtx    int
	DbSize         int
	Nbfv           int // number of slots in a ciphertext

	// sensor params
	BioType       string
	TemplateSize  int
	SensorD       int64
	SensorHasMask bool
}

func (bio JanusParams) Describe() string {
	return fmt.Sprintf("DB[%v] templates from %v Sensor(%v, %v).\n",
		bio.DbSize, bio.BioType, bio.TemplateSize, bio.SensorD)
}

// The RS component of Hyb-Janus
// This component only include the SHE distance computation portion of Hyb-Janus
// To check the SMC thresholding portion of Hyb-Janus, check smc/bio_dedup/hyb_threshold.cpp
type Janus struct {
	Params *JanusParams
	HE     *HEHandler

	db    []*PlainBio
	encDB *EncryptedDB
}

type EncryptedDB struct {
	bioType string

	fingerCtxStrip []*CtxStrip

	irisMaskCtxStrip     []*CtxStrip
	irisYMaskCtxStrip    []*CtxStrip
	irisYBarMaskCtxStrip []*CtxStrip
}

// Generates random data for the user template database
func (janus *Janus) GenerateUserDB() {
	janus.db = make([]*PlainBio, janus.Params.DbSize)
	for i := range janus.db {
		janus.db[i] = NewRandomPlainBio(janus.Params)
	}
}

// Returns a bio template that matches user db[matchIdx]
func (janus *Janus) GenerateMatchingQuery(matchIdx int) *PlainBio {
	return janus.db[matchIdx].CreateFakeMatch(0.9)
}

// Compute the identification distance using the plain database (ground truth)
func (janus *Janus) IdentificationGroundTruth(query *PlainBio) (answer []int64) {
	answer = make([]int64, janus.Params.DbSize)
	for i := range answer {
		answer[i] = query.ComputeDist(janus.db[i])
	}
	return answer
}

func (janus *Janus) EncryptDatabase() error {
	if janus.Params.BioType == "finger" {
		return janus.EncryptFingerDatabase()
	} else if janus.Params.BioType == "iris" {
		return janus.EncryptIrisDatabase()
	} else {
		fmt.Printf("BioType %v not supported.\n", janus.Params.BioType)
		return fmt.Errorf("BioType %v not supported.\n", janus.Params.BioType)
	}
}
func (janus *Janus) EncryptFingerDatabase() error {
	// plaintext strips
	records := make([][]int64, len(janus.db))
	for i := 0; i < len(janus.db); i++ {
		records[i] = janus.db[i].Data
	}
	dbPlainStrips, err := StripRecords(janus.Params, records)
	if err != nil {
		return fmt.Errorf("DB packing failed: %v.\n", err)
	}

	// encrypt strips
	dbCtxStrips := make([]*CtxStrip, len(dbPlainStrips))
	for i := 0; i < len(dbPlainStrips); i++ {
		dbCtxStrips[i], err = dbPlainStrips[i].Encrypt(janus.HE)
		if err != nil {
			return fmt.Errorf("DB encryption error: %v.\n", err)
		}
	}
	janus.encDB = &EncryptedDB{
		bioType:        "finger",
		fingerCtxStrip: dbCtxStrips,
	}
	return nil
}

func (janus *Janus) EncryptIrisDatabase() error {
	records_y := make([][]int64, len(janus.db))
	records_mask := make([][]int64, len(janus.db))
	for i := 0; i < len(janus.db); i++ {
		records_y[i] = janus.db[i].Data
		records_mask[i] = janus.db[i].Mask
	}

	yStrips, err := StripRecords(janus.Params, records_y)
	if err != nil {
		return fmt.Errorf("DB packing y failed: %v.\n", err)
	}

	maskStrips, err := StripRecords(janus.Params, records_mask)
	if err != nil {
		return fmt.Errorf("DB packing mask failed: %v.\n", err)
	}

	// encrypt strips
	irisMaskCtxStrip := make([]*CtxStrip, len(yStrips))
	irisYMaskCtxStrip := make([]*CtxStrip, len(yStrips))
	irisYBarMaskCtxStrip := make([]*CtxStrip, len(yStrips))
	for i := 0; i < len(yStrips); i++ {
		irisMaskCtxStrip[i], err = maskStrips[i].Encrypt(janus.HE)
		if err != nil {
			return fmt.Errorf("DB encryption error mask: %v.\n", err)
		}
		irisYMaskCtxStrip[i], err = (StripMul(maskStrips[i], yStrips[i])).Encrypt(janus.HE)
		if err != nil {
			return fmt.Errorf("DB encryption error Y.Mask: %v.\n", err)
		}
		irisYBarMaskCtxStrip[i], err = (StripMul(maskStrips[i], yStrips[i].LogicNot())).Encrypt(janus.HE)
		if err != nil {
			return fmt.Errorf("DB encryption error Ybar.mask: %v.\n", err)
		}
	}
	janus.encDB = &EncryptedDB{
		bioType:              "iris",
		irisMaskCtxStrip:     irisMaskCtxStrip,
		irisYMaskCtxStrip:    irisYMaskCtxStrip,
		irisYBarMaskCtxStrip: irisYBarMaskCtxStrip,
	}
	return nil
}

// The distance computation (SHE) component of Hyb-Janus
// This function computes the distance between the query and the database in cipher domain.
// In Hyb-Janus, the registration station secret shares this encrypted distance (using additive
// secret sharing) and sends the encypted share to the biometric provider who holds the key.
func (janus *Janus) Identification(query *PlainBio) (PackedEncDist []*rlwe.Ciphertext) {
	if janus.Params.BioType == "finger" {
		return janus.ComputeEucDist(query)
	} else if janus.Params.BioType == "iris" {
		return janus.ComputeNormHamDist(query)
	} else {
		fmt.Printf("BioType %v not supported.\n", janus.Params.BioType)
		return nil
	}
}

func (janus *Janus) ComputeEucDist(query *PlainBio) (PackedEncDist []*rlwe.Ciphertext) {
	queryPtxStrip, err := ReplicateAsStripeRecords(janus.Params, query.Data)
	if err != nil {
		fmt.Printf("Query packing failed: %v.\n", err)
		return
	}
	PackedEncDist = queryPtxStrip.EuclideanIdentification(janus.HE, janus.encDB.fingerCtxStrip)
	return PackedEncDist
}

func (janus *Janus) ComputeNormHamDist(query *PlainBio) (PackedEncDist []*rlwe.Ciphertext) {
	xStrip, err := ReplicateAsStripeRecords(janus.Params, query.Data)
	if err != nil {
		fmt.Printf("Query packing data failed: %v.\n", err)
		return
	}

	maskStrip, err := ReplicateAsStripeRecords(janus.Params, query.Mask)
	if err != nil {
		fmt.Printf("Query packing mask failed: %v.\n", err)
		return
	}

	PackedEncDist = NHammingDistance(janus.HE, xStrip, maskStrip, janus.encDB.irisYMaskCtxStrip, janus.encDB.irisYBarMaskCtxStrip, janus.encDB.irisMaskCtxStrip)
	return PackedEncDist
}
