use dotenvy::dotenv;
use log::{debug, error, info};
use std::string::ToString;
use std::{env, thread};
use std::{process, time::Duration};

use futures::executor::block_on;
use lapin::Channel;
use paho_mqtt as mqtt;

use producer::amqp::{amqp_connect, publish_message};
use producer::models::get_msq_byte;
use producer::models::topic::Topic;

const TOPICS: &[&str] = &[
    "sensors/+/temperature",
    "sensors/+/humidity",
    "sensors/+/light",
    "sensors/+/motion",
];
const QOS: i32 = 0;

fn on_mqtt_connect_success(cli: &mqtt::AsyncClient, _msgid: u16) {
    info!(target: "app", "MQTT Connection succeeded");
    let topics: Vec<String> = TOPICS.iter().map(|s| s.to_string()).collect();
    info!(target: "app", "Subscribing to MQTT topics: {:?}", topics);
    let qos = vec![QOS; topics.len()];
    // We subscribe to the topic(s) we want here.
    cli.subscribe_many(&topics, &qos);
}

// Callback for a failed attempt to connect to the server. We simply sleep and then try again.
// Note that normally we don't want to do a blocking operation or sleep
// from  within a callback. But in this case, we know that the client is
// *not* connected, and thus not doing anything important. So we don't worry
// too much about stopping its callback thread.
fn on_mqtt_connect_failure(cli: &mqtt::AsyncClient, _msgid: u16, rc: i32) {
    error!(target: "app", "MQTT connection attempt failed with error code {}", rc);
    thread::sleep(Duration::from_millis(5000));
    cli.reconnect_with_callbacks(on_mqtt_connect_success, on_mqtt_connect_failure);
}

#[tokio::main]
async fn main() {
    // 1. Init logger
    log4rs::init_file("log4rs.yaml", Default::default()).unwrap();
    info!(target: "app", "Starting application...");

    // 2. Load the .env file
    dotenv().ok();

    // 3. Print .env vars
    let amqp_uri = env::var("AMQP_URI").expect("AMQP_URI is not found.");
    let amqp_queue_name = env::var("AMQP_QUEUE_NAME").expect("AMQP_QUEUE_NAME is not found.");
    let mqtt_uri = env::var("MQTT_URI").expect("MQTT_URI is not found.");
    let mqtt_client_id = env::var("MQTT_CLIENT_ID").expect("MQTT_CLIENT_ID is not found.");
    info!(target: "app", "AMQP_URI = {}", amqp_uri);
    info!(target: "app", "AMQP_QUEUE_NAME = {}", amqp_queue_name);
    info!(target: "app", "MQTT_URI = {}", mqtt_uri);
    info!(target: "app", "MQTT_CLIENT_ID = {}", mqtt_client_id);

    // 4. Init RabbitMQ
    info!(target: "app", "Initializing RabbitMQ");
    let channel: Channel = match amqp_connect(amqp_uri.as_str(), amqp_queue_name.as_str()).await {
        Ok(channel) => channel,
        Err(err) => {
            error!(target: "app", "Cannot create AMQP channel. Err = {:?}", err);
            return;
        }
    };

    // 5. Init MQTT
    // Create the client. Use an unique ID for a persistent session
    let create_opts = mqtt::CreateOptionsBuilder::new()
        .server_uri(mqtt_uri)
        .client_id(mqtt_client_id)
        .finalize();
    let mqtt_client = mqtt::AsyncClient::new(create_opts).unwrap_or_else(|err| {
        error!(target: "app", "Error creating MQTT client: {:?}", err);
        process::exit(1);
    });
    // Define all callbacks
    mqtt_client.set_connected_callback(|_cli: &mqtt::AsyncClient| {
        info!(target:"app", "MQTT Connected");
    });
    mqtt_client.set_connection_lost_callback(|client: &mqtt::AsyncClient| {
        // Whenever the client loses the connection.
        // It will attempt to reconnect, and set up function callbacks to keep
        // retrying until the connection is re-established.
        error!(target:"app", "MQTT Connection lost! Attempting reconnect after a delay.");
        // tokio::time::sleep(Duration::from_millis(25000)).await;
        // async_std::task::sleep(Duration::from_millis(1000)).await;
        thread::sleep(Duration::from_millis(5000));
        client.reconnect_with_callbacks(on_mqtt_connect_success, on_mqtt_connect_failure);
    });
    mqtt_client.set_message_callback(move |_cli, msg| {
        if let Some(msg) = msg {
            debug!(target: "app", "MQTT message received");
            let topic: Topic = Topic::new(msg.topic());
            debug!(target: "app", "MQTT message topic = {}", &topic);
            let payload_str = match std::str::from_utf8(msg.payload()) {
                Ok(res) => {
                    debug!(target: "app", "MQTT utf8 payload_str: {}", res);
                    res
                }
                Err(err) => {
                    error!(target: "app", "Cannot read MQTT message payload as utf8. Error = {:?}", err);
                    ""
                }
            };
            let msg_byte: Vec<u8> = get_msq_byte(&topic, payload_str);
            if !msg_byte.is_empty() {
                // send via AMQP
                debug!(target: "app", "Publishing message via AMQP...");
                block_on(async {
                    publish_message(&channel, amqp_queue_name.as_str(), msg_byte).await;
                    info!(target: "app", "AMQP message published");
                });
            }
        } else {
            error!(target: "app", "MQTT message is not valid");
        }
    });
    // Define the set of options for the connection
    let lwt = mqtt::Message::new("test", "Subscriber lost connection", 1);
    let conn_opts = mqtt::ConnectOptionsBuilder::new()
        .keep_alive_interval(Duration::from_secs(20))
        .mqtt_version(mqtt::MQTT_VERSION_3_1_1)
        .clean_session(true)
        .will_message(lwt)
        .finalize();
    // Make the connection to the broker
    info!(target: "app", "Connecting to the MQTT server...");
    mqtt_client.connect_with_callbacks(conn_opts, on_mqtt_connect_success, on_mqtt_connect_failure);

    // 6. Wait for incoming messages
    loop {
        thread::sleep(Duration::from_millis(1000));
    }
}
