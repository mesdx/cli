package com.example;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.io.Closeable;
import java.io.IOException;

// --- Inheritance chain ---

abstract class BaseEntity {
    private long id;

    public long getId() {
        return id;
    }

    public abstract void validate();
}

class User extends BaseEntity implements Closeable {
    private String username;
    private String email;

    public User(String username, String email) {
        this.username = username;
        this.email = email;
    }

    @Override
    public void validate() {
        if (username == null || username.isEmpty()) {
            throw new IllegalArgumentException("username is required");
        }
    }

    @Override
    public void close() throws IOException {
        // cleanup
    }

    // --- Method overloading (same name, different params) ---
    public String format() {
        return username + " <" + email + ">";
    }

    public String format(boolean includeEmail) {
        if (includeEmail) {
            return format();
        }
        return username;
    }

    // --- Nested/inner class ---
    static class Permissions {
        private List<String> roles;

        public Permissions(List<String> roles) {
            this.roles = roles;
        }

        public boolean hasRole(String role) {
            return roles.contains(role);
        }
    }
}

// --- Annotation usage ---

@Deprecated
@SuppressWarnings("unchecked")
class LegacyProcessor {
    @SuppressWarnings("unused")
    private int count;

    public void process() {
        // no-op
    }
}

// --- Generics ---

class Repository<T extends BaseEntity> {
    private Map<Long, T> store = new HashMap<>();

    public void save(T entity) {
        store.put(entity.getId(), entity);
    }

    public Optional<T> findById(long id) {
        return Optional.ofNullable(store.get(id));
    }
}

// --- Static fields and methods ---

class AppConstants {
    public static final String APP_NAME = "MyApp";
    public static final int MAX_CONNECTIONS = 100;

    public static String getAppInfo() {
        return APP_NAME + " v1.0";
    }
}

// --- Enum with methods ---

enum Priority {
    LOW(1),
    MEDIUM(5),
    HIGH(10);

    private final int value;

    Priority(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}

// --- Builtins / java.lang usage (should not be unresolved-external) ---

class BuiltinUsage {
    public void demo() {
        // java.lang builtins
        String s = "hello";
        Integer i = Integer.valueOf(42);
        Double d = Double.parseDouble("3.14");
        Boolean b = Boolean.TRUE;
        Object obj = new Object();
        System.out.println(s + i + d + b + obj);

        // Exception hierarchy
        try {
            throw new RuntimeException("test");
        } catch (Exception e) {
            e.printStackTrace();
        }

        // Math
        double abs = Math.abs(-1.5);
        System.out.println(abs);
    }
}
