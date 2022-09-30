use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Temperature {
    pub value: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Humidity {
    pub value: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Light {
    pub value: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Motion {
    pub value: bool,
}

pub trait PayloadTrait {
    fn get_value(&self) -> f64;
}

impl PayloadTrait for Temperature {
    fn get_value(&self) -> f64 {
        self.value
    }
}

impl PayloadTrait for Humidity {
    fn get_value(&self) -> f64 {
        self.value
    }
}

impl PayloadTrait for Light {
    fn get_value(&self) -> f64 {
        self.value
    }
}

impl PayloadTrait for Motion {
    fn get_value(&self) -> f64 {
        if self.value {
            1.0
        } else {
            0.0
        }
    }
}