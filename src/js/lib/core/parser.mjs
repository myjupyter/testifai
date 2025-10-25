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
 * Regular expression to match generator directives
 * Matches: // js:generate testifai: -type=xunit ...
 *          // ts:generate testifai: -type=table ...
 */
const GENERATE_COMMENT_REGEX = /^\s*\/\/\s*(js|ts):generate\s+testifai:\s*-type=([a-zA-Z]+)(.*)$/i;

/**
 * Regular expressions for different function and class types
 */
const FUNCTION_PATTERNS = [
    // Classes: class Name, export class Name, abstract class Name
    {
        pattern: /(?:export\s+)?(?:abstract\s+)?class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)/,
        type: 'class',
    },

    // Short arrow functions: const fn = (a, b) => a + b;
    {
        pattern:
            /(?:export\s+)?(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*([^{;]+)[;}]/,
        type: 'short-arrow',
    },

    // Arrow functions with body: const fn = (a, b) => { return a + b; }
    {
        pattern:
            /(?:export\s+)?(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*\{/,
        type: 'arrow',
    },

    // Regular function declaration: function name() {} (with optional TypeScript types)
    {
        pattern: /(?:export\s+)?(?:async\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*:\s*[^{]*\{/,
        type: 'function',
    },
    {
        pattern: /(?:export\s+)?(?:async\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{/,
        type: 'function',
    },

    // Class methods: methodName() {}, static methodName() {}, async methodName() {}
    {
        pattern: /(?:static\s+)?(?:async\s+)?(?:get\s+|set\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*:\s*[^{]*\{/,
        type: 'method',
    },
    {
        pattern: /(?:static\s+)?(?:async\s+)?(?:get\s+|set\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{/,
        type: 'method',
    },

    // Constructor
    {
        pattern: /constructor\s*\([^)]*\)\s*\{/,
        type: 'constructor',
        name: 'constructor',
    },
];

/**
 * Check whether directive and type are allowed for a given language
 * @param {import('../types.mjs').Platform} platform
 * @param {string} directive
 * @param {string} testType
 * @returns {boolean}
 */
function isSupportedDirective(platform, directive, testType) {
    if (platform === 'ts') {
        return directive === 'ts' && testType === 'table';
    }

    return directive === 'js' && testType === 'xunit';
}

/**
 * Check if a path should be included in scanning
 * @param {string} filePath - File path to check
 * @param {ScanConfig} scanConfig - Scan configuration
 * @returns {boolean} Whether to include the file
 */
function shouldIncludeFile(
    filePath,
    scanConfig = {
        include: ['**/*.ts', '**/*.js'],
        exclude: ['node_modules/**', '**/*.spec.*', '**/*.test.*'],
    },
) {
    const effectiveScanConfig = /** @type {ScanConfig} */ (scanConfig);
    const { exclude = ['node_modules/**', '**/*.spec.*', '**/*.test.*'] } = effectiveScanConfig;

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
    scanConfig = {
        include: ['**/*.ts', '**/*.js'],
        exclude: ['node_modules/**', '**/*.spec.*', '**/*.test.*'],
    },
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
 * Extract function or class code starting from a given line
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @returns {FunctionInfo|null} Function info or null if not found
 */
function extractFunctionCode(lines, startLine) {
    for (let i = startLine; i < lines.length; i++) {
        const line = lines[i];

        // Try each function pattern
        for (const { pattern, name, type } of FUNCTION_PATTERNS) {
            const match = line.match(pattern);

            if (match) {
                const functionName = name || match[1] || 'anonymous';
                const functionCode = extractCodeByType(lines, i, type, match);

                return {
                    functionName,
                    functionCode,
                    startLine: i + 1,
                    line: line.trim(),
                };
            }
        }
    }

    return null;
}

/**
 * Extract code based on function/class type
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @param {string} type - Type of function/class
 * @param {RegExpMatchArray} match - Regex match result
 * @returns {string} Complete code
 */
function extractCodeByType(lines, startLine, type, match) {
    switch (type) {
        case 'short-arrow':
            return extractShortArrowFunction(lines, startLine, match);
        case 'class':
            return extractClassCode(lines, startLine);
        case 'arrow':
        case 'function':
        case 'method':
        case 'constructor':
        default:
            return extractBracedCode(lines, startLine);
    }
}

/**
 * Extract short arrow function code
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @param {RegExpMatchArray} _match - Regex match result (unused)
 * @returns {string} Complete arrow function code
 */
function extractShortArrowFunction(lines, startLine, _match) {
    const line = lines[startLine];

    // For short arrow functions, the entire function is usually on one line
    // But we might need to handle multi-line cases
    let code = line;

    // Check if the line ends with semicolon or is complete
    if (line.trim().endsWith(';') || line.trim().endsWith(',')) {
        return code.trim();
    }

    // If not complete, look for continuation
    for (let i = startLine + 1; i < lines.length; i++) {
        const nextLine = lines[i];

        code += '\n' + nextLine;

        if (nextLine.trim().endsWith(';') || nextLine.trim().endsWith(',')) {
            break;
        }
    }

    return code.trim();
}

/**
 * Extract class code including all methods
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @returns {string} Complete class code
 */
function extractClassCode(lines, startLine) {
    return extractBracedCode(lines, startLine);
}

/**
 * Extract complete code by matching braces (for functions, classes, methods)
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @returns {string} Complete code
 */
function extractBracedCode(lines, startLine) {
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
    const fileLanguage = filePath.endsWith('.ts') || filePath.endsWith('.tsx') ? 'ts' : 'js';

    try {
        const content = readFileSync(filePath, 'utf8');
        const lines = content.split('\n');
        const results = /** @type {ScanResult[]} */ ([]);

        // Find all testifai comments
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            const commentMatch = line.match(GENERATE_COMMENT_REGEX);

            if (!commentMatch) {
                continue;
            }

            const directive = commentMatch[1].toLowerCase();
            const testType = commentMatch[2].toLowerCase();
            const additionalParamsRaw = commentMatch[3] || '';

            if (!isSupportedDirective(fileLanguage, directive, testType)) {
                continue;
            }

            const additionalParams = additionalParamsRaw.trim();

            if (parseOptions.verbose) {
                console.log(
                    `Found testifai comment: language=${fileLanguage}, directive=${directive}, type=${testType}, params=${additionalParams}`,
                );
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
                    platform: /** @type {import('../types.mjs').Platform} */ (fileLanguage),
                });

                // If this is a class, also parse inside it for method comments
                if (
                    functionInfo.functionCode.trim().startsWith('class ') ||
                    functionInfo.functionCode.trim().startsWith('export class ')
                ) {
                    const classMethodResults = parseInsideClass(lines, i + 1, filePath, fileLanguage, parseOptions);

                    results.push(...classMethodResults);
                }
            } else if (parseOptions.verbose) {
                console.warn(`Warning: testifai comment found but no function follows at ${filePath}:${i + 1}`);
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
 * Parse comments inside class body for methods
 * @param {string[]} lines - Array of file lines
 * @param {number} classStartLine - Line where class starts (0-based)
 * @param {string} filePath - File path
 * @param {import('../types.mjs').Platform} fileLanguage - File language
 * @param {ScanOptions} parseOptions - Parse options
 * @returns {ScanResult[]} Array of scan results for class methods
 */
function parseInsideClass(lines, classStartLine, filePath, fileLanguage, parseOptions) {
    /** @type {ScanResult[]} */
    const results = [];

    // Find class boundaries
    let braceCount = 0;
    let started = false;
    let classEndLine = lines.length;

    for (let i = classStartLine; i < lines.length; i++) {
        const line = lines[i];

        for (const char of line) {
            if (char === '{') {
                braceCount++;
                started = true;
            } else if (char === '}') {
                braceCount--;
            }
        }

        if (started && braceCount === 0) {
            classEndLine = i;
            break;
        }
    }

    // Look for comments inside class body
    for (let i = classStartLine + 1; i < classEndLine; i++) {
        const line = lines[i];
        const commentMatch = line.match(GENERATE_COMMENT_REGEX);

        if (!commentMatch) {
            continue;
        }

        const directive = commentMatch[1].toLowerCase();
        const testType = commentMatch[2].toLowerCase();
        const additionalParamsRaw = commentMatch[3] || '';

        if (!isSupportedDirective(fileLanguage, directive, testType)) {
            continue;
        }

        const additionalParams = additionalParamsRaw.trim();

        if (parseOptions.verbose) {
            console.log(
                `Found testifai comment inside class: language=${fileLanguage}, directive=${directive}, type=${testType}, params=${additionalParams}`,
            );
        }

        // Filter by type if specified
        if (parseOptions.type && testType !== parseOptions.type) {
            continue;
        }

        // Look for the next method after the comment (within class boundaries)
        const methodInfo = extractMethodFromClass(lines, i + 1, classEndLine);

        if (methodInfo) {
            if (parseOptions.verbose) {
                console.log(`Found method: ${methodInfo.functionName}`);
            }

            results.push({
                file: filePath,
                line: i + 1, // Convert to 1-based line number
                comment: line.trim(),
                testType: /** @type {import('../types.mjs').TestType} */ (testType),
                additionalParams,
                functionName: methodInfo.functionName,
                functionCode: methodInfo.functionCode,
                functionStartLine: methodInfo.startLine,
                platform: /** @type {import('../types.mjs').Platform} */ (fileLanguage),
            });
        } else if (parseOptions.verbose) {
            console.warn(`Warning: testifai comment found but no method follows at ${filePath}:${i + 1}`);
        }
    }

    return results;
}

/**
 * Extract method code from class (limited to class boundaries)
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @param {number} classEndLine - End line of class (0-based)
 * @returns {FunctionInfo|null} Method info or null if not found
 */
function extractMethodFromClass(lines, startLine, classEndLine) {
    for (let i = startLine; i < Math.min(classEndLine, lines.length); i++) {
        const line = lines[i];

        // Try method patterns (skip class pattern since we're inside a class)
        for (const patternObj of FUNCTION_PATTERNS) {
            if (patternObj.type === 'class') {
                continue; // Skip class patterns when looking for methods
            }

            const match = line.match(patternObj.pattern);

            if (match) {
                const functionName = patternObj.name || match[1] || 'anonymous';

                // Extract method code (limited to class boundaries)
                const functionCode = extractMethodCode(lines, i, classEndLine);

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
 * Extract method code within class boundaries
 * @param {string[]} lines - Array of file lines
 * @param {number} startLine - Starting line number (0-based)
 * @param {number} classEndLine - End line of class (0-based)
 * @returns {string} Complete method code
 */
function extractMethodCode(lines, startLine, classEndLine) {
    const methodLines = [];
    let braceCount = 0;
    let started = false;

    for (let i = startLine; i < Math.min(classEndLine, lines.length); i++) {
        const line = lines[i];

        methodLines.push(line);

        // Count braces to find method boundaries
        for (const char of line) {
            if (char === '{') {
                braceCount++;
                started = true;
            } else if (char === '}') {
                braceCount--;
            }
        }

        // Method ends when braces are balanced
        if (started && braceCount === 0) {
            break;
        }
    }

    return methodLines.join('\n');
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
    const paramMatches = paramString.matchAll(/-([a-zA-Z]+)(?:=([^\s]+))?/g);

    for (const match of paramMatches) {
        const key = match[1].toLowerCase();
        const valueSegment = match[2];
        const rawValue = typeof valueSegment === 'string' ? valueSegment.trim() : true;

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
        Boolean(result.platform) &&
        ['xunit', 'table', 'suite'].includes(result.testType) &&
        ['ts', 'js'].includes(result.platform)
    );
}
