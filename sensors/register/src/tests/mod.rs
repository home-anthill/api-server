use super::rocket;

use rocket::http::Status;
use rocket::local::blocking::Client;
use serde_json::{json, Value};

#[test]
fn hello_world() {
    let client = Client::tracked(rocket()).expect("valid rocket instance");
    let response = client.get("/keepalive").dispatch();
    assert_eq!(response.status(), Status::Ok);

    let resp: Value = response.into_json::<Value>().unwrap();

    let expected: Value = json!({ "alive": true });

    assert_eq!(resp, expected);
}
