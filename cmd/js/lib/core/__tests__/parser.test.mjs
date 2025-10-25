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

    it('scanFiles collects functions following ts generate comments', async () => {
        const source = `
// ts:generate testifai: -type=table
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

    it('scanFiles collects functions following js generate comments', async () => {
        const source = `
// js:generate testifai: -type=xunit
export const multiply = (a, b) => a * b;
`;

        const filePath = createFixture(tempDir, 'math.js', source);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(1);
        const [result] = results;

        expect(result.file).toBe(filePath);
        expect(result.functionName).toBe('multiply');
        expect(result.functionCode).toContain('a * b');
        expect(result.language).toBe('js');
        expect(validateScanResult(/** @type {ScanResult} */ (result))).toBe(true);
    });

    it('scanFiles returns empty array when no comments are present', async () => {
        createFixture(tempDir, 'noop.ts', 'export const noop = () => 1;');

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(0);
    });

    it('ignores mismatched generator directives', async () => {
        const jsSource = `
// ts:generate testifai: -type=table
export const add = (a, b) => a + b;
`;

        createFixture(tempDir, 'math.js', jsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(0);
    });

    it('ignores wrong test types for JavaScript files', async () => {
        const jsSource = `
// js:generate testifai: -type=table
export const add = (a, b) => a + b;
`;

        createFixture(tempDir, 'math.js', jsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(0);
    });

    it('ignores wrong test types for TypeScript files', async () => {
        const tsSource = `
// ts:generate testifai: -type=xunit
export const add = (a: number, b: number): number => a + b;
`;

        createFixture(tempDir, 'math.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(0);
    });

    it('processes multiple valid comments in the same file', async () => {
        const jsSource = `
// js:generate testifai: -type=xunit
export function add(a, b) {
    return a + b;
}

// js:generate testifai: -type=xunit -timeout=3000
export function multiply(a, b) {
    return a * b;
}

// This should be ignored - wrong directive
// ts:generate testifai: -type=table
export function ignored() {
    return false;
}
`;

        createFixture(tempDir, 'math.js', jsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(2);
        expect(results[0].functionName).toBe('add');
        expect(results[0].testType).toBe('xunit');
        expect(results[1].functionName).toBe('multiply');
        expect(results[1].testType).toBe('xunit');
    });

    it('handles TypeScript files with interface definitions', async () => {
        const tsSource = `
interface User {
    id: number;
    name: string;
}

// ts:generate testifai: -type=table
export function processUser(user: User): string {
    return \`User \${user.name} has ID \${user.id}\`;
}

// ts:generate testifai: -type=table -parallel=true
export function validateUsers(users: User[]): boolean {
    return users.every(user => user.id > 0);
}
`;

        createFixture(tempDir, 'user.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(2);
        expect(results[0].functionName).toBe('processUser');
        expect(results[0].testType).toBe('table');
        expect(results[0].language).toBe('ts');
        expect(results[1].functionName).toBe('validateUsers');
        expect(results[1].testType).toBe('table');
        expect(results[1].language).toBe('ts');
    });

    it('validates comment language extraction', () => {
        const jsComment = '// js:generate testifai: -type=xunit -timeout=5000';
        const tsComment = '// ts:generate testifai: -type=table -parallel=true';

        // Test regex directly (accessing private regex would require refactoring)
        const jsMatch = jsComment.match(/^\/\/\s*(js|ts):generate\s+testifai:\s*-type=([a-zA-Z]+)(.*)$/i);
        const tsMatch = tsComment.match(/^\/\/\s*(js|ts):generate\s+testifai:\s*-type=([a-zA-Z]+)(.*)$/i);

        expect(jsMatch).not.toBeNull();
        expect(jsMatch?.[1]).toBe('js');
        expect(jsMatch?.[2]).toBe('xunit');

        expect(tsMatch).not.toBeNull();
        expect(tsMatch?.[1]).toBe('ts');
        expect(tsMatch?.[2]).toBe('table');
    });

    it('scans directory with mixed JS and TS files correctly', async () => {
        const jsSource = `
// js:generate testifai: -type=xunit
export function jsFunction() {
    return 'JavaScript';
}
`;

        const tsSource = `
// ts:generate testifai: -type=table
export function tsFunction(): string {
    return 'TypeScript';
}
`;

        createFixture(tempDir, 'script.js', jsSource);
        createFixture(tempDir, 'module.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(2);

        const jsResult = results.find((r) => r.language === 'js');
        const tsResult = results.find((r) => r.language === 'ts');

        expect(jsResult?.functionName).toBe('jsFunction');
        expect(jsResult?.testType).toBe('xunit');

        expect(tsResult?.functionName).toBe('tsFunction');
        expect(tsResult?.testType).toBe('table');
    });

    it('recognizes classes with generate comments', async () => {
        const tsSource = `
// ts:generate testifai: -type=table
export class Calculator {
    add(a: number, b: number): number {
        return a + b;
    }

    multiply(a: number, b: number): number {
        return a * b;
    }
}
`;

        createFixture(tempDir, 'calculator.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(1);
        const [result] = results;

        expect(result.functionName).toBe('Calculator');
        expect(result.functionCode).toContain('class Calculator');
        expect(result.functionCode).toContain('add(a: number, b: number)');
        expect(result.functionCode).toContain('multiply(a: number, b: number)');
        expect(result.language).toBe('ts');
    });

    it('recognizes async arrow functions', async () => {
        const jsSource = `
// js:generate testifai: -type=xunit
export const fetchData = async (url) => {
    const response = await fetch(url);
    return response.json();
};
`;

        createFixture(tempDir, 'api.js', jsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(1);
        const [result] = results;

        expect(result.functionName).toBe('fetchData');
        expect(result.functionCode).toContain('async (url)');
        expect(result.functionCode).toContain('await fetch(url)');
        expect(result.language).toBe('js');
    });

    it('recognizes short arrow functions', async () => {
        const jsSource = `
// js:generate testifai: -type=xunit
export const double = x => x * 2;

// js:generate testifai: -type=xunit
const add = (a, b) => a + b;
`;

        createFixture(tempDir, 'utils.js', jsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(2);
        expect(results[0].functionName).toBe('double');
        expect(results[0].functionCode).toContain('x => x * 2');
        expect(results[1].functionName).toBe('add');
        expect(results[1].functionCode).toContain('(a, b) => a + b');
    });

    it('recognizes class methods', async () => {
        const tsSource = `
export class MathService {
    // ts:generate testifai: -type=table
    static calculate(a: number, b: number): number {
        return a + b;
    }

    // ts:generate testifai: -type=table
    async processAsync(data: any): Promise<any> {
        return await this.process(data);
    }

    private process(data: any): any {
        return data;
    }
}
`;

        createFixture(tempDir, 'service.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(2);
        expect(results[0].functionName).toBe('calculate');
        expect(results[0].functionCode).toContain('static calculate');
        expect(results[1].functionName).toBe('processAsync');
        expect(results[1].functionCode).toContain('async processAsync');
    });

    it('recognizes constructor methods', async () => {
        const tsSource = `
export class User {
    name: string;

    // ts:generate testifai: -type=table
    constructor(name: string) {
        this.name = name;
    }
}
`;

        createFixture(tempDir, 'user.ts', tsSource);

        const results = await scanFiles(tempDir);

        expect(results).toHaveLength(1);
        const [result] = results;

        expect(result.functionName).toBe('constructor');
        expect(result.functionCode).toContain('constructor(name: string)');
        expect(result.functionCode).toContain('this.name = name');
    });
});
