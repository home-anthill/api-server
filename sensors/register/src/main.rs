#[macro_use]
extern crate rocket;

use dotenv::dotenv;
use rocket::{Build, Rocket};

use register::catchers::{self};
use register::routes;
use register::db;

#[launch]
fn rocket() -> Rocket<Build> {
    dotenv().ok();

    let our_rocket = rocket::build()
        .attach(db::init())
        .mount(
            "/",
            routes![
                routes::api::post_register_temperature,
                routes::api::post_register_humidity,
                routes::api::post_register_light,
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
        );
    our_rocket
}