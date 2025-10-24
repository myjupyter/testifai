/**
 * @fileoverview TypeScript type definitions for testifai CLI via JSDoc
 */

/**
 * @typedef {Object} TestifaiConfig
 * @property {ProviderConfig} provider - Provider configuration
 * @property {APIConfig} api - API configuration
 * @property {OutputConfig} output - Output configuration
 * @property {ScanConfig} scan - Scan configuration
 */

/**
 * @typedef {Object} ProviderConfig
 * @property {OpenAIConfig} openai - OpenAI configuration
 */

/**
 * @typedef {Object} OpenAIConfig
 * @property {string} apiKey - OpenAI API key
 */

/**
 * @typedef {Object} APIConfig
 * @property {string} endpoint - API endpoint URL
 */

/**
 * @typedef {Object} OutputConfig
 * @property {string|null} testDir - Directory for test files (null = next to source)
 * @property {string} extension - Test file extension (e.g., '.spec')
 */

/**
 * @typedef {Object} ScanConfig
 * @property {string[]} include - Glob patterns to include
 * @property {string[]} exclude - Glob patterns to exclude
 */

/**
 * @typedef {Object} ScanResult
 * @property {string} file - Source file path
 * @property {number} line - Line number of testifai comment
 * @property {string} comment - Original comment text
 * @property {TestType} testType - Type of test to generate
 * @property {string} additionalParams - Additional parameters from comment
 * @property {string} functionName - Name of the function to test
 * @property {string} functionCode - Complete function code
 * @property {number} functionStartLine - Line where function starts
 * @property {Language} language - Programming language
 */

/**
 * @typedef {'xunit'|'table'|'suite'} TestType
 */

/**
 * @typedef {'ts'|'js'} Language
 */

/**
 * @typedef {Object} APIRequest
 * @property {string} api_key - API key for authentication
 * @property {string} provider - Provider name (e.g., 'OpenAI')
 * @property {string} id - Unique request ID
 * @property {RequestContext} context - Request context
 * @property {TestifaiRequest} testifai - Testifai specific data
 */

/**
 * @typedef {Object} RequestContext
 * @property {Language} language - Programming language
 * @property {string} user_code - Function code to test
 */

/**
 * @typedef {Object} TestifaiRequest
 * @property {TestType} test_type - Type of test to generate
 */

/**
 * @typedef {Object} APIResponse
 * @property {string} id - Request ID
 * @property {GeneratedContent} generated - Generated content
 */

/**
 * @typedef {Object} GeneratedContent
 * @property {string} test_code - Generated test code
 */

/**
 * @typedef {Object} GenerationOptions
 * @property {TestifaiConfig} [config] - Configuration object
 * @property {boolean} [overwrite] - Whether to overwrite existing files
 * @property {boolean} [verbose] - Whether to show verbose output
 * @property {number} [timeout] - Request timeout in milliseconds
 */

/**
 * @typedef {Object} ScanOptions
 * @property {TestType} [type] - Filter by test type
 * @property {boolean} [verbose] - Whether to show verbose output
 */

/**
 * @typedef {Object} InitOptions
 * @property {string} [config] - Custom config file path
 * @property {boolean} [force] - Whether to overwrite existing config
 */

/**
 * @typedef {Object} FileStats
 * @property {string} path - File path
 * @property {number} totalLines - Total number of lines
 * @property {number} codeLines - Number of non-empty lines
 * @property {number} size - File size in bytes
 * @property {Date} created - Creation date
 */

/**
 * @typedef {Object} ValidationResult
 * @property {boolean} valid - Whether the file is valid
 * @property {string[]} errors - Array of error messages
 * @property {string[]} warnings - Array of warning messages
 */

/**
 * @typedef {Object} APIStatus
 * @property {'reachable'|'unreachable'|'error'} status - API status
 * @property {string} endpoint - API endpoint URL
 * @property {string} message - Status message
 * @property {Error} [error] - Error object if status is 'error'
 */

/**
 * @typedef {Object} FunctionInfo
 * @property {string} functionName - Name of the function
 * @property {string} functionCode - Complete function code
 * @property {number} startLine - Starting line number
 * @property {string} line - The line containing function declaration
 */

/**
 * @typedef {Object} ParsedParams
 * @property {string} [type] - Test type parameter
 * @property {boolean} [skip] - Skip this function
 * @property {boolean} [ignore] - Ignore this function
 * @property {number} [timeout] - Timeout for this test
 * @property {boolean} [parallel] - Run test in parallel
 */

export {};
