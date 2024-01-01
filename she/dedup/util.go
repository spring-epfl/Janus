package dedup

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type HEHandler struct {
	Params    bfv.Parameters
	Encoder   bfv.Encoder
	Encryptor rlwe.Encryptor
	Decryptor rlwe.Decryptor
	Evaluator bfv.Evaluator
}

func (he *HEHandler) KeyGen(params bfv.Parameters) {
	he.Params = params
	kgen := bfv.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPair()
	evk := rlwe.EvaluationKey{
		Rlk:  kgen.GenRelinearizationKey(sk, 2),
		Rtks: kgen.GenRotationKeysForInnerSum(sk),
	}

	he.Encoder = bfv.NewEncoder(params)
	he.Decryptor = bfv.NewDecryptor(params, sk)
	he.Encryptor = bfv.NewEncryptor(params, pk)
	he.Evaluator = bfv.NewEvaluator(params, evk)
}

func (priv *HEHandler) GetPublicHandler() *HEHandler {
	return &HEHandler{
		Params:    priv.Params,
		Encoder:   priv.Encoder,
		Encryptor: priv.Encryptor,
		Decryptor: nil,
		Evaluator: priv.Evaluator,
	}
}

func ApplyMask(data, mask []int64) []int64 {
	out := make([]int64, len(data))
	for i := range data {
		if mask[i] != 0 {
			out[i] = data[i]
		} else {
			out[i] = 0
		}
	}
	return out
}
func MergeMask(m1, m2 []int64) []int64 {
	out := make([]int64, len(m1))
	for i := range m1 {
		out[i] = m1[i] * m2[i]
	}
	return out
}

func DistEuclidean(a, b []int64) int64 {
	dist := int64(0)
	for i := range a {
		dist += (a[i] - b[i]) * (a[i] - b[i])
	}
	return dist
}

func IsPowerOf2(a int) bool {
	return (a & (a - 1)) == 0
}

func NextPow2(a int) int {
	out := 1
	for out < a {
		out *= 2
	}
	return out
}

func ExtendedRotate(
	params *bfv.Parameters,
	evaluator bfv.Evaluator,
	rot int,
	ctx *rlwe.Ciphertext,
) {
	if rot < 0 {
		rot += int(params.N() / 2)
	}

	for k := 1; rot > 0; k *= 2 {
		if rot%2 == 1 {
			evaluator.RotateColumns(ctx, k, ctx)
		}
		rot /= 2
	}
}

// Create a random ptx to randomize (additive) all SIMD slots that are not
// in the form of k*dataStep.
// This prevent leakage from internal slots in inner sum.
func InternalSlotRandomizer(dataStep int, HE *HEHandler) *rlwe.Plaintext {
	data := make([]uint64, HE.Params.N())
	for i := 0; i < int(len(data)); i++ {
		// Warning: USE CRYPTO SECURE RANDOMNESS
		if (i % dataStep) == 0 {
			data[i] = 0
		} else {
			data[i] = rand.Uint64() % HE.Params.T()
		}
	}

	return HE.Encoder.EncodeNew(data, HE.Params.MaxLevel())
}

func AppendLine(filepath string, s string) {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("***** Error opening performance log")
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		fmt.Print("***** Error writing in the performance log")
	}
}

func MarshalCtxArray(ctxs []*rlwe.Ciphertext) ([][]byte, error) {
	marshalledCiphers := make([][]byte, 0, len(ctxs))
	for _, ctx := range ctxs {
		data, err := ctx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		marshalledCiphers = append(marshalledCiphers, data)
	}
	return marshalledCiphers, nil
}

func UnMarshalCtxArray(data [][]byte) ([]*rlwe.Ciphertext, error) {
	ctxs := make([]*rlwe.Ciphertext, len(data))
	for i := range ctxs {
		ctxs[i] = new(rlwe.Ciphertext)
		err := ctxs[i].UnmarshalBinary(data[i])
		if err != nil {
			return nil, err
		}
	}
	return ctxs, nil
}

func DescribeParams(params bfv.Parameters) {
	fmt.Println("================== Parameters ==================")
	fmt.Printf("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
		1<<params.LogN(), params.T(), params.LogQP(), params.Sigma())
	fmt.Println("================================================")
}

func TestDepth(params bfv.Parameters) {
	DescribeParams(params)
	kgen := bfv.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPair()
	evk := rlwe.EvaluationKey{
		Rlk:  kgen.GenRelinearizationKey(sk, 2),
		Rtks: kgen.GenRotationKeysForInnerSum(sk),
	}

	encoder := bfv.NewEncoder(params)
	decryptor := bfv.NewDecryptor(params, sk)
	encryptorPk := bfv.NewEncryptor(params, pk)
	encryptorSk := bfv.NewEncryptor(params, sk)
	evaluator := bfv.NewEvaluator(params, evk)

	_ = evaluator
	_, _ = encryptorPk, encryptorSk

	data := []uint64{1, 2, 3, 4}

	ptx := encoder.EncodeNew(data, params.MaxLevel())
	ctx := encryptorSk.EncryptNew(ptx)

	for i := 1; i < 20; i++ {
		ctx = evaluator.MulNew(ctx, ptx)
		// ctx = evaluator.MulNew(ctx, ctx)
		// evaluator.Relinearize(ctx, ctx)
		// ctx = evaluator.MulScalarNew(ctx, 2)
		dec := encoder.DecodeUintNew(decryptor.DecryptNew(ctx))
		fmt.Printf("[Depth %v] -> %v\n", i, dec[:10])
	}
}
