pub const MAX_WORKERS: u32 = 4;

pub fn process_user_data(user_id: u64, data: &str) -> bool {
    !data.is_empty()
}

pub fn validate_email(email: &str) -> bool {
    email.contains('@')
}

pub fn format_user_name(first: &str, last: &str) -> String {
    format!("{} {}", first.trim(), last.trim())
}

pub struct UserRepository;

impl UserRepository {
    pub fn find_by_id(&self, user_id: u64) -> Option<String> {
        Some(format!("user-{}", user_id))
    }

    pub fn save(&self, user: &str) -> bool {
        !user.is_empty()
    }
}
