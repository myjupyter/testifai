import { describe, it, expect, afterEach, vi } from 'vitest';
import { sendToAPI } from '../api.mjs';
import { TESTIFAI_URL } from '../../constants.mjs';

/** @typedef {import('../../types.mjs').TestifaiConfig} TestifaiConfig */
/** @typedef {import('../../types.mjs').ScanResult} ScanResult */
/** @typedef {import('../../types.mjs').APIStatus} APIStatus /*

/**
 * Test API connection
 * @returns {Promise<boolean>} Whether API is reachable
 */
async function testAPIConnection() {
    try {
        // Try a simple request to see if the server is running
        await fetch(TESTIFAI_URL, {
            method: 'GET',
            headers: {
                Accept: 'application/json',
            },
        });

        return true;
    } catch (error) {
        return false;
    }
}

/**
 * Get API status and information
 * @returns {Promise<APIStatus>} API status information
 */
async function getAPIStatus() {
    try {
        const isReachable = await testAPIConnection();

        if (!isReachable) {
            return {
                status: 'unreachable',
                endpoint: TESTIFAI_URL,
                message: 'Cannot connect to API endpoint',
            };
        }

        return {
            status: 'reachable',
            endpoint: TESTIFAI_URL,
            message: 'API endpoint is reachable',
        };
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        return {
            status: 'error',
            endpoint: TESTIFAI_URL,
            message: err.message,
            error: err,
        };
    }
}

const baseConfig = /** @type {TestifaiConfig} */ ({
    provider: { openai: { apiKey: 'valid-api-key-12345' } },
    output: {
        testDir: null,
        extension: '.spec',
    },
    scan: {
        include: ['**/*.ts'],
        exclude: [],
    },
});

const scanResult = /** @type {ScanResult} */ ({
    file: '/tmp/source.ts',
    line: 1,
    comment: '// ts:generate testifai: -type=table',
    testType: 'table',
    additionalParams: '',
    functionName: 'sample',
    functionCode: 'export function sample() { return true; }',
    functionStartLine: 1,
    language: 'ts',
});

describe('api core utilities', () => {
    const originalFetch = global.fetch;

    afterEach(() => {
        vi.restoreAllMocks();
        global.fetch = originalFetch;
    });

    it('sendToAPI posts payload and returns parsed response', async () => {
        const fetchMock = vi.fn(async (url, options) => {
            expect(url).toBe(`${TESTIFAI_URL}/generate`);
            expect(options.method).toBe('POST');

            const body = JSON.parse(options.body);

            expect(body).toMatchObject({
                provider: 'OpenAI',
                context: { user_code: scanResult.functionCode },
            });

            return new Response(
                JSON.stringify({
                    id: 'req-1',
                    generated: { test_code: 'describe("sample", () => {});' },
                }),
                { status: 200, headers: { 'Content-Type': 'application/json' } },
            );
        });

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const response = await sendToAPI({ ...baseConfig }, scanResult, {
            timeout: 1000,
            verbose: true,
        });

        expect(response.generated.test_code).toContain('describe("sample"');
        expect(fetchMock).toHaveBeenCalledOnce();
    });

    it('sendToAPI throws with contextual error on failure', async () => {
        const fetchMock = vi.fn(async () => {
            return new Response('boom', {
                status: 500,
                statusText: 'Internal Server Error',
                headers: { 'Content-Type': 'text/plain' },
            });
        });

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        await expect(sendToAPI({ ...baseConfig }, scanResult)).rejects.toThrow(/Failed to generate tests for sample/);
    });

    it('testAPIConnection returns true when endpoint is reachable', async () => {
        const fetchMock = vi.fn(async () => new Response(null, { status: 200 }));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const reachable = await testAPIConnection();

        expect(reachable).toBe(true);
        expect(fetchMock).toHaveBeenCalledWith(TESTIFAI_URL, expect.anything());
    });

    it('testAPIConnection returns false when fetch throws', async () => {
        const fetchMock = vi.fn(async () => {
            throw new Error('connection refused');
        });

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const reachable = await testAPIConnection();

        expect(reachable).toBe(false);
    });

    it('getAPIStatus maps fetch results to status objects', async () => {
        const fetchMock = vi
            .fn()
            .mockResolvedValueOnce(new Response(null, { status: 200 }))
            .mockRejectedValueOnce(new Error('unreachable'));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const success = await getAPIStatus();

        expect(success.status).toBe('reachable');

        const failure = await getAPIStatus();

        expect(failure.status).toBe('unreachable');
        expect(failure.message).toMatch(/Cannot connect/);
    });
});
