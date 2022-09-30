use std::fmt::Error;
use serde::Serialize;

use crate::models::sensor::{HumiditySensor, LightSensor, TemperatureSensor};
use crate::models::inputs::RegisterInput;

use mongodb::bson::ser::Result;
use mongodb::bson::{to_bson, from_bson};
use mongodb::bson::oid::ObjectId;
use mongodb::bson::{Bson, DateTime, Document};
use mongodb::Database;
use rocket::serde::json::Json;

pub async fn insert_register(
    db: &Database,
    input: Json<RegisterInput>,
    sensor_type: &str,
) -> mongodb::error::Result<String> {
    let collection = db.collection::<Document>(sensor_type);

    let serialized_data: Bson;
    if sensor_type == "temperature" {
        serialized_data = temperature_bson(&input).unwrap();
    } else if sensor_type == "humidity" {
        serialized_data = humidity_bson(&input).unwrap();
    } else if sensor_type == "light" {
        serialized_data = light_bson(&input).unwrap();
    } else {
        // TODO return a custom error
        // return Err(Error)
        panic!("Unknown type")
    }

    let document = serialized_data.as_document().unwrap();

    let insert_one_result = collection
        .insert_one(document.to_owned(), None)
        .await
        .unwrap();

    Ok(insert_one_result.inserted_id.as_object_id().unwrap().to_hex())
}

fn temperature_bson(input: &Json<RegisterInput>) -> Result<Bson> {
    let date_now: DateTime = DateTime::now();
    let sensor = TemperatureSensor {
        id: ObjectId::new(),
        uuid: input.uuid.clone(),
        mac: input.mac.clone(),
        manufacturer: input.manufacturer.clone(),
        model: input.model.clone(),
        profileOwnerId: input.profileOwnerId.clone(),
        apiToken: input.apiToken.clone(),
        createdAt: date_now,
        modifiedAt: date_now,
        value: 0.0,
    };
    to_bson(&sensor)
}

fn humidity_bson(input: &Json<RegisterInput>) -> Result<Bson> {
    let date_now: DateTime = DateTime::now();
    let sensor = HumiditySensor {
        id: ObjectId::new(),
        uuid: input.uuid.clone(),
        mac: input.mac.clone(),
        manufacturer: input.manufacturer.clone(),
        model: input.model.clone(),
        profileOwnerId: input.profileOwnerId.clone(),
        apiToken: input.apiToken.clone(),
        createdAt: date_now,
        modifiedAt: date_now,
        value: 0.0,
    };
    to_bson(&sensor)
}

fn light_bson(input: &Json<RegisterInput>) -> Result<Bson> {
    let date_now: DateTime = DateTime::now();
    let sensor = LightSensor {
        id: ObjectId::new(),
        uuid: input.uuid.clone(),
        mac: input.mac.clone(),
        manufacturer: input.manufacturer.clone(),
        model: input.model.clone(),
        profileOwnerId: input.profileOwnerId.clone(),
        apiToken: input.apiToken.clone(),
        createdAt: date_now,
        modifiedAt: date_now,
        value: 0.0,
    };
    to_bson(&sensor)
}