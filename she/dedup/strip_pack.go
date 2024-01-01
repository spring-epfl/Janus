package dedup

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var VERBOSE = false

// Stripe packing
// We represent a template of size TS as a 2D array of dimensions CtxPerTemplate x SlotPerCtx
// We batch (N/SlotPerCtx) templates into CtxPerTemplate ciphertexts.
// Adding ciphertext together is one of the cheapest operations in BFV. Having a small SlotPerCtx
// leads to a lower number of BFV.rotations (needed to compute the inner sum) at
// the cost of having more ciphertext and cipher addition which is cheap.

type PlainStrip struct {
	CtxPerTemplate int // number of ciphertexts, first dimension of Strips
	SlotPerCtx     int // number of batched elements from each record in a ciphertext, second dimension of Strips, must be a power of 2 ($SlotPerCtx <= N/2)
	// The template size must be equal to $CtxPerTemplate*$SlotPerCtx
	RecPerCtx int // number of records per ciphertext, N = SlotPerCtx * recPerCtx

	Strips    [][]int64
	PtxStrips []*rlwe.Plaintext
}

type CtxStrip struct {
	CtxPerTemplate int // number of ciphertexts, first dimension of Strips
	SlotPerCtx     int // number of batched elements from each record in a ciphertext, second dimension of Strips, must be a power of 2
	RecPerCtx      int // number of records per ciphertext, N = BN * recPerCtx

	Strips []*rlwe.Ciphertext
}

// Takes a PlainStrip encoding RecPerCtx templates and return the $idx'th template
func (strip *PlainStrip) GetRecord(idx int) []int64 {
	out := make([]int64, strip.CtxPerTemplate*strip.SlotPerCtx)
	for i := 0; i < strip.CtxPerTemplate; i++ {
		for j := 0; j < strip.SlotPerCtx; j++ {
			out[i*strip.SlotPerCtx+j] = strip.Strips[i][idx*strip.SlotPerCtx+j]
		}
	}
	return out
}

// Takes m records of size TS where TS = CN*BN and stripes them into m/recPerCtx PlainStrips
func StripRecords(params *JanusParams, records [][]int64) ([]*PlainStrip, error) {
	if len(records[0]) != params.CtxPerTemplate*params.SlotsPerCtx {
		return nil, fmt.Errorf("stripeRecords: mismatching parameters. TS(%v) != CtxPerTemplate(%v)*SlotsPerCtx(%v)", len(records[0]), params.CtxPerTemplate, params.SlotsPerCtx)
	}

	if VERBOSE {
		fmt.Println("StripeRecords:")
		for i := 0; i < 4; i++ {
			fmt.Printf("Input rec[%v]: %v\n", i, records[i][0:10])
		}
	}

	recPerCtx := params.Nbfv / params.SlotsPerCtx
	// number of stripes needed to store the records
	stripeNum := (len(records) + int(recPerCtx) - 1) / int(recPerCtx)
	outStrips := make([]*PlainStrip, stripeNum)
	for st := 0; st < stripeNum; st++ {
		strip := &PlainStrip{
			Strips:         make([][]int64, params.CtxPerTemplate),
			CtxPerTemplate: params.CtxPerTemplate,
			SlotPerCtx:     params.SlotsPerCtx,
			RecPerCtx:      recPerCtx,
		}

		stEnd := (st + 1) * recPerCtx
		if stEnd > len(records) {
			stEnd = len(records)
		}
		activeRecs := records[st*recPerCtx : stEnd]

		for i := 0; i < params.CtxPerTemplate; i++ { // for each ciphertext
			strip.Strips[i] = make([]int64, params.Nbfv)
			for rec := 0; rec < len(activeRecs); rec++ { // for each record from this strip batch
				for j := 0; j < params.SlotsPerCtx; j++ { // for each batched element
					strip.Strips[i][rec*params.SlotsPerCtx+j] = activeRecs[rec][i*params.SlotsPerCtx+j]
				}
			}

			if VERBOSE {
				if i < 4 {
					fmt.Printf("Output strip[0], ctx[%v]: %v\n", i, strip.Strips[i][0:20])
				}
			}
		}
		outStrips[st] = strip
	}
	return outStrips, nil
}

// Replicate a single template (N/slotPerCtx elements) into a PlainStrip
func ReplicateAsStripeRecords(params *JanusParams, record []int64) (*PlainStrip, error) {
	if len(record) != params.CtxPerTemplate*params.SlotsPerCtx {
		return nil, fmt.Errorf("stripeRecords: mismatching parameters. TS(%v) != CtxPerTemplate(%v)*SlotsPerCtx(%v)", len(record), params.CtxPerTemplate, params.SlotsPerCtx)
	}
	recPerCtx := params.Nbfv / params.SlotsPerCtx

	replicates := make([][]int64, recPerCtx)
	for i := 0; i < recPerCtx; i++ {
		replicates[i] = record
	}

	strip, err := StripRecords(params, replicates)
	if len(strip) > 1 {
		return nil, fmt.Errorf("ReplicateAsStripeRecords: failed")
	}

	return strip[0], err
}

func (base *PlainStrip) EnsurePtxStripe(HE *HEHandler) {
	if base.PtxStrips != nil {
		return
	}
	base.PtxStrips = make([]*rlwe.Plaintext, base.CtxPerTemplate)
	for i := 0; i < base.CtxPerTemplate; i++ {
		base.PtxStrips[i] = HE.Encoder.EncodeNew(base.Strips[i], HE.Params.MaxLevel())
	}
}

func (base *PlainStrip) Encrypt(HE *HEHandler) (*CtxStrip, error) {
	out := &CtxStrip{
		CtxPerTemplate: base.CtxPerTemplate,
		SlotPerCtx:     base.SlotPerCtx,
		RecPerCtx:      base.RecPerCtx,
		Strips:         make([]*rlwe.Ciphertext, base.CtxPerTemplate),
	}

	base.EnsurePtxStripe(HE)
	for i := 0; i < base.CtxPerTemplate; i++ {
		out.Strips[i] = HE.Encryptor.EncryptNew(base.PtxStrips[i])
	}

	return out, nil
}

func (base *CtxStrip) Square(HE *HEHandler) {
	for i := 0; i < int(base.CtxPerTemplate); i++ {
		base.Strips[i] = HE.Evaluator.MulNew(base.Strips[i], base.Strips[i])
		HE.Evaluator.Relinearize(base.Strips[i], base.Strips[i])
	}

}

func (base *PlainStrip) LogicNot() *PlainStrip {
	out := &PlainStrip{
		CtxPerTemplate: base.CtxPerTemplate,
		SlotPerCtx:     base.SlotPerCtx,
		RecPerCtx:      base.RecPerCtx,
		Strips:         make([][]int64, base.CtxPerTemplate),
	}
	for i := 0; i < base.CtxPerTemplate; i++ {
		out.Strips[i] = make([]int64, len(base.Strips[i]))
		for j := 0; j < len(base.Strips[i]); j++ {
			out.Strips[i][j] = 1 - base.Strips[i][j]
		}
	}
	return out
}

func StripMul(a, b *PlainStrip) *PlainStrip {
	out := &PlainStrip{
		CtxPerTemplate: a.CtxPerTemplate,
		SlotPerCtx:     a.SlotPerCtx,
		RecPerCtx:      a.RecPerCtx,
		Strips:         make([][]int64, a.CtxPerTemplate),
	}

	for i := 0; i < a.CtxPerTemplate; i++ {
		out.Strips[i] = make([]int64, len(a.Strips[i]))
		for j := 0; j < len(a.Strips[i]); j++ {
			out.Strips[i][j] = a.Strips[i][j] * b.Strips[i][j]
		}
	}

	return out
}

func (base *CtxStrip) Sub(HE *HEHandler, target *PlainStrip) {

	target.EnsurePtxStripe(HE)
	for i := 0; i < int(base.CtxPerTemplate); i++ {
		HE.Evaluator.Sub(base.Strips[i], target.PtxStrips[i], base.Strips[i])
	}
}

func (base *CtxStrip) Mul(HE *HEHandler, target *PlainStrip) {
	target.EnsurePtxStripe(HE)
	for i := 0; i < int(base.CtxPerTemplate); i++ {
		HE.Evaluator.Mul(base.Strips[i], target.PtxStrips[i], base.Strips[i])
	}
}

// Computes the sum of all TS slots of each template
// This function randomizes internal slots that do not contain the sum values to
// prevent information leakage
func (base *CtxStrip) StripeSum(HE *HEHandler) *rlwe.Ciphertext {
	// Compute the sum of all CtxPerTemplate ciphertexts
	ctxs := base.Strips[:]
	for len(ctxs) > 1 {
		st, end := 0, 0
		for end+1 < len(ctxs) {
			ctxs[st] = HE.Evaluator.AddNew(ctxs[end], ctxs[end+1])
			end += 2
			st++
		}
		if end < len(ctxs) {
			ctxs[st] = ctxs[end]
			end++
			st++
		}
		ctxs = ctxs[:st]
	}

	// compute the sum of SlotPerCtx slots in the strip
	HE.Evaluator.InnerSum(ctxs[0], 1, base.SlotPerCtx, ctxs[0])

	// The values in C[x] where x != k*slotPerCtx are not needed in the computation,
	// but may leak information. We randomize them.
	r := InternalSlotRandomizer(base.SlotPerCtx, HE)
	HE.Evaluator.Add(ctxs[0], r, ctxs[0])

	return ctxs[0]
}

func (base *CtxStrip) EucDistance(HE *HEHandler, target *PlainStrip) (dist *rlwe.Ciphertext) {
	base.Sub(HE, target)
	base.Square(HE)
	dist = base.StripeSum(HE)
	return dist
}

func (query *PlainStrip) EuclideanIdentification(HE *HEHandler, dbStrip []*CtxStrip) []*rlwe.Ciphertext {
	out := make([]*rlwe.Ciphertext, len(dbStrip))
	for i := 0; i < len(dbStrip); i++ {
		out[i] = dbStrip[i].EucDistance(HE, query)
	}
	return out
}

func NHammingDistance(
	HE *HEHandler,
	x *PlainStrip, //query data
	xmask *PlainStrip, // query mask
	y_dot_ymask []*CtxStrip, // y.(ymask)
	ybar_dot_ymask []*CtxStrip, // ~y.(ymask)
	ymask []*CtxStrip, // ymask
) (dist []*rlwe.Ciphertext) {
	// Hamming distance = x.~y + ~x.y
	// Normalized Hamming distance = x.xmask.~y.ymask + ~x.xmask.y.ymask

	x_dot_xmask := StripMul(x, xmask)
	xbar_dot_xmask := StripMul(x.LogicNot(), xmask)

	out := make([]*rlwe.Ciphertext, len(ymask))
	for i := 0; i < len(ymask); i++ {
		// This destroys the DB.
		// To support multiple queries, temp variables should be used
		y_dot_ymask[i].Mul(HE, xbar_dot_xmask)
		ybar_dot_ymask[i].Mul(HE, x_dot_xmask)
		ymask[i].Mul(HE, xmask)

		dist := HE.Evaluator.AddNew(y_dot_ymask[i].StripeSum(HE), ybar_dot_ymask[i].StripeSum(HE))
		maskSize := ymask[i].StripeSum(HE)

		// Similarity is computed as follows:
		//   dist/maskSize < MATCH_THRESHOLD/100 =>
		//   100*dist - maskSize*MATCH_THRESHOLD < 0
		// if score > 0, then the sample is not a match (similarity < MATCH_THRESHOLD/100)
		// if score < 0, then the sample is a match
		// note that score is in [0, bfv.p] and negative values are > bfv.p/2
		score := HE.Evaluator.SubNew(HE.Evaluator.MulScalarNew(dist, 100), HE.Evaluator.MulScalarNew(maskSize, MATCH_THRESHOLD))

		out[i] = score
	}
	return out
}

func BPprocessIdReq(encDist []*rlwe.Ciphertext, HE *HEHandler, slotPerCtx int) (answer []uint64) {
	answer = make([]uint64, 0, HE.Params.N()*len(encDist)/slotPerCtx)
	for _, ctx := range encDist {
		raw_answer := HE.Encoder.DecodeUintNew(HE.Decryptor.DecryptNew(ctx))
		for k := 0; k < len(raw_answer); k += slotPerCtx {
			answer = append(answer, raw_answer[k])
		}
	}
	return answer
}
