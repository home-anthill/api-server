use log::error;

use rocket::http::Status;
use rocket::request::Request;

use crate::errors::api_error::ApiError;

#[catch(400)]
pub fn bad_request(_: &Request) -> ApiError {
    error!(target: "app", "catcher 400 - bad_request");
    ApiError {
        code: Status::BadRequest.code,
        message: "Bad request".to_string(),
    }
}

#[catch(404)]
pub fn not_found(_: &Request) -> ApiError {
    error!(target: "app", "catcher 404 - not_found");
    ApiError {
        code: Status::NotFound.code,
        message: "Not found".to_string(),
    }
}

#[catch(422)]
pub fn unprocessable_entity(_: &Request) -> ApiError {
    error!(target: "app", "catcher 422 - unprocessable_entity");
    ApiError {
        code: Status::UnprocessableEntity.code,
        message: "Unprocessable entity".to_string(),
    }
}

#[catch(500)]
pub fn internal_server_error(_: &Request) -> ApiError {
    error!(target: "app", "catcher 500 - internal_server_error");
    ApiError {
        code: Status::InternalServerError.code,
        message: "Internal server error".to_string(),
    }
}
