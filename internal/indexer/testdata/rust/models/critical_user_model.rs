/// BaseModel provides base entity identity fields.
pub struct BaseModel {
    pub id: u64,
    pub created_at: String,
}

/// CriticalUserModel is a heavily-used user representation that many files depend on.
pub struct CriticalUserModel {
    pub base: BaseModel,
    pub name: String,
    pub email: String,
}

impl CriticalUserModel {
    /// Creates a new CriticalUserModel.
    pub fn new(name: &str, email: &str) -> CriticalUserModel {
        CriticalUserModel {
            base: BaseModel { id: 0, created_at: String::new() },
            name: name.to_string(),
            email: email.to_string(),
        }
    }

    /// Validates the model fields.
    pub fn validate(&self) -> bool {
        !self.name.is_empty() && self.email.contains('@')
    }
}
