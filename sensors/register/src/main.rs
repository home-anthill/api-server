#[macro_use]
extern crate rocket;

use dotenvy::dotenv;
use log::info;
use rocket::{Build, Rocket};
use std::env;

use register::catchers;
use register::db;
use register::routes;

#[launch]
fn rocket() -> Rocket<Build> {
    // 1. Init logger
    log4rs::init_file("log4rs.yaml", Default::default()).unwrap();
    info!(target: "app", "Starting application...");

    // 2. Load the .env file
    dotenv().ok();

    // 3. Print .env vars
    info!(target: "app", "MONGO_URI = {}", env::var("MONGO_URI").expect("MONGO_URI is not found."));
    info!(target: "app", "MONGO_DB_NAME = {}", env::var("MONGO_DB_NAME").expect("MONGO_DB_NAME is not found."));

    // 4. Init Rocket
    // a) connect to DB
    // b) define APIs
    // c) define error handlers
    info!(target: "app", "Starting Rocket");
    rocket::build()
        .attach(db::init())
        .mount(
            "/",
            routes![
                routes::api::post_register_temperature,
                routes::api::post_register_humidity,
                routes::api::post_register_light,
                routes::api::post_register_motion,
                routes::api::post_register_airquality,
                routes::api::post_register_airpressure,
                routes::api::get_sensor_value,
                routes::api::keep_alive,
            ],
        )
        .register(
            "/",
            catchers![
                catchers::bad_request,
                catchers::not_found,
                catchers::unprocessable_entity,
                catchers::internal_server_error,
            ],
        )
}

// Unit testings
#[cfg(test)]
mod tests;
