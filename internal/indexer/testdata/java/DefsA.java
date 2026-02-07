package com.example;

import java.util.List;

/**
 * AppConfig holds application configuration.
 * Supports multiple environments.
 */
public class AppConfig {
    private String host;
    private int port;

    /**
     * Create a new AppConfig.
     */
    public AppConfig(String host, int port) {
        this.host = host;
        this.port = port;
    }

    /**
     * Returns the host.
     */
    public String getHost() {
        return host;
    }

    /**
     * Returns the port.
     */
    public int getPort() {
        return port;
    }

    public void validate() {
        if (host == null || host.isEmpty()) {
            throw new IllegalArgumentException("host is required");
        }
        if (port <= 0) {
            throw new IllegalArgumentException("port must be positive");
        }
    }
}

/**
 * Formatter interface for output formatting.
 */
interface AppFormatter {
    String format(String data);
    void reset();
}

@Deprecated
enum AppStatus {
    ACTIVE,
    INACTIVE,
    MAINTENANCE
}
