use serde::{Deserialize, Serialize};

use crate::{Message, Notification, Topic};
use crate::models::payload_trait::{Humidity, Light, PayloadTrait, Temperature};

pub mod message;
pub mod notification;
pub mod topic;
pub mod payload_trait;

pub fn get_msq_byte(topic: &Topic, payload_str: &str) -> Vec<u8> {
    let msg_byte: Vec<u8> = match topic.feature.as_str() {
        "temperature" => message_payload_to_bytes::<Temperature>(payload_str, topic),
        "humidity" => message_payload_to_bytes::<Humidity>(payload_str, topic),
        "light" => message_payload_to_bytes::<Light>(payload_str, topic),
        _ => vec![]
    };
    msg_byte
}

fn message_payload_to_bytes<'a, T>(payload_str: &'a str, topic: &Topic) -> Vec<u8>
    where T: Deserialize<'a> + Serialize + Clone + PayloadTrait + Sized {

    // deserialize to a Notification (with turbofish operator "::<Notification>")
    let parsed_result = serde_json::from_str::<Notification<T>>(payload_str);

    match parsed_result {
        Ok(val) => {
            // println!("Notification deserialized from JSON = {:?}", &val);
            let serialized = Message::<T>::new_as_json(
                val.uuid.clone(),
                val.api_token.clone(),
                topic.clone(),
                val.payload,
            );
            serialized.into_bytes()
        }
        Err(err) => {
            eprintln!("Cannot parse JSON from string. Returning empty data. Err = {:?}", &err);
            vec![]
        }
    }
}