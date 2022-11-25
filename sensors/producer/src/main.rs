use dotenvy::dotenv;
use log::{debug, error, info};
use std::fs::File;
use std::io::Write;
use std::path::Path;
use std::string::ToString;
use std::{env, fs, thread};
use std::{process, time::Duration};

use futures::executor::block_on;
use lapin::Channel;
use paho_mqtt as mqtt;
use paho_mqtt::{ConnectOptions, SslOptions};

use producer::amqp::{amqp_connect, publish_message};
use producer::models::get_msq_byte;
use producer::models::topic::Topic;

const TOPICS: &[&str] = &[
    "sensors/+/temperature",
    "sensors/+/humidity",
    "sensors/+/light",
    "sensors/+/motion",
    "sensors/+/airquality",
    "sensors/+/airpressure",
];
const QOS: i32 = 0;

const COMBINED_CA_FILES_PATH: &str = "./rootca_and_cert.pem";

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
    let mqtt_url = env::var("MQTT_URL").expect("MQTT_URL is not found.");
    let mqtt_port = env::var("MQTT_PORT").expect("MQTT_PORT is not found.");
    let mqtt_client_id = env::var("MQTT_CLIENT_ID").expect("MQTT_CLIENT_ID is not found.");
    let mqtt_tls = env::var("MQTT_TLS").expect("MQTT_TLS is not found.");
    let root_ca = env::var("ROOT_CA").expect("ROOT_CA is not found.");
    let mqtt_cert_file = env::var("MQTT_CERT_FILE").expect("MQTT_CERT_FILE is not found.");
    let mqtt_key_file = env::var("MQTT_KEY_FILE").expect("MQTT_KEY_FILE is not found.");
    info!(target: "app", "AMQP_URI = {}", amqp_uri);
    info!(target: "app", "AMQP_QUEUE_NAME = {}", amqp_queue_name);
    info!(target: "app", "MQTT_URL = {}", mqtt_url);
    info!(target: "app", "MQTT_PORT = {}", mqtt_port);
    info!(target: "app", "MQTT_CLIENT_ID = {}", mqtt_client_id);
    info!(target: "app", "MQTT_TLS = {}", mqtt_tls);
    info!(target: "app", "ROOT_CA = {}", root_ca);
    info!(target: "app", "MQTT_CERT_FILE = {}", mqtt_cert_file);
    info!(target: "app", "MQTT_KEY_FILE = {}", mqtt_key_file);

    // 4. Init RabbitMQ
    info!(target: "app", "Initializing RabbitMQ");
    let channel: Channel = match amqp_connect(amqp_uri.as_str(), amqp_queue_name.as_str()).await {
        Ok(channel) => channel,
        Err(err) => {
            error!(target: "app", "Cannot create AMQP channel. Err = {:?}", err);
            return;
        }
    };

    // 5. Create CA file in 'COMBINED_CA_FILES_PATH'
    // merging 'root_ca' and 'mqtt_cert_file',
    // otherwise, paho.mqtt.rust won't be able to connect.
    if mqtt_tls == "true" {
        info!(target: "app", "Preparing MQTT CA file");
        merge_ca_files(&root_ca, &mqtt_cert_file);
    }

    // 6. Init MQTT
    let mqtt_uri = if mqtt_tls == "true" {
        format!("ssl://{}:{}", mqtt_url, mqtt_port)
    } else {
        format!("tcp://{}:{}", mqtt_url, mqtt_port)
    };
    info!(target: "app", "mqtt_uri = {}", mqtt_uri);

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

    info!(target: "app", "Creating MQTT ConnectOptions...");
    let conn_opts = build_mqtt_connect_options(&mqtt_tls, &mqtt_cert_file, &mqtt_key_file);

    // Make the connection to the broker
    info!(target: "app", "Connecting to the MQTT server with ConnectOptions...");
    mqtt_client.connect_with_callbacks(conn_opts, on_mqtt_connect_success, on_mqtt_connect_failure);

    // 6. Wait for incoming messages
    info!(target: "app", "Waiting for incoming MQTT messages");
    loop {
        thread::sleep(Duration::from_millis(1000));
    }
}

fn build_mqtt_connect_options(
    mqtt_tls: &String,
    mqtt_cert_file: &String,
    mqtt_key_file: &String,
) -> ConnectOptions {
    // Define the set of options for the connection
    let lwt = mqtt::Message::new("test", "Subscriber lost connection", 1);
    let conn_opts: ConnectOptions;
    let mut new_con_builder = mqtt::ConnectOptionsBuilder::new();
    let connect_options_builder = new_con_builder
        .keep_alive_interval(Duration::from_secs(20))
        .mqtt_version(mqtt::MQTT_VERSION_3_1_1)
        .clean_session(true)
        .will_message(lwt);

    if mqtt_tls == "true" {
        info!(target: "app", "build_mqtt_connect_options - MQTT TLS is enabled, creating ConnectOptions with certificates");
        let ssl_options = build_ssl_options(&mqtt_cert_file, mqtt_key_file);
        if let Some(ssl_opt) = ssl_options {
            conn_opts = connect_options_builder.ssl_options(ssl_opt).finalize();
            info!(target: "app", "build_mqtt_connect_options - MQTT ConnectOptions with SSL created successfully");
        } else {
            error!(target: "app", "build_mqtt_connect_options - Cannot create MQTT ConnectOptions with certificates.");
            process::exit(1);
        }
    } else {
        conn_opts = connect_options_builder.finalize();
        info!(target: "app", "build_mqtt_connect_options - MQTT ConnectOptions created WITHOUT SSL");
    };
    conn_opts
}

fn merge_ca_files(root_ca: &String, mqtt_cert_file: &String) {
    // re-create a new file appending two certificates:
    // - ROOT_CA file (ISRG_Root_X1.pem in case of Let's Encrypt)
    // - MQTT_CERT_FILE file (cert.pem in case of Let's Encrypt)
    if Path::new(COMBINED_CA_FILES_PATH).exists() {
        fs::remove_file(COMBINED_CA_FILES_PATH)
            .expect("Cannot remove existing COMBINED_CA_FILES_PATH");
    }
    let combined_root_ca_res = File::create(COMBINED_CA_FILES_PATH);
    match &combined_root_ca_res {
        Ok(_res) => {
            info!(target: "app", "merge_ca_files - {} file created", COMBINED_CA_FILES_PATH);
        }
        Err(err) => {
            error!(target: "app", "merge_ca_files - cannot create {} file, err = {:?}", COMBINED_CA_FILES_PATH, err);
        }
    }
    let mut combined_root_ca = combined_root_ca_res.unwrap();
    let root_ca_vec = fs::read(&root_ca).expect("Cannot read root_ca_vec as byte array");
    let mqtt_cert_file_vec =
        fs::read(&mqtt_cert_file).expect("Cannot read mqtt_cert_file as byte array");
    combined_root_ca
        .write_all(&root_ca_vec)
        .expect("Cannot write root_ca_vec to COMBINED_CA_FILES_PATH");
    combined_root_ca
        .write_all(b"\n")
        .expect("Cannot write new line to COMBINED_CA_FILES_PATH");
    combined_root_ca
        .write_all(&mqtt_cert_file_vec)
        .expect("Cannot write mqtt_cert_file_vec to COMBINED_CA_FILES_PATH");
}

fn build_ssl_options(mqtt_cert_file: &String, mqtt_key_file: &String) -> Option<SslOptions> {
    // I need COMBINED_CA_FILES_PATH, check above at step 5.
    let mut trust_store = env::current_dir().expect("Cannot get current dir for trust_store.");
    trust_store.push(COMBINED_CA_FILES_PATH);
    let mut key_store = env::current_dir().expect("Cannot get current dir for key_store.");
    key_store.push(mqtt_cert_file);
    let mut private_key = env::current_dir().expect("Cannot get current dir for private_key.");
    private_key.push(mqtt_key_file);
    if !trust_store.exists() {
        error!(target: "app", "get_ssl_options - trust_store file does not exist: {:?}", trust_store);
        return None;
    }
    if !key_store.exists() {
        error!(target: "app", "get_ssl_options - key_store file does not exist: {:?}", key_store);
        return None;
    }
    if !private_key.exists() {
        error!(target: "app", "get_ssl_options - private_key file does not exist: {:?}", private_key);
        return None;
    }

    debug!(target: "app", "get_ssl_options - trust_store {:?}", trust_store);
    debug!(target: "app", "get_ssl_options - key_store {:?}", key_store);
    debug!(target: "app", "get_ssl_options - private_key {:?}", private_key);

    let ssl_opts = mqtt::SslOptionsBuilder::new()
        .trust_store(trust_store)
        .unwrap()
        .key_store(key_store)
        .unwrap()
        .private_key(private_key)
        .unwrap()
        .finalize();
    Some(ssl_opts)
}
