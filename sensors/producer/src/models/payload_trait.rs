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

pub trait PayloadTrait {}

impl PayloadTrait for Temperature {}

impl PayloadTrait for Humidity {}

impl PayloadTrait for Light {}

impl PayloadTrait for Motion {}
