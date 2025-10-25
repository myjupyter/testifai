import { existsSync, mkdirSync, readFileSync, statSync, writeFileSync } from 'fs';
import { basename, dirname, extname, join, relative } from 'path';
import { getEffectiveConfig } from './config.mjs';

/**
 * @typedef {import('../types.mjs').TestifaiConfig} TestifaiConfig
 * @typedef {import('../types.mjs').ScanResult} ScanResult
 * @typedef {import('../types.mjs').APIResponse} APIResponse
 * @typedef {import('../types.mjs').GenerationOptions} GenerationOptions
 * @typedef {import('../types.mjs').FileStats} FileStats
 * @typedef {import('../types.mjs').ValidationResult} ValidationResult
 */

/**
 * Get test file path based on source file and configuration
 * @param {string} sourceFile - Source file path
 * @param {TestifaiConfig} config - Configuration object
 * @returns {string} Test file path
 */
function getTestFilePath(sourceFile, config = getEffectiveConfig()) {
    const { output } = config;
    const { testDir, extension } = output;

    const sourceDir = dirname(sourceFile);
    const sourceName = basename(sourceFile, extname(sourceFile));
    const sourceExt = extname(sourceFile);

    // Determine test file name
    const testFileName = `${sourceName}${extension}${sourceExt}`;

    if (testDir) {
        // Put tests in a specific directory
        const relativePath = sourceFile.replace(process.cwd(), '').replace(/^\//, '');

        return join(testDir, dirname(relativePath), testFileName);
    }

    // Default: put test in __tests__ subdirectory next to source file
    return join(sourceDir, '__tests__', testFileName);
}

/**
 * Ensure directory exists
 * @param {string} filePath - File path
 */
function ensureDirectoryExists(filePath) {
    const dir = dirname(filePath);

    if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true });
    }
}

/**
 * Format generated test code
 * @param {string} testCode - Raw test code from API
 * @param {ScanResult} scanResult - Original scan result
 * @param {string} testFilePath - Target test file path
 * @returns {string} Formatted test code
 */
function formatTestCode(testCode, scanResult, testFilePath) {
    const { functionName, file } = scanResult;

    // Add header comment
    const header = `// Generated tests for ${functionName} from ${basename(file)}
// Created with testifai - https://github.com/myjupyter/testifai

`;

    // Clean up the test code
    let formattedCode = testCode.trim();

    // Ensure proper imports for TypeScript
    if (scanResult.language === 'ts' && !formattedCode.includes('import')) {
        // Add basic imports if not present
        const testDir = dirname(testFilePath);
        const sourceWithoutExt = file.slice(0, -extname(file).length);
        let importPath = relative(testDir, sourceWithoutExt);

        if (!importPath.startsWith('.')) {
            importPath = `./${importPath}`;
        }

        importPath = importPath.replace(/\\/g, '/');
        const importLine = `import { ${functionName} } from '${importPath}';\n\n`;

        formattedCode = importLine + formattedCode;
    }

    return header + formattedCode + '\n';
}

/**
 * Check if test file already exists and handle conflicts
 * @param {string} testFilePath - Test file path
 * @param {GenerationOptions} [options={}] - Options for handling conflicts
 * @returns {boolean} Whether to proceed with writing
 */
function handleExistingFile(testFilePath, options = {}) {
    if (!existsSync(testFilePath)) {
        return true;
    }

    const generationOptions = /** @type {GenerationOptions} */ (options);
    const { overwrite = false, verbose = false } = generationOptions;

    if (overwrite) {
        if (verbose) {
            console.log(`Overwriting existing test file: ${testFilePath}`);
        }

        return true;
    } else {
        if (verbose) {
            console.log(`Test file already exists, skipping: ${testFilePath}`);
        }

        return false;
    }
}

/**
 * Generate test file from API response
 * @param {ScanResult} scanResult - Original scan result
 * @param {APIResponse} apiResponse - Response from API
 * @param {GenerationOptions} [options={}] - Generation options
 * @returns {Promise<string>} Path to generated test file
 */
export async function generateTestFiles(scanResult, apiResponse, options = {}) {
    const generationOptions = /** @type {GenerationOptions} */ (options);
    const { config: providedConfig, overwrite = false, verbose = false } = generationOptions;
    const effectiveConfig = getEffectiveConfig(providedConfig);

    try {
        // Validate inputs
        if (!scanResult || !scanResult.file || !scanResult.functionName) {
            throw new Error('Invalid scan result');
        }

        if (!apiResponse || !apiResponse.generated || !apiResponse.generated.test_code) {
            throw new Error('Invalid API response');
        }

        // Determine test file path
        const testFilePath = getTestFilePath(scanResult.file, effectiveConfig);

        // Check if file already exists
        if (!handleExistingFile(testFilePath, { overwrite, verbose })) {
            return testFilePath; // File exists and not overwriting
        }

        // Ensure directory exists
        ensureDirectoryExists(testFilePath);

        // Format the test code
        const formattedCode = formatTestCode(apiResponse.generated.test_code, scanResult, testFilePath);

        // Write the test file
        writeFileSync(testFilePath, formattedCode, 'utf8');

        if (verbose) {
            console.log(`✅ Generated test file: ${testFilePath}`);
            console.log(`   Function: ${scanResult.functionName}`);
            console.log(`   Test type: ${scanResult.testType}`);
            console.log(`   Lines: ${formattedCode.split('\n').length}`);
        }

        return testFilePath;
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        throw new Error(`Failed to generate test file: ${err.message}`);
    }
}

/**
 * Generate multiple test files from scan results
 * @param {ScanResult[]} scanResults - Array of scan results
 * @param {APIResponse[]} apiResponses - Array of API responses
 * @param {GenerationOptions} [options={}] - Generation options
 * @returns {Promise<string[]>} Array of generated file paths
 */
export async function generateMultipleTestFiles(scanResults, apiResponses, options = {}) {
    if (scanResults.length !== apiResponses.length) {
        throw new Error('Scan results and API responses count mismatch');
    }

    const generationOptions = /** @type {GenerationOptions} */ (options);
    const { verbose = false } = generationOptions;

    const tasks = scanResults.map(async (result, index) => {
        try {
            return await generateTestFiles(result, apiResponses[index], generationOptions);
        } catch (error) {
            const err = error instanceof Error ? error : new Error(String(error));

            if (verbose) {
                console.error(`Failed to generate test for ${result.functionName}: ${err.message}`);
            }

            return null;
        }
    });

    const generatedFiles = await Promise.all(tasks);

    return generatedFiles.filter((filePath) => typeof filePath === 'string');
}

/**
 * Get test file statistics
 * @param {string} testFilePath - Test file path
 * @returns {FileStats|null} File statistics
 */
export function getTestFileStats(testFilePath) {
    if (!existsSync(testFilePath)) {
        return null;
    }

    try {
        const content = readFileSync(testFilePath, 'utf8');
        const lines = content.split('\n');
        const nonEmptyLines = lines.filter((line) => line.trim().length > 0);

        return {
            path: testFilePath,
            totalLines: lines.length,
            codeLines: nonEmptyLines.length,
            size: content.length,
            created: statSync(testFilePath).birthtime,
        };
    } catch {
        return null;
    }
}

/**
 * Validate generated test file
 * @param {string} testFilePath - Test file path
 * @returns {ValidationResult} Validation result
 */
export function validateTestFile(testFilePath) {
    const result = /** @type {ValidationResult} */ ({
        valid: false,
        errors: [],
        warnings: [],
    });

    if (!existsSync(testFilePath)) {
        result.errors.push('Test file does not exist');

        return result;
    }

    try {
        const content = readFileSync(testFilePath, 'utf8');

        // Basic validation
        if (content.trim().length === 0) {
            result.errors.push('Test file is empty');
        }

        // Check for common test patterns
        const hasDescribe = content.includes('describe(') || content.includes('describe ');
        const hasTest =
            content.includes('test(') ||
            content.includes('it(') ||
            content.includes('test ') ||
            content.includes('it ');

        if (!hasDescribe && !hasTest) {
            result.warnings.push('No test blocks found (describe/test/it)');
        }

        // Check for imports
        const hasImports = content.includes('import ') || content.includes('require(');

        if (!hasImports) {
            result.warnings.push('No imports found');
        }

        if (result.errors.length === 0) {
            result.valid = true;
        }
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        result.errors.push(`Cannot read test file: ${err.message}`);
    }

    return result;
}
