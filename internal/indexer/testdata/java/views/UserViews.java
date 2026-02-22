package views;

import static services.UserService.processUserData;
import static services.UserService.validateEmail;
import services.UserService;

import java.util.Map;

public class UserViews {
    public boolean handleUpdate(int userId, Map<String, String> data) {
        boolean result = processUserData(userId, data);
        return result;
    }

    public boolean handleValidate(String email) {
        return validateEmail(email);
    }

    public String handleFormat(String first, String last) {
        return UserService.formatUserName(first, last);
    }
}
