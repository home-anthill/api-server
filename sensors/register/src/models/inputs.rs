use serde::{Deserialize, Serialize};

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct RegisterInput {
    pub uuid: String,
    pub mac: String,
    pub manufacturer: String,
    pub model: String,
    pub profileOwnerId: String,
    pub apiToken: String,
}
