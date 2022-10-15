pub mod message;
pub mod payload_trait;
pub mod sensor;
pub mod topic;

use crate::models::message::{GenericMessage, Message};
use crate::models::payload_trait::{Humidity, Light, Motion, Temperature};

pub fn new_temperature_message(val: GenericMessage) -> Message<Temperature> {
    let value: Option<f64> = val.payload.get("value").and_then(|value| value.as_f64());
    let payload = Temperature {
        value: value.unwrap() as f32,
    };
    Message::new(val.uuid, val.api_token, val.topic, payload)
}

pub fn new_humidity_message(val: GenericMessage) -> Message<Humidity> {
    let value: Option<f64> = val.payload.get("value").and_then(|value| value.as_f64());
    let payload = Humidity {
        value: value.unwrap() as f32,
    };
    Message::new(val.uuid, val.api_token, val.topic, payload)
}

pub fn new_light_message(val: GenericMessage) -> Message<Light> {
    let value: Option<f64> = val.payload.get("value").and_then(|value| value.as_f64());
    let payload = Light {
        value: value.unwrap() as f32,
    };
    Message::new(val.uuid, val.api_token, val.topic, payload)
}

pub fn new_motion_message(val: GenericMessage) -> Message<Motion> {
    let value: Option<bool> = val.payload.get("value").and_then(|value| value.as_bool());
    let payload = Motion {
        value: value.unwrap(),
    };
    Message::new(val.uuid, val.api_token, val.topic, payload)
}
