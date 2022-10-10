use rocket::http::{ContentType, Status};
use rocket::request::Request;
use rocket::response::{Responder, Response, Result};
use rocket::serde::json::Value;
use rocket::serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct ApiResponse {
    pub json: Value,
    pub code: u16,
}

#[rocket::async_trait]
impl<'r> Responder<'r, 'r> for ApiResponse {
    fn respond_to(self, req: &'r Request<'_>) -> Result<'static> {
        Response::build_from(self.json.respond_to(req).unwrap())
            .status(Status { code: self.code })
            .header(ContentType::JSON)
            .ok()
    }
}

#[derive(Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(crate = "rocket::serde")]
pub struct ApiError {
    pub message: String,
    pub code: u16,
}

#[rocket::async_trait]
impl<'r> Responder<'r, 'r> for ApiError {
    fn respond_to(self, req: &'r Request<'_>) -> Result<'static> {
        Response::build_from(self.message.respond_to(req).unwrap())
            .status(Status { code: self.code })
            .header(ContentType::JSON)
            .ok()
    }
}
