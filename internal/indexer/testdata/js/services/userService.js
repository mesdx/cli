const MAX_WORKERS = 4;

function processUserData(userId, data) {
    return Object.keys(data).length > 0;
}

function validateEmail(email) {
    return email.includes("@");
}

const formatUserName = (first, last) => {
    return `${first.trim()} ${last.trim()}`;
};

class UserRepository {
    findById(userId) {
        return { id: String(userId) };
    }

    save(user) {
        return true;
    }
}

module.exports = { processUserData, validateEmail, formatUserName, UserRepository };
