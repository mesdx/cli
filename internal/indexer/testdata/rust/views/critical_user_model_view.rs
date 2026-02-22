use crate::models::critical_user_model::CriticalUserModel;

/// CriticalUserModelRenderer renders a CriticalUserModel for display.
pub struct CriticalUserModelRenderer {
    pub model: CriticalUserModel,
}

/// render_critical_user_model formats a CriticalUserModel as a display string.
pub fn render_critical_user_model(m: &CriticalUserModel) -> String {
    format!("{} <{}>", m.name, m.email)
}

/// create_critical_user_model creates a CriticalUserModel with the given name and email.
pub fn create_critical_user_model(name: &str, email: &str) -> CriticalUserModel {
    CriticalUserModel::new(name, email)
}
