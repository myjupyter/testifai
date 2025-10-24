import { randomUUID } from 'crypto';
import { getEffectiveConfig, validateConfig } from './config.mjs';

/**
 * @typedef {import('../types.mjs').TestifaiConfig} TestifaiConfig
 * @typedef {import('../types.mjs').ScanResult} ScanResult
 * @typedef {import('../types.mjs').APIRequest} APIRequest
 * @typedef {import('../types.mjs').APIResponse} APIResponse
 * @typedef {import('../types.mjs').APIStatus} APIStatus
 * @typedef {import('../types.mjs').GenerationOptions} GenerationOptions
 */

/**
 * Create API request payload from scan result
 * @param {TestifaiConfig} config - Configuration object
 * @param {ScanResult} scanResult - Scan result from parser
 * @returns {APIRequest} API request payload
 */
function createAPIRequest(config, scanResult) {
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
}

/**
 * Send HTTP request to backend API
 * @param {string} endpoint - API endpoint URL
 * @param {APIRequest} payload - Request payload
 * @param {GenerationOptions} [options={}] - Request options
 * @returns {Promise<APIResponse>} API response
 */
async function makeAPIRequest(endpoint, payload, options = {}) {
    const requestOptions = /** @type {GenerationOptions} */ (options);
    const { timeout = 30000 } = requestOptions;

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
        const response = await fetch(`${endpoint}/generate`, {
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
            throw new Error(`Cannot connect to backend API at ${endpoint}. Is the server running?`);
        }

        throw err;
    }
}

/**
 * Validate API response
 * @param {APIResponse} response - API response to validate
 * @throws {Error} If response is invalid
 */
function validateAPIResponse(response) {
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
}

/**
 * Send scan result to API and get generated tests
 * @param {TestifaiConfig} config - Configuration object
 * @param {ScanResult} scanResult - Scan result from parser
 * @param {GenerationOptions} [options={}] - Request options
 * @returns {Promise<APIResponse>} API response with generated tests
 */
export async function sendToAPI(config, scanResult, options = {}) {
    // Validate inputs
    const effectiveConfig = getEffectiveConfig(config);

    validateConfig(effectiveConfig);

    if (!scanResult || !scanResult.functionCode || !scanResult.testType) {
        throw new Error('Invalid scan result: missing required fields');
    }

    const generationOptions = /** @type {GenerationOptions} */ (options);
    const { verbose = false, timeout = 30000 } = generationOptions;

    try {
        // Create request payload
        const payload = createAPIRequest(effectiveConfig, scanResult);

        if (verbose) {
            console.log(`Sending request to ${effectiveConfig.api.endpoint}/generate`);
            console.log(`Function: ${scanResult.functionName}`);
            console.log(`Test type: ${scanResult.testType}`);
            console.log(`Language: ${scanResult.language}`);
        }

        // Make API request
        const response = await makeAPIRequest(effectiveConfig.api.endpoint, payload, generationOptions);

        // Validate response
        validateAPIResponse(response);

        if (verbose) {
            console.log(`Received response with ${response.generated.test_code.length} characters of test code`);
        }

        return response;
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        // Add context to error messages
        const contextualError = new Error(`Failed to generate tests for ${scanResult.functionName}: ${err.message}`);

        /** @type {Error & { originalError?: Error; scanResult?: ScanResult }} */ (contextualError).originalError = err;
        /** @type {Error & { originalError?: Error; scanResult?: ScanResult }} */ (contextualError).scanResult =
            scanResult;
        throw contextualError;
    }
}

/**
 * Test API connection
 * @param {TestifaiConfig} config - Configuration object
 * @returns {Promise<boolean>} Whether API is reachable
 */
export async function testAPIConnection(config) {
    const effectiveConfig = getEffectiveConfig(config);

    try {
        // Try a simple request to see if the server is running
        await fetch(effectiveConfig.api.endpoint, {
            method: 'GET',
            headers: { Accept: 'application/json' },
        });

        return true;
    } catch (error) {
        return false;
    }
}

/**
 * Get API status and information
 * @param {TestifaiConfig} config - Configuration object
 * @returns {Promise<APIStatus>} API status information
 */
export async function getAPIStatus(config) {
    const effectiveConfig = getEffectiveConfig(config);
    const { endpoint } = effectiveConfig.api;

    try {
        const isReachable = await testAPIConnection(effectiveConfig);

        if (!isReachable) {
            return {
                status: 'unreachable',
                endpoint,
                message: 'Cannot connect to API endpoint',
            };
        }

        return {
            status: 'reachable',
            endpoint,
            message: 'API endpoint is reachable',
        };
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        return {
            status: 'error',
            endpoint,
            message: err.message,
            error: err,
        };
    }
}
