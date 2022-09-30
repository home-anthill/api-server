use std::{process, time::Duration};
use std::string::ToString;

use futures::{executor::block_on, stream::StreamExt, TryFutureExt};
use paho_mqtt as mqtt;
use lapin::Channel;

mod models;
mod amqp;

use crate::models::get_msq_byte;
use crate::models::message::Message;
use crate::models::notification::Notification;
use crate::models::topic::Topic;
use crate::amqp::{publish_message, create_connection, create_channel};

#[tokio::main]
async fn main() {
    println!("starting up");
    // AMQP (RABBIT MQ CONNECTION)
    let connection_res = create_connection().await;
    let connection = connection_res.unwrap();
    connection.on_error(|err| {
        println!("Connection error = {:?}", err);
    });

    let channel = create_channel(&connection).await;

    // ************************************************************************
    // ******************************** MQTT **********************************
    // ************************************************************************
    let host = "tcp://localhost:1883".to_string();

    // Topics to subscribe
    const TOPICS: &[&str] = &[
        "sensors/+/temperature",
        "sensors/+/humidity",
        "sensors/+/light",
    ];
    let topics: Vec<String> = TOPICS.iter().map(|s| s.to_string()).collect();

    // Create the client. Use an ID for a persistent session.
    // A real system should try harder to use a unique ID.
    let create_opts = mqtt::CreateOptionsBuilder::new()
        .server_uri(host)
        .client_id("rust_async_subscribe")
        .finalize();

    // Create the client connection
    let mut mqtt_client = mqtt::AsyncClient::new(create_opts).unwrap_or_else(|e| {
        println!("Error creating the client: {:?}", e);
        process::exit(1);
    });
    // ************************************************************************

    if let Err(err) = block_on(async {
        // Get message stream before connecting.
        let mut stream = mqtt_client.get_stream(25);

        // Define the set of options for the connection
        let lwt = mqtt::Message::new("test", "Async subscriber lost connection", mqtt::QOS_1);

        let conn_opts = mqtt::ConnectOptionsBuilder::new()
            .keep_alive_interval(Duration::from_secs(30))
            .mqtt_version(mqtt::MQTT_VERSION_3_1_1)
            .clean_session(false)
            .will_message(lwt)
            .finalize();

        // Make the connection to the broker
        println!("Connecting to the MQTT server...");
        mqtt_client.connect(conn_opts).await?;

        println!("Subscribing to topics: {:?}", topics);
        const QOS: &[i32] = &[1, 1, 1];
        mqtt_client.subscribe_many(&topics, QOS).await?;

        // Just loop on incoming messages.
        println!("Waiting for messages...");

        // Note that we're not providing a way to cleanly shut down and
        // disconnect. Therefore, when you kill this app (with a ^C or
        // whatever) the server will get an unexpected drop and then
        // should emit the LWT message.
        while let Some(msg_opt) = stream.next().await {
            println!("Checking msg_opt...");
            if let Some(msg) = msg_opt {
                println!("MQTT messaged received: {}", msg);

                let topic: Topic = Topic::new(msg.topic());
                println!("Topic = {}", &topic);

                let payload_str = match std::str::from_utf8(msg.payload()) {
                    Ok(res) => {
                        println!("payload_str: {}", res);
                        res
                    }
                    Err(err) => {
                        eprintln!("cannot read payload as utf8. Error = {}", err);
                        ""
                    }
                };

                let msg_byte: Vec<u8> = get_msq_byte(&topic, payload_str);
                if !msg_byte.is_empty() {
                    // send to RabbitMQ
                    println!("Sending message to RabbitMQ");
                    publish_message(&channel, msg_byte).await;
                }
            } else {
                // A "None" means we were disconnected. Try to reconnect...
                println!("Lost connection. Attempting reconnect.");
                while let Err(err) = mqtt_client.reconnect().await {
                    println!("Error reconnecting: {:?}", err);
                    // For tokio use: tokio::time::delay_for()
                    async_std::task::sleep(Duration::from_millis(1000)).await;
                    println!("Reconnecting sleep done");
                }
            }
        }

        // Explicit return type for the async block
        Ok::<(), mqtt::Error>(())
    }) {
        eprintln!("{}", err);
    }
}