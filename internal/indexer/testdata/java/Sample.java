package com.example;

import java.util.List;
import java.util.ArrayList;

public class Sample {
    private String name;
    private int count;

    public Sample(String name) {
        this.name = name;
        this.count = 0;
    }

    public String getName() {
        return name;
    }

    public void increment() {
        count++;
    }

    public int getCount() {
        return count;
    }
}

interface Processor {
    void process(String input);
}

enum Status {
    ACTIVE,
    INACTIVE,
    PENDING
}

class Worker implements Processor {
    private Sample sample;

    public Worker(Sample s) {
        this.sample = s;
    }

    public void process(String input) {
        sample.increment();
        String name = sample.getName();
        System.out.println(name + ": " + input);
    }

    public void run() {
        Sample local = new Sample("worker");
        local.increment();
        process(local.getName());
    }
}
