package models;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.ArrayList;
import java.util.HashMap;

/**
 * TypeUsageEdges — fixture testing that CoreModel is found as a ref when it
 * appears inside generic/container type expressions.
 *
 * Every method uses CoreModel exclusively in type positions so the fixture
 * exercises Java's type-identifier capture ({@code (type_identifier) @ref.type}).
 */
public class TypeUsageEdges {

    /** Receives a List of CoreModel. */
    public static int takesList(List<CoreModel> models) {
        return models.size();
    }

    /** Returns an Optional wrapping a CoreModel. */
    public static Optional<CoreModel> returnsOptional(String title) {
        if (title == null || title.isEmpty()) {
            return Optional.empty();
        }
        return Optional.of(new CoreModel(title, 1.0));
    }

    /** Receives a Map from String to CoreModel. */
    public static CoreModel getFromMap(Map<String, CoreModel> map, String key) {
        return map.get(key);
    }

    /** Receives a CoreModel array. */
    public static int takesArray(CoreModel[] models) {
        return models.length;
    }

    /** Returns a new CoreModel array. */
    public static CoreModel[] returnsArray(int size) {
        return new CoreModel[size];
    }

    /** Receives a Map with List values that contain CoreModel. */
    public static int deepNested(Map<String, List<CoreModel>> data) {
        int total = 0;
        for (List<CoreModel> list : data.values()) {
            total += list.size();
        }
        return total;
    }

    /** Local variable declaration of type CoreModel. */
    public static boolean localVar(String title) {
        CoreModel m = new CoreModel(title, 0.5);
        return m.isValid();
    }

    /** Builds a List<CoreModel> and returns it. */
    public static List<CoreModel> buildList(int n) {
        List<CoreModel> list = new ArrayList<>();
        for (int i = 0; i < n; i++) {
            list.add(new CoreModel("item" + i, i));
        }
        return list;
    }

    /** Builds a Map<String, CoreModel>. */
    public static Map<String, CoreModel> buildMap(String[] titles) {
        Map<String, CoreModel> map = new HashMap<>();
        for (String t : titles) {
            map.put(t, new CoreModel(t, 1.0));
        }
        return map;
    }
}
