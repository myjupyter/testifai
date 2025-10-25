import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtempSync, rmSync, readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { tmpdir } from 'node:os';
import { generateTestFiles, generateMultipleTestFiles, getTestFileStats, validateTestFile } from '../generator.mjs';

/** @typedef {import('../../types.mjs').ScanResult} ScanResult */
/** @typedef {import('../../types.mjs').APIResponse} APIResponse */
/** @typedef {import('../../types.mjs').TestifaiConfig} TestifaiConfig */

const baseConfig = /** @type {TestifaiConfig} */ ({
    provider: { openai: { apiKey: 'valid-api-key-12345' } },
});

const scanResult = /** @type {ScanResult} */ ({
    file: '',
    functionName: 'add',
    functionCode: 'export function add(a, b) { return a + b; }',
    functionStartLine: 1,
    testType: 'table',
    language: 'ts',
    comment: '// ts:generate testifai: -type=table',
    line: 1,
    additionalParams: '',
});

const apiResponse = /** @type {APIResponse} */ ({
    id: 'response-1',
    generated: { test_code: "describe('add', () => { it('adds numbers', () => { expect(add(1,2)).toBe(3); }); });" },
});

describe('generator core utilities', () => {
    const originalCwd = process.cwd();
    /** @type {string} */
    let tempDir;
    /** @type {string} */
    let sourceFile;

    beforeEach(() => {
        tempDir = mkdtempSync(join(tmpdir(), 'testifai-generator-'));
        sourceFile = join(tempDir, 'src', 'math.ts');

        process.chdir(tempDir);
        mkdirSync(dirname(sourceFile), { recursive: true });
        writeFileSync(sourceFile, 'export const noop = () => {};', { encoding: 'utf8' });
    });

    afterEach(() => {
        process.chdir(originalCwd);
        rmSync(tempDir, { recursive: true, force: true });
    });

    it('generateTestFiles creates tests next to source file by default', async () => {
        const result = { ...scanResult, file: sourceFile };

        const testPath = await generateTestFiles(result, apiResponse, { config: baseConfig, verbose: true });

        expect(existsSync(testPath)).toBe(true);

        const content = readFileSync(testPath, 'utf8');

        expect(content).toMatch(/Generated tests for add/);
        expect(content).toMatch(/describe\('add'/);
    });

    it('generateTestFiles respects overwrite flag', async () => {
        const result = { ...scanResult, file: sourceFile };
        const testPath = await generateTestFiles(result, apiResponse, { config: baseConfig });

        writeFileSync(testPath, '// existing test', 'utf8');

        const secondPath = await generateTestFiles(result, apiResponse, { config: baseConfig });

        expect(readFileSync(secondPath, 'utf8')).toBe('// existing test');

        const overwrittenPath = await generateTestFiles(result, apiResponse, { config: baseConfig, overwrite: true });

        expect(readFileSync(overwrittenPath, 'utf8')).toMatch(/Generated tests for add/);
    });

    it('generateMultipleTestFiles processes array sequentially', async () => {
        const resultA = { ...scanResult, file: sourceFile };
        const otherFile = join(tempDir, 'src', 'other.ts');

        mkdirSync(dirname(otherFile), { recursive: true });
        writeFileSync(otherFile, 'export const mul = () => {};', 'utf8');

        const resultB = { ...scanResult, file: otherFile, functionName: 'mul' };

        const responses = [
            apiResponse,
            { ...apiResponse, id: 'response-2', generated: { test_code: "describe('mul', () => {});" } },
        ];

        const files = await generateMultipleTestFiles([resultA, resultB], responses, { config: baseConfig });

        expect(files).toHaveLength(2);
        expect(existsSync(files[0])).toBe(true);
        expect(existsSync(files[1])).toBe(true);
    });

    it('getTestFileStats returns file metadata', async () => {
        const result = { ...scanResult, file: sourceFile };
        const testPath = await generateTestFiles(result, apiResponse, { config: baseConfig });

        const stats = getTestFileStats(testPath);

        expect(stats && stats.path).toBe(testPath);
        expect(stats && stats.totalLines).toBeGreaterThan(1);
        expect(stats && stats.codeLines).toBeGreaterThan(1);
    });

    it('validateTestFile flags missing files and validates existing ones', async () => {
        const missing = validateTestFile(join(tempDir, 'missing.spec.ts'));

        expect(missing.valid).toBe(false);
        expect(missing.errors).toContain('Test file does not exist');

        const result = { ...scanResult, file: sourceFile };
        const testPath = await generateTestFiles(result, apiResponse, { config: baseConfig });

        const validation = validateTestFile(testPath);

        expect(validation.valid).toBe(true);
        expect(validation.errors).toHaveLength(0);
    });
});
