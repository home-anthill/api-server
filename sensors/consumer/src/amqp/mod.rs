use lapin::{
    message::Delivery,
    options::{BasicAckOptions, BasicConsumeOptions, QueueDeclareOptions},
    types::FieldTable, Channel, Connection, ConnectionProperties, Consumer,
};

const QUEUE_NAME: &str = "queue_test";

pub async fn init() -> Consumer {
    let uri = "amqp://localhost:5672";
    let options = ConnectionProperties::default()
        // Use tokio executor and reactor.
        // At the moment the reactor is only available for unix.
        .with_executor(tokio_executor_trait::Tokio::current())
        .with_reactor(tokio_reactor_trait::Tokio);

    let connection: Connection = Connection::connect(uri, options).await.unwrap();
    println!("CONNECTED");

    let channel: Channel = connection.create_channel().await.unwrap();
    println!("conn status state: {:?}", connection.status().state());

    let _queue = channel
        .queue_declare(
            QUEUE_NAME,
            QueueDeclareOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();

    println!("Declared queue {:?}", _queue);

    let consumer = channel
        .basic_consume(
            QUEUE_NAME,
            "tag_foo",
            BasicConsumeOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();
    println!("conn status state: {:?}", connection.status().state());

    consumer
}

pub async fn read_message(delivery: &Delivery) -> &str {
    delivery
        .ack(BasicAckOptions::default())
        .await
        .expect("basic_ack");

    let payload_str = match std::str::from_utf8(&delivery.data) {
        Ok(res) => {
            println!("payload_str: {}", res);
            res
        }
        Err(err) => {
            eprintln!("cannot read payload as utf8. Error = {}", err);
            ""
        }
    };
    payload_str
}