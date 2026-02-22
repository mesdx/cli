package models;

/** BaseModel provides base entity identity fields. */
public class BaseModel {
    private int id;
    private String createdAt;

    public BaseModel() {}

    public BaseModel(int id) {
        this.id = id;
    }

    public int getId() {
        return id;
    }

    public String getCreatedAt() {
        return createdAt;
    }
}
