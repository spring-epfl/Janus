# Janus: TEE components
TEE-Janus is a TEE software running on SGX written in Rust and using Fortanix EDP.

## Prerequisites

To run TEE-Janus, you will need a machine supporting SGX. Please, refer to [Intel's documentation](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/get-started.html) to check if your hardware is compatible, and if you will need to install some extra software to make it work.

You will also need the latest nightly Rust toolchain. Please, refer to [Rust's documentation](https://www.rust-lang.org/tools/install) how to install the nightly toolchain of Rust.

And finally, you will need to run the AESM service on your machine. Please refer to [Fortanix' documentation](https://edp.fortanix.com/docs/installation/guide/) to install it.

## Running the Program

Once the software stack is installed, you can build and run the program with `cargo`:

```
$ cargo run --release
```
