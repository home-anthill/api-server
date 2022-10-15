use mongodb::bson::oid::ObjectId;
use mongodb::bson::DateTime;
use serde::{Deserialize, Serialize};

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct SensorDocument {
    pub _id: ObjectId,
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
    pub createdAt: DateTime,
    pub modifiedAt: DateTime,
    pub value: f32,
}

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Sensor {
    pub _id: String,
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
    pub createdAt: String,
    pub modifiedAt: String,
    pub value: f32,
}
