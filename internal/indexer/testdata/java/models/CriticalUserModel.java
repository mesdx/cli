package models;

/** CriticalUserModel is a heavily-used user representation that many files depend on. */
public class CriticalUserModel extends BaseModel {
    private String name;
    private String email;

    public CriticalUserModel(String name, String email) {
        super(0);
        this.name = name;
        this.email = email;
    }

    public String getName() {
        return name;
    }

    public String getEmail() {
        return email;
    }

    public boolean validate() {
        return name != null && !name.isEmpty() && email != null && email.contains("@");
    }
}
