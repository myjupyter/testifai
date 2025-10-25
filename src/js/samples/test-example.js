// Тестовый файл для проверки testifai CLI

// js:generate testifai: -type=xunit
function calculateSum(a, b) {
    return a + b;
}

// js:generate testifai: -type=table
function validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    return emailRegex.test(email);
}

export { calculateSum, validateEmail };
