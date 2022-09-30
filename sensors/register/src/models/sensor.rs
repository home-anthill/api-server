use mongodb::bson::oid::ObjectId;
use mongodb::bson::DateTime;
use serde::{Deserialize, Serialize};

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct TemperatureSensor {
    #[serde(rename = "_id")]
    pub id: ObjectId,
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
    pub createdAt: DateTime,
    pub modifiedAt: DateTime,
    pub value: f64,
}

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct HumiditySensor {
    #[serde(rename = "_id")]
    pub id: ObjectId,
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
    pub createdAt: DateTime,
    pub modifiedAt: DateTime,
    pub value: f64,
}

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct LightSensor {
    #[serde(rename = "_id")]
    pub id: ObjectId,
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
    pub createdAt: DateTime,
    pub modifiedAt: DateTime,
    pub value: f64,
}