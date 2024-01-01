# Measurement data
In this directory, we provide the measurement data used to plot the graphs in the paper.
This folder includes:

 - ***finger/iris_smc***: provides the cost of running SMC-Janus. We have manually merged the two performance report files generated for RS and BP to have a single CSV file.
 - ***finger/iris_hyb_dist***: provides the cost of computing the biometric similarity in Hyb-Janus using the BFV encryption scheme. 
 - ***finger/iris_hyb_thresh***: provides the cost of checking the biometric similarity against a threshold in Hyb-Janus using a garbled circuit. We have manually merged the two performance report files generated for RS and BP to have a single CSV file.
 - ***finger/iris_tee***: provides the cost of running TEE-Janus using the Fortanix Enclave Development library. *Note:* All computation costs of TEE-Janus are reported in microseconds instead of milliseconds used in other cases.
 - ***Huang***: provides the measurements taken from 
