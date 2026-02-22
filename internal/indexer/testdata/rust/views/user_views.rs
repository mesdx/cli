use crate::services::user_service::{process_user_data, validate_email, format_user_name};

pub fn handle_update(user_id: u64, data: &str) -> bool {
    let result = process_user_data(user_id, data);
    result
}

pub fn handle_validate(email: &str) -> bool {
    validate_email(email)
}

pub fn handle_format(first: &str, last: &str) -> String {
    format_user_name(first, last)
}
