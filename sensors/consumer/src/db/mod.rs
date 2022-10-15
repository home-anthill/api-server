use log::info;
use std::env;

use mongodb::options::ClientOptions;
use mongodb::{Client, Database};

pub mod sensor;

pub async fn connect(mongo_uri: String, mongo_db_name: String) -> mongodb::error::Result<Database> {
    let mut client_options = ClientOptions::parse(mongo_uri).await?;
    client_options.app_name = Some("register".to_string());
    let client = Client::with_options(client_options)?;
    let database = client.database(mongo_db_name.as_str());

    info!(target: "app", "MongoDB connected!");

    Ok(database)
}
