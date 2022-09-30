pub mod message;
pub mod topic;
pub mod payload_trait;

pub mod sensor;

use crate::models::message::{GenericMessage, Message};
use crate::models::payload_trait::{Humidity, Light, Temperature};

// this fn is very bad -> find a generic way to process all T in a single fn
pub fn new_temperature_message(val: GenericMessage) -> Message<Temperature> {
    let temperature: Option<f64> = val.payload
        .get("value")
        .and_then(|value| value.as_f64());

    let payload = Temperature {
        value: temperature.unwrap()
    };
    let message: Message<Temperature> = Message::new(
        val.uuid,
        val.api_token,
        val.topic,
        payload,
    );
    message
}

// this fn is very bad -> find a generic way to process all T in a single fn
pub fn new_humidity_message(val: GenericMessage) -> Message<Humidity> {
    let humidity: Option<f64> = val.payload
        .get("value")
        .and_then(|value| value.as_f64());

    let payload = Humidity {
        value: humidity.unwrap()
    };
    let message: Message<Humidity> = Message::new(
        val.uuid,
        val.api_token,
        val.topic,
        payload,
    );
    message
}


// this fn is very bad -> find a generic way to process all T in a single fn
pub fn new_light_message(val: GenericMessage) -> Message<Light> {
    let light: Option<f64> = val.payload
        .get("value")
        .and_then(|value| value.as_f64());

    let payload = Light {
        value: light.unwrap()
    };
    let message: Message<Light> = Message::new(
        val.uuid,
        val.api_token,
        val.topic,
        payload,
    );
    message
}