// @ts-check

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { scanFiles, parseCommentParams, validateScanResult } from '../parser.mjs';

/** @typedef {import('../../types.mjs').ScanResult} ScanResult */

/**
 * @param {string} dir
 * @param {string} filename
 * @param {string} contents
 * @returns {string}
 */
function createFixture(dir, filename, contents) {
    const path = join(dir, filename);

    writeFileSync(path, contents, 'utf8');

    return path;
}

describe('parser core utilities', () => {
    /** @type {string} */
    let tempDir;

    beforeEach(() => {
        tempDir = mkdtempSync(join(tmpdir(), 'testifai-parser-'));
    });

    afterEach(() => {
        rmSync(tempDir, { recursive: true, force: true });
    });

    it('parseCommentParams handles flags and key/value pairs', () => {
        const params = parseCommentParams('-type=table -skip -timeout=5000 -parallel');

        expect(params.type).toBe('table');
        expect(params.skip).toBe(true);
        expect(params.timeout).toBe(5000);
        expect(params.parallel).toBe(true);
    });

    it('scanFiles collects functions following testifai comments', async () => {
        const source = `
// testifai: -type=suite
export function add(a, b) {
  return a + b;
}
`;

        const filePath = createFixture(tempDir, 'math.ts', source);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(1);
        const [result] = results;

        expect(result.file).toBe(filePath);
        expect(result.functionName).toBe('add');
        expect(result.functionCode).toContain('return a + b;');
        expect(result.language).toBe('ts');
        expect(validateScanResult(/** @type {ScanResult} */ (result))).toBe(true);
    });

    it('scanFiles returns empty array when no comments are present', async () => {
        createFixture(tempDir, 'noop.ts', 'export const noop = () => 1;');

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(0);
    });
});
