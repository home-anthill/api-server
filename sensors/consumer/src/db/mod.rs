
use mongodb::options::ClientOptions;
use mongodb::{Client, Database};
use std::env;

pub mod sensor;

pub async fn connect() -> mongodb::error::Result<Database> {
    let mongo_uri = env::var("MONGO_URI").expect("MONGO_URI is not found.");
    let mongo_db_name = env::var("MONGO_DB_NAME").expect("MONGO_DB_NAME is not found.");

    let client_options = ClientOptions::parse(mongo_uri).await?;
    let client = Client::with_options(client_options)?;
    let database = client.database(mongo_db_name.as_str());

    println!("MongoDB Connected!");

    Ok(database)
}