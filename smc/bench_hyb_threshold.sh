#!/bin/sh
echo "Usage: ./bench_hyb_threshold.sh <bio_type> <fuse>"
base_port=$(( 10000 + 1*1000 ))

bio_type=$1  # biometric type in [iris,face]
fuse=$2      # number of fused templates
rep=1        # number of repetitions
file_name="hybth_${bio_type}_f${fuse}" # base output address

echo "base port: $base_port"
echo "Bio type: $bio_type"
echo "File name: $file_name"
echo "Rep num: $rep"
echo "Fues: $fuse"



echo "Build..."
(cd build && pwd && make -j ) 

app=./build/bin/hyb_threshold

echo "users,fuse,bp_comp,bp_transfer" >> "${file_name}_bp.csv"
echo "users,fuse,rs_comp,rs_transfer" >> "${file_name}_rs.csv"

for N in 32 64 128 256 512 1024 2048 4096 8192 16384
do
    echo "Running with $N registered users."
    for r in `seq 1 $rep`
        do
        echo "Run $r."
        port=$(( $base_port + $N/4 + $r ))
        $app BP $port --bio-type $bio_type --addr "${file_name}_bp.csv" -N $N -f $fuse 2> err_bp &
        sleep 0.1
        $app RS $port --bio-type $bio_type --addr "${file_name}_rs.csv" -N $N -f $fuse 2> err_rs
        sleep 0.1
    done
done



