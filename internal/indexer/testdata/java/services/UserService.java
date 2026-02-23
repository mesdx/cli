package services;

import java.util.Map;
import java.util.Optional;

public class UserService {
    public static final int MAX_WORKERS = 4;

    public static boolean processUserData(int userId, Map<String, String> data) {
        return data != null && !data.isEmpty();
    }

    public static boolean validateEmail(String email) {
        return email != null && email.contains("@");
    }

    public static String formatUserName(String first, String last) {
        return first.trim() + " " + last.trim();
    }

    public Optional<Map<String, String>> findById(int userId) {
        return Optional.of(Map.of("id", String.valueOf(userId)));
    }

    public boolean save(Map<String, String> user) {
        return true;
    }
}
