use std::fmt;

/// AppConfig holds application configuration.
/// Supports multiple environments.
pub struct AppConfig {
    pub host: String,
    pub port: u16,
    pub log_level: String,
}

impl AppConfig {
    /// Creates a new AppConfig with defaults.
    pub fn new(host: String) -> Self {
        AppConfig {
            host,
            port: 8080,
            log_level: "info".to_string(),
        }
    }

    /// Validates the configuration.
    pub fn validate(&self) -> Result<(), String> {
        if self.host.is_empty() {
            return Err("host is required".to_string());
        }
        if self.port == 0 {
            return Err("port must be positive".to_string());
        }
        Ok(())
    }
}

/// Formatter trait for output formatting.
pub trait AppFormatter {
    fn format(&self, data: &str) -> String;
    fn reset(&mut self);
}

/// Application status.
#[derive(Debug, Clone)]
pub enum AppStatus {
    Active,
    Inactive,
    Maintenance,
}

/// Default port constant.
const DEFAULT_PORT: u16 = 8080;

/// Application name.
type AppName = String;
