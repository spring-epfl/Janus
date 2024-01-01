#!/bin/sh
echo "Building the benchmark for the distance computation component of Hyb-Janus."

go build ./cmd/hyb_janus.go
app=./hyb_janus
rep=2

fname="hybdist_finger.csv"
echo "Storing the finger biometric benchmark in $fname"
echo "users,template_size,rs_comp,bp_comp,he_transfer" >> $fname
for N in 64 128 256 512 1024 2048 4096 8192 # 16384 32768
do
    echo "Running finger membership with $N registered users."
    for r in `seq 1 $rep`
    do
        echo "Run $r."
        $app -biotype "finger" -n $N -addr $fname -ts 64  -d 256 -ctxPerTemplate 64  -slotPerCtx 1
        $app -biotype "finger" -n $N -addr $fname -ts 640 -d 256 -ctxPerTemplate 320 -slotPerCtx 2
    done
done

fname="hybdist_iris.csv"
echo "Storing the finger biometric benchmark in $fname"
echo "users,template_size,rs_comp,bp_comp,he_transfer" >> $fname
for N in 64 128 256 512 1024 2048 4096 8192 # 16384
do
    echo "Running  iris membership with $N registered users."
    for r in `seq 1 $rep`
    do
        echo "Run $r."
        $app -biotype "iris" -n $N -addr $fname -ts 2048  -d 2 -ctxPerTemplate 512  -slotPerCtx 4
        $app -biotype "iris" -n $N -addr $fname -ts 10240 -d 2 -ctxPerTemplate 1280 -slotPerCtx 8
    done
done


