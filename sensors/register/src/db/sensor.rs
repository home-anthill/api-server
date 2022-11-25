use log::{debug, error, info};

use mongodb::bson::{doc, Bson, Document};
use mongodb::options::FindOneOptions;
use mongodb::Database;
use rocket::serde::json::Json;

use crate::models::inputs::RegisterInput;
use crate::models::sensor::{new_from_register_input, FloatSensor, IntSensor};

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
    } else if sensor_type == "motion" || sensor_type == "airquality" || sensor_type == "airpressure" {
        serialized_data = new_from_register_input::<IntSensor>(input).unwrap();
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

pub async fn get_sensor(
    db: &Database,
    uuid: &String,
    sensor_type: &String,
) -> mongodb::error::Result<Option<Document>> {
    info!(target: "app", "get_sensor - Called with uuid = {}, sensor_type = {}", uuid, sensor_type);

    let collection = db.collection::<Document>(sensor_type.as_str());

    let filter = doc! { "uuid": uuid };
    let projection = doc! {"_id": 0, "value": 1};
    let find_options = FindOneOptions::builder().projection(projection).build();

    debug!(target: "app", "get_sensor - Getting sensor value with uuid = {} from db", uuid);

    collection.find_one(filter, find_options).await
}
