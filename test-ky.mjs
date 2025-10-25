#!/usr/bin/env node

// Тестовый скрипт для проверки ky реализации
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Добавим путь к модулям
process.chdir(join(__dirname, 'src', 'js'));

try {
    console.log('🧪 Testing ky implementation...');
    console.log('Working directory:', process.cwd());

    // Импортируем CLI
    const { main } = await import('./lib/index.mjs');

    console.log('✅ Successfully imported CLI');

    // Попробуем выполнить check команду на примере файла
    process.argv = [
        'node',
        'testifai.mjs',
        'check',
        './samples/test-example.js',
        '--verbose'
    ];

    console.log('🔍 Running check command on test-example.js...');

    await main();

    console.log('✅ Check command completed successfully');

} catch (error) {
    console.error('❌ Test failed:', error.message);
    console.error('Stack:', error.stack);
    process.exit(1);
}