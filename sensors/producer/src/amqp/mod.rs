use log::{debug, error, info};

use lapin::publisher_confirm::PublisherConfirm;
use lapin::{
    options::{BasicPublishOptions, QueueDeclareOptions},
    types::FieldTable,
    BasicProperties, Channel, Connection, ConnectionProperties, Queue, Result,
};

pub async fn amqp_connect(amqp_uri: &str, amqp_queue_name: &str) -> Result<Channel> {
    // Create connection
    info!(target: "app", "amqp_connect - trying to connect via AMQP...");
    let connection_result = create_connection(amqp_uri).await;
    let connection = connection_result.unwrap();
    connection.on_error(|err| {
        error!(target: "app", "amqp_connect - AMQP connection error = {:?}", err);
    });
    // Create channel
    let channel_result = create_channel(&connection).await;
    let channel: Channel = match channel_result {
        Ok(channel) => channel,
        Err(err) => {
            error!(target: "app", "amqp_connect - cannot create AMQP channel");
            return Err(err);
        }
    };
    // Create queue
    let queue_result = create_queue(&channel, amqp_queue_name).await;
    let queue: Queue = match queue_result {
        Ok(queue) => queue,
        Err(err) => {
            error!(target: "app", "amqp_connect - cannot create AMQP queue");
            return Err(err);
        }
    };
    debug!(target: "app", "amqp_connect - declared queue = {:?}", queue);
    info!(target: "app", "amqp_connect - AMQP connection success, returning channel");
    Ok(channel)
}

pub async fn publish_message(channel: &Channel, amqp_queue_name: &str, msg_byte: Vec<u8>) {
    debug!(target: "app", "publish_message - publishing byte message to queue {}", amqp_queue_name);
    let publisher_result: Result<PublisherConfirm> = channel
        .basic_publish(
            "",
            amqp_queue_name,
            BasicPublishOptions::default(),
            msg_byte.as_slice(),
            BasicProperties::default(),
        )
        .await;
    if publisher_result.is_err() {
        error!(target: "app", "publish_message - cannot publish AMQP message to queue {}. Err = {:?}", amqp_queue_name, publisher_result.err());
    }
}

async fn create_connection(amqp_uri: &str) -> Result<Connection> {
    // Use tokio executor and reactor.
    // At the moment the reactor is only available for unix.
    let options = ConnectionProperties::default()
        .with_executor(tokio_executor_trait::Tokio::current())
        .with_reactor(tokio_reactor_trait::Tokio);
    Connection::connect(amqp_uri, options).await
}

async fn create_channel(connection: &Connection) -> Result<Channel> {
    connection.create_channel().await
}

async fn create_queue(channel: &Channel, amqp_queue_name: &str) -> Result<Queue> {
    channel
        .queue_declare(
            amqp_queue_name,
            QueueDeclareOptions::default(),
            FieldTable::default(),
        )
        .await
}
