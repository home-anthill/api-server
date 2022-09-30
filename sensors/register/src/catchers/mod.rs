use rocket::http::Status;
use rocket::request::Request;

use crate::errors::api_error::{ApiError, ApiResponse};

#[catch(400)]
pub fn bad_request(_: &Request) -> ApiResponse {
    ApiResponse {
        json: serde_json::to_value(ApiError { code: 0, message: "Bad request".to_string() }).unwrap(),
        status: Status::BadRequest,
    }
}

#[catch(404)]
pub fn not_found(_: &Request) -> ApiResponse {
    ApiResponse {
        json: serde_json::to_value(ApiError { code: 0, message: "Not found".to_string() }).unwrap(),
        status: Status::NotFound,
    }
}

#[catch(422)]
pub fn unprocessable_entity(r: &Request) -> ApiResponse {
    ApiResponse {
        json: serde_json::to_value(ApiError { code: 0, message: "Unprocessable entity".to_string() }).unwrap(),
        status: Status::UnprocessableEntity,
    }
}

#[catch(500)]
pub fn internal_server_error(_: &Request) -> ApiResponse {
    ApiResponse {
        json: serde_json::to_value(ApiError { code: 0, message: "Something went wrong".to_string() }).unwrap(),
        status: Status::InternalServerError,
    }
}
