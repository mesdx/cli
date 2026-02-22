package services;

import models.CriticalUserModel;
import java.util.List;
import java.util.ArrayList;

/** CriticalUserModelService manages CriticalUserModel persistence and retrieval. */
public class CriticalUserModelService {
    private List<CriticalUserModel> store = new ArrayList<>();

    public boolean save(CriticalUserModel model) {
        return store.add(model);
    }

    public CriticalUserModel findByEmail(String email) {
        for (CriticalUserModel m : store) {
            if (email.equals(m.getEmail())) {
                return m;
            }
        }
        return null;
    }

    public List<CriticalUserModel> findAll() {
        return store;
    }
}
