use crate::models::sensor::Sensor;
use crate::models::sensor::SensorDocument;

use mongodb::bson::{doc, DateTime};
use mongodb::options::FindOneAndUpdateOptions;
use mongodb::options::ReturnDocument;
use mongodb::{Database};
use serde::Serialize;
use crate::models::message::Message;
use crate::models::payload_trait::{PayloadTrait};

pub async fn update_message<T: PayloadTrait + Sized + Serialize>(
    db: &Database,
    message: &Message<T>,
) -> mongodb::error::Result<Option<Sensor>> {
    let collection = db.collection::<SensorDocument>(&message.topic.feature);

    let find_one_and_update_options = FindOneAndUpdateOptions::builder()
        .return_document(ReturnDocument::After)
        .build();

    let sensor_doc = collection
        .find_one_and_update(
            doc! { "uuid": &message.uuid, "apiToken": &message.api_token },
            doc! { "$set": {
                "value": message.payload.get_value(),
                "modifiedAt": DateTime::now()}},
            find_one_and_update_options,
        )
        .await.unwrap();

    // return result
    match sensor_doc {
        Some(sensor_doc) => {
            Ok(Some(document_to_json(&sensor_doc)))
        }
        None => {
            println!("none!!!!!");
            Ok(None)
        }
    }
}

fn document_to_json(sensor_doc: &SensorDocument) -> Sensor {
    Sensor {
        _id: sensor_doc._id.to_string(),
        uuid: sensor_doc.uuid.to_string(),
        mac: sensor_doc.mac.to_string(),
        manufacturer: sensor_doc.manufacturer.to_string(),
        model: sensor_doc.model.to_string(),
        profileOwnerId: sensor_doc.profileOwnerId.to_string(),
        apiToken: sensor_doc.apiToken.to_string(),
        createdAt: sensor_doc.createdAt.to_string(),
        modifiedAt: sensor_doc.modifiedAt.to_string(),
        value: sensor_doc.value,
    }
}