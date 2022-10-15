use log::{debug, error, info};
use std::fmt::Debug;

use mongodb::bson::{doc, DateTime};
use mongodb::options::FindOneAndUpdateOptions;
use mongodb::options::ReturnDocument;
use mongodb::Database;
use serde::Serialize;

use crate::models::message::Message;
use crate::models::payload_trait::PayloadTrait;
use crate::models::sensor::Sensor;
use crate::models::sensor::SensorDocument;

pub async fn update_message<T: PayloadTrait + Sized + Serialize + Debug>(
    db: &Database,
    message: &Message<T>,
) -> mongodb::error::Result<Option<Sensor>> {
    info!(target: "app", "update_message - Called with message = {:?}", message);

    let collection = db.collection::<SensorDocument>(&message.topic.feature);

    let find_one_and_update_options = FindOneAndUpdateOptions::builder()
        .return_document(ReturnDocument::After)
        .build();

    debug!(target: "app", "update_message - Finding and updating sensor type = {} with uuid = {}", &message.topic.feature, &message.uuid);

    let sensor_doc = collection
        .find_one_and_update(
            doc! { "uuid": &message.uuid, "apiToken": &message.api_token },
            doc! { "$set": {
                    "value": message.payload.get_value(),
                    "modifiedAt": DateTime::now()
                }
            },
            find_one_and_update_options,
        )
        .await
        .unwrap();

    // return result
    match sensor_doc {
        Some(sensor_doc) => Ok(Some(document_to_json(&sensor_doc))),
        None => {
            error!(target: "app", "update_message - Cannot find and update sensor with uuid = {}", &message.uuid);
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
