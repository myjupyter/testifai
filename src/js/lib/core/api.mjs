import { randomUUID } from 'crypto';
import { getEffectiveConfig, validateConfig } from './config.mjs';
import { TESTIFAI_URL } from '../constants.mjs';

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
        api_key: config.provider.openai.apiKey,
        provider: 'OpenAI',
        id: requestId,
        context: {
            language: scanResult.language,
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

    const controller = new AbortController();

    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
        const response = await fetch(`${TESTIFAI_URL}/generate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Accept: 'application/json',
            },
            body: JSON.stringify(payload),
            signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
            const errorText = await response.text();

            throw new Error(`API request failed: ${response.status} ${response.statusText}\n${errorText}`);
        }

        return await response.json();
    } catch (error) {
        clearTimeout(timeoutId);
        const err = error instanceof Error ? error : new Error(String(error));

        if (err.name === 'AbortError') {
            throw new Error(`API request timed out after ${timeout}ms`);
        }

        if ('code' in err && err.code === 'ECONNREFUSED') {
            throw new Error(`Cannot connect to backend API at ${TESTIFAI_URL}. Is the server running?`);
        }

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
    // Validate inputs
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
            console.log(`Language: ${scanResult.language}`);
        }

        // Make API request
        const response = await makeAPIRequest(payload, generationOptions);

        // Validate response
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
        throw contextualError;
    }
};
