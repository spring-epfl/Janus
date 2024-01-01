
use std::hint::black_box;

use rand::prelude::Rng;

use crate::signature::Validable;

const BITS_IN_IRIS_CODE: usize = 10240;
const U64_IN_IRIS_CODE: usize = BITS_IN_IRIS_CODE / 64;
const BYTES_IN_FINGER_CODE: usize = 640;
const FINGER_CODE_MATCH_THRESHOLD: f32 = 1_000_000.0;
const IRIS_CODE_MATCH_THRESHOLD: f64 = 0.4;


pub trait BiometricCode {
    fn do_match(&self, other: &Self) -> bool;
}

pub trait BiometricCodeSystem<T: BiometricCode> {
    fn verify(&self, code: &T, signature: &impl Validable) -> bool;
}

pub trait RandomlyGenerated<R: Rng> {
    fn random(rng: &mut R) -> Self;
}

#[derive(Debug, Clone, Copy)]
pub struct IrisCode {
    pub biometric_data: [u64; U64_IN_IRIS_CODE],
    pub mask: [u64; U64_IN_IRIS_CODE],
}

impl Default for IrisCode {
    fn default() -> Self {
        IrisCode { biometric_data: [0; U64_IN_IRIS_CODE], mask: [0; U64_IN_IRIS_CODE] }
    }

}

impl BiometricCode for IrisCode {
    fn do_match(&self, other: &Self) -> bool {
        let distance: u32 = self.biometric_data.iter().zip(self.mask.iter()).zip(other.biometric_data.iter().zip(other.mask.iter())).map(
            |((q1, m1), (q2, m2))|
            (((q1 & m1) & m2) ^ ((q2 & m1) & m2)).count_ones()
        ).sum();
        let sz: u32 = self.mask.iter().zip(other.mask.iter()).map(|(m1 ,m2)|(m1 | m2).count_ones()).sum();

        10_000.0 * f64::from(distance) / (f64::from(sz) + 1.0) < IRIS_CODE_MATCH_THRESHOLD
    }
}

impl<R: Rng> RandomlyGenerated<R> for IrisCode {
    fn random(rng: &mut R) -> Self {
        let mut code = IrisCode::default();
        rng.fill::<[u64]>(&mut code.biometric_data);
        rng.fill::<[u64]>(&mut code.mask);
        code
    }
}


#[derive(Debug, Clone, Copy)]
pub struct FingerCode {
    pub biometric_data: [u8; BYTES_IN_FINGER_CODE]
}

impl Default for FingerCode {
    fn default() -> Self {
        FingerCode { biometric_data: [0; BYTES_IN_FINGER_CODE] }
    }

}

impl BiometricCode for FingerCode {
    fn do_match(&self, other: &Self) -> bool {
        // Compute Euclidian distance for each elements of the fingerprint.
        let distance: f32 = self.biometric_data.iter().zip(other.biometric_data.iter()).map(|(&a, &b)| f32::from(a) - f32::from(b)).map(|x| x * x).sum();
        distance < FINGER_CODE_MATCH_THRESHOLD
    }
}


impl<R: Rng> RandomlyGenerated<R> for FingerCode {
    fn random(rng: &mut R) -> Self {
        let mut code = FingerCode::default();
        rng.fill::<[u8]>(&mut code.biometric_data);
        code
    }
}


#[derive(Debug, Clone, Copy)]
pub struct BiometricCodeSystemDatabase<const N: usize, T: BiometricCode> {
    database: [T; N]
}

impl<const N: usize, T: BiometricCode + Default + Copy> Default for BiometricCodeSystemDatabase<N, T> {
    fn default() -> Self {
        BiometricCodeSystemDatabase { database: [T::default(); N] }
    }

}
impl<const N: usize, T: BiometricCode> BiometricCodeSystem<T> for BiometricCodeSystemDatabase<N, T> {
    fn verify(&self, code: &T, signature: &impl Validable) -> bool {
        let mut is_match = false;
        for db_code in self.database.iter() {
            is_match |= black_box(code.do_match(db_code));
        }
        is_match & signature.verify()
    }
}


impl<const N: usize, T: BiometricCode + RandomlyGenerated<R> + Copy, R: Rng> RandomlyGenerated<R> for BiometricCodeSystemDatabase<N, T> {
    fn random(rng: &mut R) -> Self {
        BiometricCodeSystemDatabase { database: [T::random(rng); N] }
    }
}
