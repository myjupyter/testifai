import { readFileSync, readdirSync, statSync } from 'fs';
import { join, extname, relative } from 'path';

/**
 * @typedef {import('../types.mjs').ScanResult} ScanResult
 * @typedef {import('../types.mjs').ScanOptions} ScanOptions
 * @typedef {import('../types.mjs').ScanConfig} ScanConfig
 * @typedef {import('../types.mjs').FunctionInfo} FunctionInfo
 * @typedef {import('../types.mjs').ParsedParams} ParsedParams
 */

/**
 * Regular expression to match testifai comments
 * Matches: // testifai: -type=xunit|table|suite
 */
const TESTIFAI_COMMENT_REGEX = /\/\/\s*testifai:\s*-type=(xunit|table|suite)(?:\s+(.+))?/;

/**
 * Regular expressions for different function types
 */
const FUNCTION_PATTERNS = [
    // Regular function declaration: function name() {} (with optional TypeScript types)
    /function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*:\s*[^{]*\{/,
    /function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{/,
    // Arrow function: const name = () => {}
    /(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*\{/,
    // Method in class: methodName() {}
    /(?:async\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*:\s*[^{]*\{/,
    /(?:async\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{/,
    // Export function: export function name() {}
    /export\s+(?:default\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*:\s*[^{]*\{/,
    /export\s+(?:default\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{/,
    // Export arrow function: export const name = () => {}
    /export\s+(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*\{/,
];

/**
 * Check if a path should be included in scanning
 * @param {string} filePath - File path to check
 * @param {ScanConfig} scanConfig - Scan configuration
 * @returns {boolean} Whether to include the file
 */
function shouldIncludeFile(
    filePath,
    scanConfig = { include: ['**/*.ts', '**/*.js'], exclude: ['node_modules/**', '**/*.spec.*', '**/*.test.*'] },
) {
    const effectiveScanConfig = /** @type {ScanConfig} */ (scanConfig);
    const { include = ['**/*.ts', '**/*.js'], exclude = ['node_modules/**', '**/*.spec.*', '**/*.test.*'] } =
        effectiveScanConfig;

    // Check file extension
    const ext = extname(filePath);

    if (!['.ts', '.js', '.tsx', '.jsx'].includes(ext)) {
        return false;
    }

    // Simple pattern matching (could be improved with glob library)
    const relativePath = relative(process.cwd(), filePath);

    // Check exclude patterns
    for (const pattern of exclude) {
        if (relativePath.includes(pattern.replace('**/', '').replace('/**', ''))) {
            return false;
        }
    }

    return true;
}

/**
 * Get all TypeScript and JavaScript files in a directory
 * @param {string} dirPath - Directory path
 * @param {ScanConfig} scanConfig - Scan configuration
 * @returns {string[]} Array of file paths
 */
function getFiles(
    dirPath,
    scanConfig = { include: ['**/*.ts', '**/*.js'], exclude: ['node_modules/**', '**/*.spec.*', '**/*.test.*'] },
) {
    const effectiveScanConfig = /** @type {ScanConfig} */ (scanConfig);
    const files = /** @type {string[]} */ ([]);

    /**
     * @param {string} currentPath
     */
    function traverse(currentPath) {
        try {
            const stats = statSync(currentPath);

            if (stats.isDirectory()) {
                // Skip node_modules and common excluded directories
                const dirName = currentPath.split('/').pop() ?? '';

                if (['node_modules', '.git', 'dist', 'build', '.next'].includes(dirName)) {
                    return;
                }

                const entries = readdirSync(currentPath);

                for (const entry of entries) {
                    traverse(join(currentPath, entry));
                }
            } else if (stats.isFile() && shouldIncludeFile(currentPath, effectiveScanConfig)) {
                files.push(currentPath);
            }
        } catch (error) {
            const err = error instanceof Error ? error : new Error(String(error));

            if (process.env.NODE_ENV !== 'test') {
                console.warn(`Warning: Cannot access ${currentPath}: ${err.message}`);
            }
        }
    }

    traverse(dirPath);

    return files;
}

/**
 * Extract function code starting from a given line
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @returns {FunctionInfo|null} Function info or null if not found
 */
function extractFunctionCode(lines, startLine) {
    for (let i = startLine; i < lines.length; i++) {
        const line = lines[i];

        // Try each function pattern
        for (const pattern of FUNCTION_PATTERNS) {
            const match = line.match(pattern);

            if (match) {
                const functionName = match[1];

                // Find the opening brace and extract the complete function
                const functionCode = extractCompleteFunction(lines, i);

                return {
                    functionName,
                    functionCode,
                    startLine: i + 1, // Convert to 1-based
                    line: line.trim(),
                };
            }
        }
    }

    return null;
}

/**
 * Extract complete function code by matching braces
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @returns {string} Complete function code
 */
function extractCompleteFunction(lines, startLine) {
    const functionLines = [];
    let braceCount = 0;
    let started = false;

    for (let i = startLine; i < lines.length; i++) {
        const line = lines[i];

        functionLines.push(line);

        // Count braces to find function boundaries
        for (const char of line) {
            if (char === '{') {
                braceCount++;
                started = true;
            } else if (char === '}') {
                braceCount--;
            }
        }

        // Function ends when braces are balanced
        if (started && braceCount === 0) {
            break;
        }
    }

    return functionLines.join('\n');
}

/**
 * Parse a single file for testifai comments
 * @param {string} filePath - Path to the file
 * @param {ScanOptions} options - Parse options
 * @returns {ScanResult[]} Array of parsed results
 */
function parseFile(filePath, options = { verbose: false }) {
    /** @type {ScanOptions} */
    const parseOptions = options;

    try {
        const content = readFileSync(filePath, 'utf8');
        const lines = content.split('\n');
        const results = /** @type {ScanResult[]} */ ([]);

        // Find all testifai comments
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const commentMatch = line.match(TESTIFAI_COMMENT_REGEX);

            if (commentMatch) {
                const testType = commentMatch[1];
                const additionalParams = commentMatch[2] || '';

                if (parseOptions.verbose) {
                    console.log(`Found testifai comment: type=${testType}, params=${additionalParams}`);
                }

                // Filter by type if specified
                if (parseOptions.type && testType !== parseOptions.type) {
                    continue;
                }

                // Look for the next function after the comment
                const functionInfo = extractFunctionCode(lines, i + 1);

                if (functionInfo) {
                    if (parseOptions.verbose) {
                        console.log(`Found function: ${functionInfo.functionName}`);
                    }

                    results.push({
                        file: filePath,
                        line: i + 1, // Convert to 1-based line number
                        comment: line.trim(),
                        testType: /** @type {import('../types.mjs').TestType} */ (testType),
                        additionalParams,
                        functionName: functionInfo.functionName,
                        functionCode: functionInfo.functionCode,
                        functionStartLine: functionInfo.startLine,
                        language: /** @type {import('../types.mjs').Language} */ (
                            filePath.endsWith('.ts') || filePath.endsWith('.tsx') ? 'ts' : 'js'
                        ),
                    });
                } else {
                    if (parseOptions.verbose) {
                        console.warn(`Warning: testifai comment found but no function follows at ${filePath}:${i + 1}`);
                    }
                }
            }
        }

        return results;
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        if (parseOptions.verbose) {
            console.error(`Error parsing file ${filePath}:`, err.message);
        }

        return [];
    }
}

/**
 * Scan files for testifai comments
 * @param {string} targetPath - File or directory path to scan
 * @param {ScanOptions} options - Scan options
 * @returns {Promise<ScanResult[]>} Array of scan results
 */
export async function scanFiles(targetPath, options = { verbose: false }) {
    /** @type {ScanOptions} */
    const scanOptions = options;
    const results = [];

    try {
        const stats = statSync(targetPath);

        if (stats.isFile()) {
            // Single file
            if (shouldIncludeFile(targetPath)) {
                results.push(...parseFile(targetPath, scanOptions));
            }
        } else if (stats.isDirectory()) {
            // Directory - get all relevant files
            const files = getFiles(targetPath);

            if (scanOptions.verbose) {
                console.log(`Scanning ${files.length} files...`);
            }

            for (const file of files) {
                results.push(...parseFile(file, scanOptions));
            }
        }
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        throw new Error(`Cannot access path: ${targetPath} - ${err.message}`);
    }

    return results;
}

/**
 * Parse testifai comment parameters
 * @param {string} paramString - Parameter string from comment
 * @returns {ParsedParams} Parsed parameters
 */
export function parseCommentParams(paramString) {
    const params = /** @type {ParsedParams} */ ({});

    if (!paramString) {
        return params;
    }

    // Simple parameter parsing: -key=value -flag
    const paramMatches = paramString.matchAll(/-([a-zA-Z]+)(?:=([^\\s]+))?/g);

    for (const match of paramMatches) {
        const key = match[1].toLowerCase();
        const rawValue = match[2] || true;

        switch (key) {
            case 'type':
                if (typeof rawValue === 'string') {
                    params.type = rawValue;
                }

                break;
            case 'skip':
                params.skip = rawValue === true || String(rawValue).toLowerCase() === 'true';
                break;
            case 'ignore':
                params.ignore = rawValue === true || String(rawValue).toLowerCase() === 'true';
                break;
            case 'timeout': {
                const timeoutValue = typeof rawValue === 'string' ? Number.parseInt(rawValue, 10) : Number(rawValue);

                if (!Number.isNaN(timeoutValue)) {
                    params.timeout = timeoutValue;
                }

                break;
            }
            case 'parallel':
                params.parallel = rawValue === true || String(rawValue).toLowerCase() === 'true';
                break;
            default:
                break;
        }
    }

    return params;
}

/**
 * Validate scan result
 * @param {ScanResult} result - Scan result to validate
 * @returns {boolean} Whether the result is valid
 */
export function validateScanResult(result) {
    return (
        Boolean(result.file) &&
        Boolean(result.testType) &&
        Boolean(result.functionName) &&
        Boolean(result.functionCode) &&
        Boolean(result.language) &&
        ['xunit', 'table', 'suite'].includes(result.testType) &&
        ['ts', 'js'].includes(result.language)
    );
}
