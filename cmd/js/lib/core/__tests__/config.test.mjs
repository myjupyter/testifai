import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { mkdtempSync, rmSync, existsSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { initConfig, loadConfig, validateConfig, getEffectiveConfig } from '../config.mjs';

/** @typedef {import('../../types.mjs').TestifaiConfig} TestifaiConfig */
/** @typedef {{ cwd: string, dir: string }} CleanupContext */

const validConfig = /** @type {TestifaiConfig} */ ({
    provider: {
        openai: {
            apiKey: 'test-api-key-12345',
        },
    },
});

describe('config core utilities', () => {
    const originalCwd = process.cwd();
    /** @type {CleanupContext} */
    let context;

    beforeEach(() => {
        const dir = mkdtempSync(join(tmpdir(), 'testifai-config-'));

        process.chdir(dir);
        context = { cwd: originalCwd, dir };
        vi.spyOn(console, 'log').mockImplementation(() => {});
    });

    afterEach(() => {
        process.chdir(context.cwd);
        rmSync(context.dir, { recursive: true, force: true });
        vi.restoreAllMocks();
    });

    it('initConfig creates default config and loadConfig reads it back', async () => {
        await initConfig();

        const configPath = join(process.cwd(), 'testifai.json');

        expect(existsSync(configPath)).toBe(true);

        const loaded = await loadConfig(undefined);

        expect(loaded.provider.openai.apiKey).toBe('<YOUR_OPENAI_API_KEY>');
        expect(loaded.output.extension).toBe('.spec');
    });

    it('initConfig respects force flag', async () => {
        await initConfig(undefined);
        const configPath = join(process.cwd(), 'testifai.json');

        const firstContent = readFileSync(configPath, 'utf8');

        await initConfig();
        expect(readFileSync(configPath, 'utf8')).toBe(firstContent);

        writeFileSync(configPath, '{}', { encoding: 'utf8' });

        await initConfig({ force: true });
        expect(readFileSync(configPath, 'utf8')).toBe(firstContent);
    });

    it('validateConfig rejects missing or placeholder values', () => {
        expect(() => validateConfig(/** @type {any} */ (undefined))).toThrow(/Configuration is required/);
        expect(() => validateConfig(/** @type {TestifaiConfig} */ ({}))).toThrow(/OpenAI API key is required/);

        const placeholderConfig = /** @type {TestifaiConfig} */ ({
            provider: {
                openai: {
                    apiKey: '<YOUR_OPENAI_API_KEY>',
                },
            },
        });

        expect(() => validateConfig(placeholderConfig)).toThrow(/Please set a valid OpenAI API key/);
    });

    it('getEffectiveConfig merges defaults with overrides', () => {
        const effective = getEffectiveConfig({
            ...validConfig,
            output: {
                testDir: 'tests',
                extension: '.spec',
            },
        });

        expect(effective.provider.openai.apiKey).toBe('test-api-key-12345');
        expect(effective.output.testDir).toBe('tests');
        expect(effective.output.extension).toBe('.spec');
        expect(effective.scan.include).toContain('**/*.ts');
    });
});
