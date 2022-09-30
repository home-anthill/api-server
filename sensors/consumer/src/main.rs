use dotenv::dotenv;
use futures_lite::StreamExt; // required to use 'consumer.next()'
use lapin::{Consumer};

use consumer::models::message::{GenericMessage, Message};
use consumer::models::payload_trait::{Humidity, Light, Temperature};
use consumer::amqp::{init, read_message};
use consumer::db::{connect};
use consumer::db::sensor::update_message;
use consumer::models::{new_humidity_message, new_temperature_message, new_light_message};

#[tokio::main]
async fn main() {
    println!("starting up");
    dotenv().ok();

    let client = connect().await.unwrap();
    let mut consumer: Consumer = init().await;

    while let Some(delivery) = consumer.next().await {
        if let Ok(delivery) = delivery {
            let payload_str: &str = read_message(&delivery).await;

            // deserialize to a GenericMessage (with turbofish operator "::<GenericMessage>")
            match serde_json::from_str::<GenericMessage>(payload_str) {
                Ok(generic_msg) => {
                    println!("GenericMessage deserialized from JSON = {:?}", generic_msg);

                    if generic_msg.topic.feature == "temperature"  {
                        let message: Message<Temperature> = new_temperature_message(generic_msg);
                        println!("message temperature {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                println!("update success {:?}", _register_doc_id);
                            }
                            Err(_error) => {
                                println!("{:?}", _error);
                            }
                        };
                    } else if generic_msg.topic.feature == "humidity" {
                        let message: Message<Humidity> = new_humidity_message(generic_msg);
                        println!("message humidity {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                println!("update success");
                            }
                            Err(_error) => {
                                println!("{:?}", _error);
                            }
                        }
                    } else if generic_msg.topic.feature == "light" {
                        let message: Message<Light> = new_light_message(generic_msg);
                        println!("message light {:?}", &message);
                        match update_message(&client, &message).await {
                            Ok(_register_doc_id) => {
                                println!("update success");
                            }
                            Err(_error) => {
                                println!("{:?}", _error);
                            }
                        }
                    } else {
                        eprintln!("Cannot recognize Message payload type");
                    }
                }
                Err(err) => {
                    eprintln!("Cannot convert payload as json Message. Error = {:?}", err);
                }
            }
        } else {
            println!("delivery None = {:?}", delivery);
        }
    }
}