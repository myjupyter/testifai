#!/usr/bin/env node

// Простой тест для проверки ky реализации
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Добавим путь к модулям js
process.chdir(join(__dirname, 'src', 'js'));

try {
    console.log('🧪 Testing ky implementation...');
    console.log('Current directory:', process.cwd());

    // Импортируем функции напрямую
    const { sendToAPI } = await import('./lib/core/api.mjs');
    console.log('✅ Successfully imported sendToAPI');

    // Создаем тестовую конфигурацию
    const testConfig = {
        provider: { openai: { apiKey: 'test-key' } },
        output: { testDir: null, extension: '.spec' },
        scan: { include: ['**/*.ts'], exclude: [] },
    };

    // Создаем тестовый результат сканирования
    const testScanResult = {
        file: '/tmp/test.js',
        line: 1,
        comment: '// js:generate testifai: -type=xunit',
        testType: 'xunit',
        additionalParams: '',
        functionName: 'calculateSum',
        functionCode: 'function calculateSum(a, b) { return a + b; }',
        functionStartLine: 1,
        platform: 'js',
    };

    console.log('🔍 Testing API call with ky...');

    // Попробуем отправить запрос (ожидаем ошибку соединения, но не bad port)
    try {
        const response = await sendToAPI(testConfig, testScanResult, {
            verbose: true,
            timeout: 5000
        });
        console.log('✅ API call successful:', response);
    } catch (error) {
        console.log('📊 API call result:');
        console.log('  Error name:', error.name);
        console.log('  Error message:', error.message);

        // Проверяем что это НЕ "bad port" ошибка
        if (error.message.includes('bad port')) {
            console.log('❌ Still getting "bad port" error - ky not working');
            throw error;
        } else if (error.message.includes('ECONNREFUSED') || error.message.includes('connect')) {
            console.log('✅ ky working! Got connection error (expected when server not running)');
        } else {
            console.log('🤔 Got different error:', error.message);
        }
    }

    console.log('✅ ky test completed successfully');

} catch (error) {
    console.error('❌ Test failed:', error.message);
    if (error.stack) {
        console.error('Stack:', error.stack);
    }
    process.exit(1);
}