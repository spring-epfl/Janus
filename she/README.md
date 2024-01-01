# Janus: SHE components
This directory includes the SHE component used to compute the biometric distance in Hyb-Janus.


## Building 
This component is developed with the Go language. Please follow the instructions in [install go](https://go.dev/doc/install) to install the Go compiler. You can use the following commands to compile our code:

```bash
$ go get ./...
$ go build cmd/hyb_janus.go
```

## Running
As we aim to measure the performance cost, we create a single application to measure the cost for both the registration station and biometric provider instead of fully deploying a client/server architecture.

You can use the command line interface as follows:
```bash
Usage of ./hyb_janus:
  -addr string
      The address for storing the output file. (default "log.csv")
  -biotype string
      The biometric mode from ['finger', 'iris']. (default "finger")
  -ctxPerTemplate int
      Strip parameter: number of ciphertexts in strip batching. (Following must hold TS == ctxPerTemplate*slotPerCtx) (default 16)
  -d int
      The domain of biometric values. (default 256)
  -n int
      Number of users in the membership database. (default 100)
  -slotPerCtx int
      Strip parameter: number of batched elements in strip batching. (Following must hold TS == ctxPerTemplate*slotPerCtx) (default 4)
  -ts int
      The size of the biometric template. (default 64)
```

We provide a script `bench.sh` to store the configuration of our experiments in the paper to facilitate their recreation. This script generates two files `hybdist_finger.csv` and `hybdist_iris.csv` that record the performance of running identification with the following sensor configurations: `[FingerSensor(64, 256), FingerSensor(64, 256), IrisSensor(2048, 2), IrisSensor(10240, 2)]`.


If you want to set parameters manually, you should check the [Strip packing section](#strip-packing) for information on how to set `ctxPerTemplate` and `slotPerCtx`.


## Strip packing
We use the SIMD packing capability of the BFV scheme to pack N scalar values into each cipher.
While it is possible to pack each template into a single ciphertext, the cost of performing the inner sum of TS elements packed into a single ciphertext is higher than adding TS ciphertexts. This is due to requiring bfv.rotation to perform operations such as addition between different slots of a single ciphertext. 
We created a new packing scheme called strip packing to provide a transfer/performance trade-off. Strip packing represents each template of size TS as a 2D array of dimensions CtxPerTemplate x SlotPerCtx.
To effectively use this packing, we batch (N/SlotPerCtx) templates into CtxPerTemplate ciphertexts.
Adding ciphertext together is one of the cheapest operations in BFV. Having a small SlotPerCtx
leads to a lower number of BFV.rotations (needed to compute the inner sum) at
the cost of having more ciphertext to send.

We provide recommended configuration for experiments in the `bench.sh` script. If you want to manually set these parameters, you need to ensure that `templateSize = CtxPerTemplate * SlotPerCtx` and `SlotPerCtx = 2^k` is a power of two.


## Acknowledgement
This work is based on the Lattigo FHE library.
