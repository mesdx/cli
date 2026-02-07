package com.example;

/**
 * AppProcessor handles data processing.
 */
public class AppProcessor {
    private String name;

    public AppProcessor(String name) {
        this.name = name;
    }

    // Run the processor
    public void run() {
        System.out.println("Running " + name);
    }

    // Stop the processor
    public void stop() {
        System.out.println("Stopping " + name);
    }
}
