use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Temperature {
    pub value: f32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Humidity {
    pub value: f32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Light {
    pub value: f32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Motion {
    pub value: bool,
}

pub trait PayloadTrait {
    fn get_value(&self) -> f32;
}

impl PayloadTrait for Temperature {
    fn get_value(&self) -> f32 {
        self.value
    }
}

impl PayloadTrait for Humidity {
    fn get_value(&self) -> f32 {
        self.value
    }
}

impl PayloadTrait for Light {
    fn get_value(&self) -> f32 {
        self.value
    }
}

impl PayloadTrait for Motion {
    fn get_value(&self) -> f32 {
        if self.value {
            1.0
        } else {
            0.0
        }
    }
}
