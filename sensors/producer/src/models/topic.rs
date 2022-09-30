use std::fmt;

use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct Topic {
    pub family: String,
    pub device_id: String,
    pub feature: String,
}

impl Topic {
    pub fn new(topic: &str) -> Self {
        let items: Vec<&str> = topic.split('/').collect();
        Self {
            family: items.first().unwrap().to_string(),
            device_id: items.get(1).unwrap().to_string(),
            feature: items.last().unwrap().to_string(),
        }
    }
}

impl fmt::Display for Topic {
    fn fmt(&self, fmt: &mut fmt::Formatter) -> fmt::Result {
        fmt.write_str(self.family.as_str())?;
        fmt.write_str("/")?;
        fmt.write_str(self.device_id.as_str())?;
        fmt.write_str("/")?;
        fmt.write_str(self.feature.as_str())?;
        Ok(())
    }
}
