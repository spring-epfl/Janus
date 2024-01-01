pub trait Validable {
    fn verify(&self) -> bool;
}

pub struct DummySignature;

impl Validable for DummySignature {
    fn verify(&self) -> bool {
        true
    }
}
