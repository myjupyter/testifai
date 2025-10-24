// Тестовый файл для проверки testifai CLI

// testifai: -type=xunit
function calculateSum(a: number, b: number): number {
    return a + b;
}

// testifai: -type=table
function validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

export { calculateSum, validateEmail };
