// @ts-check

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { sendToAPI, testAPIConnection, getAPIStatus } from '../api.mjs';

/** @typedef {import('../../types.mjs').TestifaiConfig} TestifaiConfig */
/** @typedef {import('../../types.mjs').ScanResult} ScanResult */
/** @typedef {import('../../types.mjs').APIResponse} APIResponse */

const minimalConfig = /** @type {TestifaiConfig} */ ({
    provider: { openai: { apiKey: 'valid-api-key-12345' } },
    api: { endpoint: 'http://localhost:7777' },
});

const scanResult = /** @type {ScanResult} */ ({
    file: '/tmp/source.ts',
    line: 1,
    comment: '// testifai: -type=suite',
    testType: 'suite',
    additionalParams: '',
    functionName: 'sample',
    functionCode: 'export function sample() { return true; }',
    functionStartLine: 1,
    language: 'ts',
});

describe('api core utilities', () => {
    const originalFetch = global.fetch;

    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.clearAllTimers();
        vi.restoreAllMocks();
        global.fetch = originalFetch;
    });

    it('sendToAPI posts payload and returns parsed response', async () => {
        const mockResponse = /** @type {APIResponse} */ ({
            id: 'req-1',
            generated: { test_code: 'describe("sample", () => {});' },
        });

        const fetchMock = vi.fn(async () => ({
            ok: true,
            status: 200,
            statusText: 'OK',
            json: async () => mockResponse,
            text: async () => JSON.stringify(mockResponse),
        }));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const result = await sendToAPI(minimalConfig, scanResult, { timeout: 1000, verbose: true });

        expect(result).toEqual(mockResponse);
        expect(fetchMock).toHaveBeenCalledTimes(1);
        expect(fetchMock.mock.calls[0][0]).toBe(`${minimalConfig.api.endpoint}/generate`);
    });

    it('sendToAPI throws with contextual error on failure', async () => {
        const fetchMock = vi.fn(async () => ({
            ok: false,
            status: 500,
            statusText: 'Internal Server Error',
            text: async () => 'boom',
        }));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        await expect(sendToAPI(minimalConfig, scanResult)).rejects.toThrow(/Failed to generate tests for sample/);
    });

    it('testAPIConnection returns true when endpoint is reachable', async () => {
        const fetchMock = vi.fn(async () => ({
            ok: true,
            status: 200,
            statusText: 'OK',
        }));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const reachable = await testAPIConnection(minimalConfig);

        expect(reachable).toBe(true);
        expect(fetchMock).toHaveBeenCalledWith(minimalConfig.api.endpoint, expect.anything());
    });

    it('testAPIConnection returns false when fetch throws', async () => {
        const fetchMock = vi.fn(async () => {
            throw new Error('connection refused');
        });

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const reachable = await testAPIConnection(minimalConfig);

        expect(reachable).toBe(false);
    });

    it('getAPIStatus maps fetch results to status objects', async () => {
        const fetchMock = vi.fn().mockResolvedValueOnce({ ok: true }).mockRejectedValueOnce(new Error('unreachable'));

        global.fetch = /** @type {typeof fetch} */ (fetchMock);

        const success = await getAPIStatus(minimalConfig);

        expect(success.status).toBe('reachable');

        const failure = await getAPIStatus(minimalConfig);

        expect(failure.status).toBe('error');
        expect(failure.message).toMatch(/unreachable/);
    });
});
