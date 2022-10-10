use log::{error, info};
use std::env;

use mongodb::options::ClientOptions;
use mongodb::{Client, Database};
use rocket::fairing::AdHoc;

pub mod sensor;

pub fn init() -> AdHoc {
    AdHoc::on_ignite("Connecting to MongoDB", |rocket| async {
        match connect().await {
            Ok(database) => rocket.manage(database),
            Err(error) => {
                error!(target: "app", "MongoDB - cannot connect {:?}", error);
                panic!("Cannot connect to MongoDB:: {:?}", error)
            }
        }
    })
}

async fn connect() -> mongodb::error::Result<Database> {
    let mongo_uri = env::var("MONGO_URI").expect("MONGO_URI is not found.");
    let mongo_db_name = env::var("MONGO_DB_NAME").expect("MONGO_DB_NAME is not found.");

    let mut client_options = ClientOptions::parse(mongo_uri).await?;
    client_options.app_name = Some("register".to_string());
    let client = Client::with_options(client_options)?;
    let database = client.database(mongo_db_name.as_str());

    info!(target: "app", "MongoDB connected!");

    Ok(database)
}
