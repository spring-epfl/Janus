# Janus: Safe Biometric Deduplication for Humanitarian Aid Distribution
This repository accompanies the paper "Janus: Safe Biometric Deduplication for Humanitarian Aid Distribution" by Kasra EdalatNejad (SPRING Lab, EPFL), Wouter Lueks (CISPA Helmholtz Center for Information Security), Justinas Sukaitis (International Committee of the Red Cross), Vincent Graf Narbel (International Committee of the Red Cross), Massimo Marelli (International Committee of the Red Cross), and Carmela Troncoso (SPRING Lab, EPFL).

***Warning. This code is an academic prototype and is not suitable to be used in production as is.*** 

> **Abstract:**
> Humanitarian organizations provide aid to people in need. To use their limited budget efficiently, their distribution processes must ensure that legitimate recipients cannot receive more aid than they are entitled to. Thus, it is essential that recipients can register at most once per aid program.
> Taking the International Committee of the Red Cross's aid distribution registration process as a use case, we identify the requirements to detect double registration without creating new risks for aid recipients. We then design Janus, which combines privacy-enhancing technologies with biometrics to prevent double registration in a safe manner. Janus does not create plaintext biometric databases and reveals only *one bit* of information at registration time (whether the user registering is present in the database or not). We implement and evaluate three instantiations of Janus based on secure multiparty computation (SMC) alone, a hybrid of somewhat homomorphic encryption and SMC, and trusted execution environments. We demonstrate that they support the privacy, accuracy, and performance needs of humanitarian organizations. We compare Janus with existing alternatives and show it is the first system that provides the accuracy our scenario requires while providing strong protection.


## This repository
This repository serves three goals:

1. It contains the **implementation of Janus**. We organize our code based on the techniques used:
 * **TEE:** Provides the rust code for TEE-Janus based on SGX and using Fortanix Enclave Development Kit.
 * **SMC:** Provides the C++ code for SMC-Janus and the thresholding portion of Hyb-Janus using a garbled circuit compiler called EMP-toolkit.
 * **HE:** Provides the Go code for computing the biometric similarity using a somewhat homomorphic encryption scheme (BFV)in Hyb-Janus using the Lattigo library.

2. To enable **reproducing the graphs** in the paper. The scripts in `plot/` provide utilities for processing raw measurements and plotting the graphs in the paper.

3. To **store measurements** for the performance of Janus and of related work. Measurement data can be found in `data`.