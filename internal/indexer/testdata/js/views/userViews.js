const { processUserData, validateEmail, formatUserName } = require("../services/userService");

function handleUpdate(userId, data) {
    const result = processUserData(userId, data);
    return result;
}

function handleValidate(email) {
    return validateEmail(email);
}

function handleFormat(first, last) {
    return formatUserName(first, last);
}

module.exports = { handleUpdate, handleValidate, handleFormat };
