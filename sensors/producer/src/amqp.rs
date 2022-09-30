use std::future::Future;
use lapin::{Result, options::{BasicPublishOptions, QueueDeclareOptions}, types::FieldTable, Channel, BasicProperties, Connection, ConnectionProperties, Queue};

const QUEUE_NAME: &str = "queue_test";

// fn retry_rabbit_stuff() {
//     std::thread::sleep(std::time::Duration::from_millis(2000));
//     println!("Reconnecting to rabbitmq");
//     try_rabbit_stuff();
// }
//
// pub fn try_rabbit_stuff() {
//     let returned = async_global_executor::spawn(async {
//         println!("In async_global_executor");
//         let result: Result<Connection> = match create_connection().await {
//             Err(err) => {
//                 println!("Error: {}", err);
//                 // retry_rabbit_stuff();
//                 Err(err)
//             }
//             Ok(conn) => {
//                 Ok(conn)
//             }
//         };
//         result
//     });
//     println!("retruned = {:?}", &returned);
//     returned.detach();
// }

pub async fn create_connection() -> Result<Connection> {
    let uri = "amqp://localhost:5672";
    let options = ConnectionProperties::default()
        // Use tokio executor and reactor.
        // At the moment the reactor is only available for unix.
        .with_executor(tokio_executor_trait::Tokio::current())
        .with_reactor(tokio_reactor_trait::Tokio);

    let connection = Connection::connect(uri, options).await.unwrap();
    Ok(connection)
}

pub async fn create_channel(connection: &Connection) -> Channel {
    let channel = connection.create_channel().await.unwrap();

    let _queue = channel
        .queue_declare(
            QUEUE_NAME,
            QueueDeclareOptions::default(),
            FieldTable::default(),
        )
        .await.unwrap();
    channel
}

pub async fn publish_message(channel: &Channel, msg_byte: Vec<u8>) {
    // send to RabbitMQ
    channel.basic_publish(
        "",
        QUEUE_NAME,
        BasicPublishOptions::default(),
        msg_byte.as_slice(),
        BasicProperties::default(),
    ).await.unwrap().await.unwrap();
}