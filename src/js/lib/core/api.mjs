import { randomUUID } from 'crypto';
import { getEffectiveConfig, validateConfig } from './config.mjs';
import { TESTIFAI_URL } from '../constants.mjs';
import { request } from '../utils/request.mjs';

/**
 * @typedef {import('../types.mjs').TestifaiConfig} TestifaiConfig
 * @typedef {import('../types.mjs').ScanResult} ScanResult
 * @typedef {import('../types.mjs').APIRequest} APIRequest
 * @typedef {import('../types.mjs').APIResponse} APIResponse
 * @typedef {import('../types.mjs').GenerationOptions} GenerationOptions
 */

/**
 * Create API request payload from scan result
 * @param {TestifaiConfig} config - Configuration object
 * @param {ScanResult} scanResult - Scan result from parser
 * @returns {APIRequest} API request payload
 */
const createAPIRequest = (config, scanResult) => {
    const requestId = randomUUID();

    return {
        id: requestId,
        context: {
            platform: scanResult.platform,
            user_code: scanResult.functionCode,
        },
        testifai: {
            test_type: scanResult.testType,
        },
    };
};

/**
 * Send HTTP request to backend API
 * @param {APIRequest} payload - Request payload
 * @param {GenerationOptions} [options={}] - Request options
 * @returns {Promise<APIResponse>} API response
 */
const makeAPIRequest = async (payload, options = {}) => {
    const requestOptions = /** @type {GenerationOptions} */ (options);
    const { timeout = 30000 } = requestOptions;

    try {
        console.log('🔄 Starting request.post...');
        const response = await request.post('generate', {
            json: payload,
            timeout: timeout,
        });

        console.log('✅ Got response, parsing JSON...');

        return await response.json();
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        if ('cause' in err) {
            console.log('Error Cause:', err.cause);
        }

        if (err.message.includes('bad port') || err.message.includes('6667')) {
            throw new Error(`CRITICAL: ky failed to bypass "bad port" restriction: ${err.message}`);
        }

        if (err.name === 'HTTPError') {
            /** @type {any} */
            const { response } = err;

            const errorText = await response.text();

            throw new Error(`API request failed: ${response.status} ${response.statusText}\n${errorText}`);
        }

        if (err.name === 'TimeoutError') {
            throw new Error(`API request timed out after ${timeout}ms`);
        }

        if ('code' in err && err.code === 'ECONNREFUSED') {
            throw new Error(`Cannot connect to backend API at ${TESTIFAI_URL}. Is the server running?`);
        }

        if ('code' in err && (err.code === 'ENOTFOUND' || err.code === 'ECONNRESET')) {
            throw new Error(`Network error (${err.code}): ${err.message}`);
        }

        if (err.name === 'RequestError') {
            throw new Error(`Request failed: ${err.message}`);
        }

        console.log('🔧 Unhandled error type, rethrowing original');
        throw err;
    }
};

/**
 * Validate API response
 * @param {APIResponse} response - API response to validate
 * @throws {Error} If response is invalid
 */
const validateAPIResponse = (response) => {
    if (!response) {
        throw new Error('Empty response from API');
    }

    if (!response.id) {
        throw new Error('Response missing required field: id');
    }

    if (!response.generated || !response.generated.test_code) {
        throw new Error('Response missing generated test code');
    }

    if (typeof response.generated.test_code !== 'string') {
        throw new Error('Generated test code must be a string');
    }

    if (response.generated.test_code.trim().length === 0) {
        throw new Error('Generated test code is empty');
    }
};

/**
 * Send scan result to API and get generated tests
 * @param {TestifaiConfig} config - Configuration object
 * @param {ScanResult} scanResult - Scan result from parser
 * @param {GenerationOptions} [options={}] - Request options
 * @returns {Promise<APIResponse>} API response with generated tests
 */
export const sendToAPI = async (config, scanResult, options = {}) => {
    const effectiveConfig = getEffectiveConfig(config);

    validateConfig(effectiveConfig);

    if (!scanResult || !scanResult.functionCode || !scanResult.testType) {
        throw new Error('Invalid scan result: missing required fields');
    }

    const generationOptions = /** @type {GenerationOptions} */ (options);
    const { verbose = false } = generationOptions;

    try {
        const payload = createAPIRequest(effectiveConfig, scanResult);

        if (verbose) {
            console.log(`Sending request to ${TESTIFAI_URL}/generate`);
            console.log(`Function: ${scanResult.functionName}`);
            console.log(`Test type: ${scanResult.testType}`);
            console.log(`Language: ${scanResult.platform}`);
        }

        console.log('1)');

        const response = await makeAPIRequest(payload, generationOptions);

        console.log('2)');

        validateAPIResponse(response);

        if (verbose) {
            console.log(`Received response with ${response.generated.test_code.length} characters of test code`);
        }

        return response;
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        const contextualError = new Error(`Failed to generate tests for ${scanResult.functionName}: ${err.message}`);

        /** @type {Error & { originalError?: Error; scanResult?: ScanResult }} */ (contextualError).originalError = err;
        /** @type {Error & { originalError?: Error; scanResult?: ScanResult }} */ (contextualError).scanResult =
            scanResult;

        console.log('🔧 Rethrowing contextual error');
        throw contextualError;
    }
};
