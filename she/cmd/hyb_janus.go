package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/tuneinsight/lattigo/v4/bfv"
	"local.com/dedup/dedup"
)

var output_addr string = "log.csv"

func bioIdPerformance(bioParam *dedup.JanusParams, bfvParams bfv.Parameters) {
	fmt.Printf("Bio setting: %v\n", bioParam.Describe())

	// Generate the biometric provider's key
	bpHE := &dedup.HEHandler{}
	bpHE.KeyGen(bfvParams)
	rsHE := bpHE.GetPublicHandler()
	janus := dedup.Janus{
		Params: bioParam,
		HE:     rsHE,
	}

	// Initializing a random database
	// In a real application, the database is stored in a file
	start := time.Now()
	janus.GenerateUserDB()
	query := janus.GenerateMatchingQuery(2)
	err := janus.EncryptDatabase()
	if err != nil {
		fmt.Printf("DB encryption error: %v.\n", err)
		return
	}
	initEnd := time.Now()

	// The registration stations computation:
	// Compute the distance between the query and each template in the database
	encDistance := janus.Identification(query)
	data, err := dedup.MarshalCtxArray(encDistance)
	if err != nil {
		fmt.Printf("Marshal encrypted distance error: %v.\n", err)
		return
	}
	rsTimeEnd := time.Now()

	// We do not secret share the encrypted distance in this example.
	// Send the encrypted distance to the biometric provider
	// We do not do this in this example, but we measure the transfer size
	transfer := 0
	for _, v := range data {
		transfer += len(v)
	}

	encDistance, err = dedup.UnMarshalCtxArray(data)
	if err != nil {
		fmt.Printf("UnMarshal encrypted distance error: %v.\n", err)
		return
	}
	answer := dedup.BPprocessIdReq(encDistance, bpHE, janus.Params.SlotsPerCtx)
	answer = answer[:janus.Params.DbSize]
	bpTimeEnd := time.Now()

	// Compute and compare against ground truth
	plainComputation := janus.IdentificationGroundTruth(query)
	if bioParam.BioType == "iris" {
		fmt.Printf("Answer:\n    A negative score (values larger than %v) shows a match.\n", rsHE.Params.T()/2)
		fmt.Printf("    Score:    %v ... %v (%v)\n", answer[:10], answer[len(answer)-10:], len(answer))
	} else if bioParam.BioType == "finger" {
		fmt.Printf("Answer:\n    Computed distance between query and each template.\n")
		fmt.Printf("    Distance: %v ... %v (%v)\n", answer[:10], answer[len(answer)-10:], len(answer))
	}
	fmt.Printf("Ground truth:\n")
	fmt.Printf("    Distance: %v ... %v \n", plainComputation[:10], plainComputation[len(plainComputation)-10:])

	// Print performance measures
	fmt.Printf("*******************************************************\n")
	fmt.Printf("* Performace:\n")
	fmt.Printf("* Initialization (gen and encrypt DB) %v\n", initEnd.Sub(start))
	fmt.Printf("* RS cost (Compute encrypted distance): %v\n", rsTimeEnd.Sub(initEnd))
	fmt.Printf("* BP cost (Decrypt ): %v\n", bpTimeEnd.Sub(rsTimeEnd))
	fmt.Printf("*******************************************************\n")
	fmt.Printf("* Transfer (Bytes): %v\n", transfer)
	fmt.Printf("* Transfer (MB): %v\n", transfer/1024/1024)
	fmt.Printf("*******************************************************\n")

	// Write performance measures to a file
	rs_comp := rsTimeEnd.Sub(initEnd)
	bp_comp := bpTimeEnd.Sub(rsTimeEnd)
	log := fmt.Sprintf("%v,%v,%v,%v,%v\n", janus.Params.DbSize, janus.Params.TemplateSize, rs_comp.Milliseconds(), bp_comp.Milliseconds(), transfer)
	dedup.AppendLine(output_addr, log)
}

func main() {

	db_size := flag.Int("n", 100, "Number of users in the membership database.")
	sensorTS := flag.Int("ts", 64, "The size of the biometric template.")
	sensorD := flag.Int64("d", 256, "The domain of biometric values.")
	bioType := flag.String("biotype", "finger", "The biometric mode from ['finger', 'iris'].")
	ctxPerBatch := flag.Int("ctxPerTemplate", 16, "Strip parameter: number of ciphertexts in strip batching. (Following must hold TS == ctxPerTemplate*slotPerCtx)")
	slotPerCtx := flag.Int("slotPerCtx", 4, "Strip parameter: number of batched elements in strip batching. (Following must hold TS == ctxPerTemplate*slotPerCtx)")
	addr := flag.String("addr", "log.csv", "The address for storing the output file.")
	flag.Parse()
	output_addr = *addr

	// alternative parameters
	// paramDef := bfv.PN12QP109
	// paramDef := bfv.PN13QP218
	// paramDef := bfv.PN14QP438
	// paramDef.T = 0x3ee0001
	// paramDef.T = 4079617
	// paramDef.T = 163841

	// set bfv parameters
	paramDef := bfv.PN12QP101pq // Provides 128-bit post quantom security
	paramDef.T = 4079617
	if *bioType == "finger" && *sensorTS >= 256 {
		// Increasing the template size, increases the max distance and impacts the noise (additive)
		// supporting larger template sizes requires either larger parameter (PN13QP218, N=8192)
		// or if keeping N fixed to 4096, then moving to the non quantion secure version with
		// 109-bit pq.
		paramDef = bfv.PN12QP109 // Provides 128-bit security (slightly lower post quantom security)
		// paramDef = bfv.PN13QP202pq // Slower, but provides 128-bit post quantom security
		paramDef.T = 0x3ee0001
	}
	bfvParams, err := bfv.NewParametersFromLiteral(paramDef)
	if err != nil {
		panic(err)
	}

	hasMask := false
	if *bioType == "iris" {
		hasMask = true
	}
	bioParam := &dedup.JanusParams{
		DbSize:         *db_size,
		TemplateSize:   *sensorTS,
		SensorD:        *sensorD,
		BioType:        *bioType,
		CtxPerTemplate: *ctxPerBatch,
		SlotsPerCtx:    *slotPerCtx,
		SensorHasMask:  hasMask,
		Nbfv:           bfvParams.N(),
	}

	bioIdPerformance(bioParam, bfvParams)
}
