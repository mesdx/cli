/**
 * type_usage_edges.js — fixture testing that CoreModel is found as a usage ref
 * inside various JavaScript runtime patterns.
 *
 * JavaScript has no static type system, so this fixture focuses on runtime
 * usage patterns where CoreModel appears: instanceof checks, destructuring,
 * array.filter / array.map callbacks, and JSDoc @type annotations.
 */

const { CoreModel } = require("./models/coreModel");

/** Filter an array to keep only CoreModel instances. */
function filterModels(items) {
    return items.filter(item => item instanceof CoreModel);
}

/** Map over models and call CoreModel.describe(). */
function describeAll(models) {
    return models.map(m => {
        if (m instanceof CoreModel) {
            return m.describe();
        }
        return null;
    });
}

/** Factory: returns a new CoreModel. */
function createModel(title, score) {
    return new CoreModel(title, score);
}

/** Guard: checks whether value is a CoreModel. */
function isModel(value) {
    return value instanceof CoreModel;
}

/** Accumulate CoreModel instances into a result array. */
function collectModels(items) {
    const result = [];
    for (const item of items) {
        if (item instanceof CoreModel) {
            result.push(item);
        }
    }
    return result;
}

/** Destructure from an object that has a model property. */
function extractModel({ model }) {
    if (model instanceof CoreModel) {
        return model.isValid();
    }
    return false;
}

module.exports = {
    filterModels,
    describeAll,
    createModel,
    isModel,
    collectModels,
    extractModel,
};
