use log::{debug, error, info};
use rocket::http::Status;
use rocket::serde::json::{json, Json};
use rocket::State;

use mongodb::Database;

use crate::db::sensor;
use crate::errors::api_error::{ApiError, ApiResponse};
use crate::models::inputs::RegisterInput;

/// keepalive
#[get("/keepalive")]
pub async fn keep_alive() -> ApiResponse {
    info!(target: "app", "REST - GET - keep_alive");
    ApiResponse {
        json: json!({ "alive": true }),
        code: Status::Ok.code,
    }
}

/// register a new temperature sensor
#[post("/register/temperature", data = "<input>")]
pub async fn post_register_temperature(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_temperature");
    insert_register(db, input, "temperature").await
}

/// register a new humidity sensor
#[post("/register/humidity", data = "<input>")]
pub async fn post_register_humidity(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_humidity");
    insert_register(db, input, "humidity").await
}

/// register a new light sensor
#[post("/register/light", data = "<input>")]
pub async fn post_register_light(db: &State<Database>, input: Json<RegisterInput>) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_light");
    insert_register(db, input, "light").await
}

async fn insert_register(
    db: &State<Database>,
    input: Json<RegisterInput>,
    sensor_type: &str,
) -> ApiResponse {
    debug!(target: "app", "insert_register - called with sensor_type = {}", sensor_type);
    match sensor::insert_register(db, input, sensor_type).await {
        Ok(register_doc_id) => {
            debug!(target: "app", "insert_register - document inserted with id = {}", register_doc_id);
            ApiResponse {
                json: json!({ "id": register_doc_id }),
                code: Status::Ok.code,
            }
        }
        Err(error) => {
            error!(target: "app", "insert_register - error = {:?}", error);
            ApiResponse {
                json: serde_json::to_value(ApiError {
                    message: "Invalid input".to_string(),
                    code: Status::BadRequest.code,
                })
                .unwrap(),
                code: Status::BadRequest.code,
            }
        }
    }
}
