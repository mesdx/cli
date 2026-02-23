package com.example;

import models.CriticalUserModel;

/** DefsC contains additional utilities using CriticalUserModel. */
public class DefsC {
    /** Transform a CriticalUserModel for external use. */
    public static String transform(CriticalUserModel model) {
        return "[" + model.getName() + "]";
    }

    /** CriticalUserModelWrapper wraps a CriticalUserModel with extra metadata. */
    public static class CriticalUserModelWrapper {
        private CriticalUserModel inner;
        private String tag;

        public CriticalUserModelWrapper(CriticalUserModel inner, String tag) {
            this.inner = inner;
            this.tag = tag;
        }

        public CriticalUserModel unwrap() {
            return inner;
        }
    }
}
