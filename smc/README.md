# Janus: SMC components
This directory includes the SMC components used in Janus including the full instantiation of SMC-Janus and the thresholding portion of Hyb-Janus.

## Building 
To build the code you need to have `git`, `gcc`, `g++`, `cmake`, and `python` installed on your system. We have only tested our code on Debian.

We rely on the EMP-toolkit to instantiate our gabled circuits. 

***Note***: We use a modified version of EMP-tool ([repo](https://github.com/spring-epfl/emp-tool/)) that allows exporting the communication cost.

### Building with script
You can either use our installer script to build the project or manually build dependencies following the detailed instructions.

```bash
$ bash installer.sh
```

This creates two executable files `build/bin/smc_janus` and  `build/bin/hyb_threshold`.

### Manual build
First, you need to build and install the EMP-toolkit as follows:

```bash
$ cd libs
$ python3 install_emp.py --deps --tool --ot --sh2pc
```

After installing EMP, you can build SMC-Janus and the thresholding portion of Hyb-Janus as follows:

```bash
$ cmake -S . -B build
$ cd build
$ make -j
```

This creates two executable files `build/bin/smc_janus` and  `build/bin/hyb_threshold`.

## Running

### Benchmark script
We provide benchmark scripts to run experiments in the paper and evaluate the performance.

*SMC-Janus.* You can benchmark the instantiation as follows:
```bash
$ ./bench_smc_janus.sh <bio_type> <fuse> <template_size>
```
The script takes the biometric mode `bio_type` chosen from `['finger', 'iris']`,  `fuse` the number of fused templates per user (f), `template_size` the size of each biometric template (TS). The script saves the performance measure of the registration station and the biometric provider to the CSV files `smc_${bio_type}_f${fuse}_rs.csv` and `smc_${bio_type}_f${fuse}_bp.csv`. Computation cost is reported in milliseconds and transfer cost is reported in bytes. We have used the following parameters for the experiments in the paper:

```bash
$ bench_smc_janus 'finger' 1 64 # 1 finger code of size 64
$ bench_smc_janus 'finger' 4 64 # 4 finger codes of size 64
$ bench_smc_janus 'iris' 1 2048 # 1 iris code of size 2048
$ bench_smc_janus 'iris' 2 2048 # 2 iris codes of size 2048
```



*Hyb-Janus thresholding.* You can benchmark the thresholding portion of Hyb-Janus as follows:
```bash
$ ./bench_hyb_threshold.sh <bio_type> <fuse>
```
The script takes the biometric mode `bio_type` chosen from `['finger', 'iris']` and `fuse` the number of fused templates per user (f). The threshold component uses random secret-shared similarity scores based on the similarity component of Hyb-Janus in the `she` directory. This component does not take raw templates and is not affected by the template size, so unlike SMC-Janus there is no argument for `template_size`.
The script saves the performance measure of the registration station and the biometric provider to the CSV files `smc_${bio_type}_f${fuse}_rs.csv` and `smc_${bio_type}_f${fuse}_bp.csv`. Computation cost is reported in milliseconds and transfer cost is reported in bytes. We have used the following parameters for the experiments in the paper:

```bash
$ bench_hyb_threshold 'finger' 1 # 1 finger code 
$ bench_hyb_threshold 'finger' 4 # 4 finger codes
$ bench_hyb_threshold 'iris' 1 # 1 iris code 
$ bench_hyb_threshold 'iris' 2 # 2 iris codes
```

### Direct 
Janus is a two-party protocol. The registration station acts as a server and starts listening on the provided `port` and the biometric provider connects to it. Running the executable with the help flag `-h, --help` prints the following manual:


*SMC-Janus*: You can run parties as follows:
```bash
$ build/bin/smc_janus --help
SMC-Janus
Usage: ./smc_janus [OPTIONS] party port

Positionals:
  party TEXT REQUIRED         Party: [RS, BP]
  port INT REQUIRED           Network port

Options:
  -h,--help                   Print this help message and exit
  --bio-type TEXT             Biometric type: [finger, iris]
  --addr TEXT                 Address of the benchmark file
  -N INT                      Number of users
  -f INT                      Number of templates per user
  --ts INT                    The size of the biometric template
```

*Hyb-Janus threshold component*: You can run parties as follows:
```bash
$ build/bin/smc_janus --help
Hyb-Janus threshold component
Usage: ./hyb_threshold [OPTIONS] party port

Positionals:
  party TEXT REQUIRED         Party: [RS, BP]
  port INT REQUIRED           Network port

Options:
  -h,--help                   Print this help message and exit
  --bio-type TEXT             Biometric type: [finger, iris]
  --addr TEXT                 Address of the output benchmark file
  -N INT                      Number of registered users (N)
  -f INT                      Number of biometric templates per user (f)
```



## Acknowledgement
This work uses the EMP-toolkit SMC compiler and CLI11 library.

