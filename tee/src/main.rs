pub mod code;
pub mod signature;

use std::time::Instant;


use rand::SeedableRng;
use rand_xoshiro::Xoshiro256Plus;

use code::{BiometricCodeSystemDatabase, FingerCode, IrisCode, BiometricCodeSystem, RandomlyGenerated};
use signature::DummySignature;


fn finger<const N: usize>(seed: u64) {
    let mut rng = Xoshiro256Plus::seed_from_u64(seed);

    let db_finger: BiometricCodeSystemDatabase<N, FingerCode> = BiometricCodeSystemDatabase::random(&mut rng);
    let finger = FingerCode::random(&mut rng);
    let signature = DummySignature;

    let time = Instant::now();
    db_finger.verify(&finger, &signature);
    println!("Fingerprint N={}: {} µs", N, time.elapsed().as_micros());
}

fn iris<const N: usize>(seed: u64) {
    let mut rng = Xoshiro256Plus::seed_from_u64(seed);

    let db_iris: BiometricCodeSystemDatabase<N, IrisCode> = BiometricCodeSystemDatabase::random(&mut rng);
    let iris = IrisCode::random(&mut rng);
    let signature = DummySignature;

    let time = Instant::now();
    db_iris.verify(&iris, &signature);
    println!("Iris Pattern N={}: {} µs", N, time.elapsed().as_micros());
}

fn main() {
    for i in 1..5 {
        finger::<0x1>(100 * i + 1);
        finger::<0x2>(100 * i + 2);
        finger::<0x4>(100 * i + 3);
        finger::<0x8>(100 * i + 4);
        finger::<0x10>(100 * i + 5);
        finger::<0x20>(100 * i + 6);
        finger::<0x40>(100 * i + 7);
        finger::<0x80>(100 * i + 8);
        finger::<0x100>(100 * i + 9);
        finger::<0x200>(100 * i + 10);
        finger::<0x400>(100 * i + 11);
        finger::<0x800>(100 * i + 12);
        finger::<0x1000>(100 * i + 13);
        finger::<0x2000>(100 * i + 14);
        finger::<0x4000>(100 * i + 15);
        finger::<0x8000>(100 * i + 16);
    }
    for i in 1..5 {
        iris::<0x1>(100 * i + 1);
        iris::<0x2>(100 * i + 2);
        iris::<0x4>(100 * i + 3);
        iris::<0x8>(100 * i + 4);
        iris::<0x10>(100 * i + 5);
        iris::<0x20>(100 * i + 6);
        iris::<0x40>(100 * i + 7);
        iris::<0x80>(100 * i + 8);
        iris::<0x100>(100 * i + 9);
        iris::<0x200>(100 * i + 10);
        iris::<0x400>(100 * i + 11);
        iris::<0x800>(100 * i + 12);
        iris::<0x1000>(100 * i + 13);
        iris::<0x2000>(100 * i + 14);
        iris::<0x4000>(100 * i + 15);
        iris::<0x8000>(100 * i + 16);
    }
}
