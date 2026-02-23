package views;

import models.CriticalUserModel;

/** CriticalUserModelView renders a CriticalUserModel for display. */
public class CriticalUserModelView {
    private CriticalUserModel model;

    public CriticalUserModelView(CriticalUserModel model) {
        this.model = model;
    }

    public String render() {
        return model.getName() + " <" + model.getEmail() + ">";
    }

    public static CriticalUserModel createDefault(String name, String email) {
        return new CriticalUserModel(name, email);
    }
}
