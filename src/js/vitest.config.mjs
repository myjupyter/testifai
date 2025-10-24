import { defineConfig } from 'vitest/config';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const rootDir = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
    root: rootDir,
    test: {
        environment: 'node',
        include: ['**/__tests__/**/*.{test,spec}.{js,mjs,ts}', '**/*.{test,spec}.{js,mjs,ts}'],
    },
});
