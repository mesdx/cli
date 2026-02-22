package models;

/** CoreModel is a high-usage domain model referenced by many files.
 * Used as a fixture to verify coupling score distribution for popular symbols. */
public class CoreModel extends BaseModel {
    private String title;
    private double score;

    public CoreModel(String title, double score) {
        super(0);
        this.title = title;
        this.score = score;
    }

    public String getTitle() {
        return title;
    }

    public double getScore() {
        return score;
    }

    public boolean isValid() {
        return title != null && !title.isEmpty() && score >= 0;
    }

    public String describe() {
        return title;
    }
}
