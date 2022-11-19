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
#[post("/sensors/register/temperature", data = "<input>")]
pub async fn post_register_temperature(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_temperature");
    insert_register(db, input, "temperature").await
}

/// register a new humidity sensor
#[post("/sensors/register/humidity", data = "<input>")]
pub async fn post_register_humidity(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_humidity");
    insert_register(db, input, "humidity").await
}

/// register a new light sensor
#[post("/sensors/register/light", data = "<input>")]
pub async fn post_register_light(db: &State<Database>, input: Json<RegisterInput>) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_light");
    insert_register(db, input, "light").await
}

/// register a new motion sensor
#[post("/sensors/register/motion", data = "<input>")]
pub async fn post_register_motion(db: &State<Database>, input: Json<RegisterInput>) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_motion");
    insert_register(db, input, "motion").await
}

/// register a new airquality sensor
#[post("/sensors/register/airquality", data = "<input>")]
pub async fn post_register_airquality(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_airquality");
    insert_register(db, input, "airquality").await
}

/// register a new airpressure sensor
#[post("/sensors/register/airpressure", data = "<input>")]
pub async fn post_register_airpressure(
    db: &State<Database>,
    input: Json<RegisterInput>,
) -> ApiResponse {
    info!(target: "app", "REST - POST - post_register_airpressure");
    insert_register(db, input, "airpressure").await
}

/// get sensor value by UUID and type
#[get("/sensors/<uuid>/<sensor_type>")]
pub async fn get_sensor_value(
    db: &State<Database>,
    uuid: String,
    sensor_type: String,
) -> ApiResponse {
    info!(target: "app", "REST - GET - get_sensor_value");
    get_sensor(db, uuid, sensor_type).await
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

async fn get_sensor(db: &State<Database>, uuid: String, sensor_type: String) -> ApiResponse {
    debug!(target: "app", "get_sensor - called with sensor_type = {}, uuid = {}", sensor_type, uuid);
    match sensor::get_sensor(db, &uuid, &sensor_type).await {
        Ok(sensor_doc_opt) => {
            debug!(target: "app", "get_sensor - result sensor_doc_opt = {:?}", sensor_doc_opt);
            match sensor_doc_opt {
                Some(sensor_doc) => {
                    info!(target: "app", "get_sensor - result sensor_doc = {}", sensor_doc);
                    ApiResponse {
                        json: json!(sensor_doc),
                        code: Status::Ok.code,
                    }
                }
                None => {
                    error!(target: "app", "get_sensor - sensor with uuid = {} not found", &uuid);
                    ApiResponse {
                        json: serde_json::to_value(ApiError {
                            message: "Sensor not found".to_string(),
                            code: Status::NotFound.code,
                        })
                        .unwrap(),
                        code: Status::NotFound.code,
                    }
                }
            }
        }
        Err(error) => {
            error!(target: "app", "get_sensor - error {:?}", error);
            ApiResponse {
                json: serde_json::to_value(ApiError {
                    message: "Internal server error".to_string(),
                    code: Status::InternalServerError.code,
                })
                .unwrap(),
                code: Status::InternalServerError.code,
            }
        }
    }
}
