use log::{info, debug, error};

use mongodb::bson::{Bson, Document};
use mongodb::Database;
use rocket::serde::json::Json;

use crate::models::inputs::RegisterInput;
use crate::models::sensor::{new_from_register_input, BooleanSensor, FloatSensor};

pub async fn insert_register(
    db: &Database,
    input: Json<RegisterInput>,
    sensor_type: &str,
) -> mongodb::error::Result<String> {
    info!(target: "app", "insert_register - Called with sensor_type = {}", sensor_type);

    let collection = db.collection::<Document>(sensor_type);

    let serialized_data: Bson;
    if sensor_type == "temperature" || sensor_type == "humidity" || sensor_type == "light" {
        serialized_data = new_from_register_input::<FloatSensor>(input).unwrap();
    } else if sensor_type == "motion" {
        serialized_data = new_from_register_input::<BooleanSensor>(input).unwrap();
    } else {
        error!(target: "app", "insert_register - Unknown sensor_type = {}", sensor_type);
        // TODO return a custom error instead of use `panic`
        panic!("Unknown type")
    }

    debug!(target: "app", "insert_register - Adding sensor into db");

    let document = serialized_data.as_document().unwrap();
    let insert_one_result = collection
        .insert_one(document.to_owned(), None)
        .await
        .unwrap();
    Ok(insert_one_result
        .inserted_id
        .as_object_id()
        .unwrap()
        .to_hex())
}
