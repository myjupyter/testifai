import { statSync, existsSync } from 'fs';
import { join } from 'path';

/**
 * Check if a path exists and is accessible
 * @param {string} path - Path to check
 * @returns {boolean} Whether path exists
 */
export function exists(path) {
    try {
        return existsSync(path);
    } catch {
        return false;
    }
}

/**
 * Check if path is a file
 * @param {string} path - Path to check
 * @returns {boolean} Whether path is a file
 */
export function isFile(path) {
    try {
        return statSync(path).isFile();
    } catch {
        return false;
    }
}

/**
 * Check if path is a directory
 * @param {string} path - Path to check
 * @returns {boolean} Whether path is a directory
 */
export function isDirectory(path) {
    try {
        return statSync(path).isDirectory();
    } catch {
        return false;
    }
}

/**
 * Find project root by looking for common project files
 * @param {string} startPath - Starting directory
 * @returns {string|null} Project root path or null if not found
 */
export function findProjectRoot(startPath = process.cwd()) {
    const indicators = ['package.json', 'go.mod', '.git', 'tsconfig.json'];

    let currentPath = startPath;

    while (currentPath !== '/') {
        for (const indicator of indicators) {
            if (exists(join(currentPath, indicator))) {
                return currentPath;
            }
        }

        currentPath = join(currentPath, '..');
    }

    return null;
}

/**
 * Get file extension
 * @param {string} filePath - File path
 * @returns {string} File extension (including dot)
 */
export function getExtension(filePath) {
    const lastDot = filePath.lastIndexOf('.');

    return lastDot > 0 ? filePath.substring(lastDot) : '';
}

/**
 * Check if file is a TypeScript or JavaScript file
 * @param {string} filePath - File path
 * @returns {boolean} Whether file is TS/JS
 */
export function isTSJSFile(filePath) {
    const ext = getExtension(filePath);

    return ['.ts', '.js', '.tsx', '.jsx'].includes(ext);
}
