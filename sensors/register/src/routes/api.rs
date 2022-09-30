use rocket::serde::json::{Json, json};
use rocket::State;
use rocket::http::{Status};

use mongodb::Database;

use crate::errors::api_error::{ApiResponse, ApiError};
use crate::models::inputs::RegisterInput;
use crate::db::sensor;

/// register a new temperature sensor
#[post("/register/temperature", data = "<input>")]
pub async fn post_register_temperature(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    insert_register(db, input, "temperature").await
}

/// register a new humidty sensor
#[post("/register/humidity", data = "<input>")]
pub async fn post_register_humidity(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    insert_register(db, input, "humidity").await
}

/// register a new light sensor
#[post("/register/light", data = "<input>")]
pub async fn post_register_light(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    insert_register(db, input, "light").await
}

/// keepalive
#[get("/keepalive")]
pub async fn keep_alive() -> &'static str {
    "ok"
}

async fn insert_register(
    db: &State<Database>,
    input: Json<RegisterInput>,
    sensor_type: &str
) -> ApiResponse {
    // can set with a single error like this.
    match sensor::insert_register(db, input, sensor_type).await {
        Ok(_register_doc_id) => {
            ApiResponse {
                json: json!({ "id": _register_doc_id }),
                status: Status::Ok,
            }
        }
        Err(_error) => {
            println!("{:?}", _error);
            ApiResponse {
                json: serde_json::to_value(ApiError { code: 0, message: "Invalid input".to_string() }).unwrap(),
                status: Status::BadRequest,
            }
        }
    }
}