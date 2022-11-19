use dotenvy::dotenv;
use futures_lite::StreamExt;
use lapin::{Channel, Consumer};
use log::{debug, error, info};
use std::env;

use consumer::amqp::{amqp_connect, read_message};
use consumer::db::connect;
use consumer::db::sensor::update_message;
use consumer::models::message::{GenericMessage, Message};
use consumer::models::payload_trait::{AirQuality, Humidity, Light, Motion, Temperature};
use consumer::models::{
    new_airquality_message, new_humidity_message, new_light_message, new_motion_message,
    new_temperature_message,
};

#[tokio::main]
async fn main() {
    // 1. Init logger
    log4rs::init_file("log4rs.yaml", Default::default()).unwrap();
    info!(target: "app", "Starting application...");

    // 2. Load the .env file
    dotenv().ok();

    // 3. Print .env vars
    let mongo_uri = env::var("MONGO_URI").expect("MONGO_URI is not found.");
    let mongo_db_name = env::var("MONGO_DB_NAME").expect("MONGO_DB_NAME is not found.");
    let amqp_uri = env::var("AMQP_URI").expect("AMQP_URI is not found.");
    let amqp_queue_name = env::var("AMQP_QUEUE_NAME").expect("AMQP_QUEUE_NAME is not found.");
    let amqp_consumer_tag = env::var("AMQP_CONSUMER_TAG").expect("AMQP_CONSUMER_TAG is not found.");
    info!(target: "app", "MONGO_URI = {}", env::var("MONGO_URI").expect("MONGO_URI is not found."));
    info!(target: "app", "MONGO_DB_NAME = {}", env::var("MONGO_DB_NAME").expect("MONGO_DB_NAME is not found."));
    info!(target: "app", "AMQP_URI = {}", amqp_uri);
    info!(target: "app", "AMQP_QUEUE_NAME = {}", amqp_queue_name);
    info!(target: "app", "AMQP_CONSUMER_TAG = {}", amqp_consumer_tag);

    // 4. Init MongoDB
    let client = match connect(mongo_uri, mongo_db_name).await {
        Ok(database) => database,
        Err(error) => {
            error!(target: "app", "MongoDB - cannot connect {:?}", error);
            panic!("Cannot connect to MongoDB:: {:?}", error)
        }
    };

    // 5. Init RabbitMQ
    let mut consumer: Consumer = match amqp_connect(
        amqp_uri.as_str(),
        amqp_queue_name.as_str(),
        amqp_consumer_tag.as_str(),
    )
    .await
    {
        Ok(consumer) => consumer,
        Err(err) => {
            error!(target: "app", "Cannot create AMQP consumer. Err = {:?}", err);
            return;
        }
    };
    while let Some(delivery_res) = consumer.next().await {
        if let Ok(delivery) = delivery_res {
            let payload_str: &str = read_message(&delivery).await;
            // deserialize to a GenericMessage (with turbofish operator "::<GenericMessage>")
            match serde_json::from_str::<GenericMessage>(payload_str) {
                Ok(generic_msg) => {
                    debug!(target: "app", "GenericMessage deserialized from JSON = {:?}", generic_msg);

                    if generic_msg.topic.feature == "temperature" {
                        let message: Message<Temperature> = new_temperature_message(generic_msg);
                        debug!(target: "app", "message temperature {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                debug!(target: "app", "update success {:?}", _register_doc_id);
                            }
                            Err(_error) => {
                                error!(target: "app", "{:?}", _error);
                            }
                        };
                    } else if generic_msg.topic.feature == "humidity" {
                        let message: Message<Humidity> = new_humidity_message(generic_msg);
                        debug!(target: "app", "message humidity {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                debug!(target: "app", "update success");
                            }
                            Err(_error) => {
                                error!(target: "app", "{:?}", _error);
                            }
                        }
                    } else if generic_msg.topic.feature == "light" {
                        let message: Message<Light> = new_light_message(generic_msg);
                        debug!(target: "app", "message light {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                debug!(target: "app", "update success {:?}", _register_doc_id);
                            }
                            Err(_error) => {
                                error!(target: "app", "{:?}", _error);
                            }
                        };
                    } else if generic_msg.topic.feature == "motion" {
                        let message: Message<Motion> = new_motion_message(generic_msg);
                        debug!(target: "app", "message motion {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                debug!(target: "app", "update success {:?}", _register_doc_id);
                            }
                            Err(_error) => {
                                error!(target: "app", "{:?}", _error);
                            }
                        };
                    } else if generic_msg.topic.feature == "airquality" {
                        let message: Message<AirQuality> = new_airquality_message(generic_msg);
                        debug!(target: "app", "message airquality {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                debug!(target: "app", "update success {:?}", _register_doc_id);
                            }
                            Err(_error) => {
                                error!(target: "app", "{:?}", _error);
                            }
                        };
                    } else {
                        error!(target: "app", "Cannot recognize Message payload type");
                    }
                }
                Err(err) => {
                    error!(target: "app", "Cannot convert payload as json Message. Error = {:?}", err);
                }
            }
        } else {
            error!(target: "app", "AMQP consumer - delivery_res error = {:?}", delivery_res.err());
        }
    }
}
