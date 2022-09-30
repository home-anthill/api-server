use serde::{Deserialize, Serialize};

use crate::{Topic};
use crate::models::payload_trait::PayloadTrait;

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct Message<T> where T: PayloadTrait + Sized + Serialize {
    pub uuid: String,
    pub api_token: String,
    pub topic: Topic,
    pub payload: T,
}

impl<T> Message<T> where T: PayloadTrait + Sized + Serialize {
    pub fn new(uuid: String, api_token: String, topic: Topic, payload: T) -> Message<T> {
        Self {
            uuid,
            api_token,
            topic,
            payload,
        }
    }

    pub fn new_as_json(uuid: String, api_token: String, topic: Topic, payload: T) -> String {
        // println!("Notification deserialized from JSON = {:?}", &val);
        let message = Self::new(
            uuid,
            api_token,
            topic,
            payload,
        );
        // println!("message {:?}", &queue_message);
        serde_json::to_string(&message).unwrap()
    }
}