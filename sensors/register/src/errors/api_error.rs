use rocket::request::Request;
use rocket::response::{Response, Responder, Result};
use rocket::http::{ContentType, Status};
use rocket::serde::{Serialize};
use rocket::serde::json::{Value, json};

#[derive(Debug)]
pub struct ApiResponse {
    pub json: Value,
    pub status: Status,
}

#[rocket::async_trait]
impl<'r> Responder<'r, 'r> for ApiResponse {
    fn respond_to(self, req: &'r Request<'_>) -> Result<'static> {
        Response::build_from(self.json.respond_to(&req).unwrap())
            .status(self.status)
            .header(ContentType::JSON)
            .ok()
    }
}

#[derive(Debug, Serialize)]
#[serde(crate = "rocket::serde")]
pub struct ApiError {
    pub code: u16,
    pub message: String,
}