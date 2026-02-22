package services;

import models.CoreModel;
import java.util.ArrayList;
import java.util.List;

/** CoreModelService manages CoreModel instances with high-coupling patterns.
 * Demonstrates: import (0.40), extends (1.0 via lexical), instantiation (0.95), method calls (0.65). */
public class CoreModelService extends CoreModel {

    private List<CoreModel> store = new ArrayList<>();

    /** Creates a CoreModelService seeded with an initial CoreModel. */
    public CoreModelService() {
        super("service-root", 1.0);
        CoreModel initial = new CoreModel("initial", 0.5);
        store.add(initial);
    }

    /** save stores a CoreModel entry (method call coupling). */
    public boolean save(CoreModel model) {
        if (model.isValid()) {
            store.add(model);
            return true;
        }
        return false;
    }

    /** findByTitle looks up a CoreModel by its title. */
    public CoreModel findByTitle(String title) {
        for (CoreModel m : store) {
            if (title.equals(m.getTitle())) {
                return m;
            }
        }
        return null;
    }

    /** findAll returns every stored CoreModel. */
    public List<CoreModel> findAll() {
        return store;
    }

    /** create instantiates a fresh CoreModel (instantiation coupling). */
    public static CoreModel create(String title, double score) {
        return new CoreModel(title, score);
    }
}
