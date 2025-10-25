// Test file for TypeScript with new comment format

// ts:generate testifai: -type=table
function processUser(user) {
    return `User ${user.name} has ID ${user.id}`;
}

// This should be ignored (wrong directive)
// js:generate testifai: -type=xunit
function shouldBeIgnored() {
    return false;
}

// ts:generate testifai: -type=table -parallel=true
function validateUserData(users) {
    return users.every((user) => user.id > 0 && user.name.length > 0);
}

// This should be ignored (wrong test type for ts)
// ts:generate testifai: -type=xunit
function anotherIgnored() {
    return 42;
}

export { processUser, validateUserData };
