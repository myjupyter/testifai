import { readFileSync, writeFileSync, existsSync } from 'fs';
import { join } from 'path';

/**
 * @typedef {import('../types.mjs').TestifaiConfig} TestifaiConfig
 * @typedef {import('../types.mjs').InitOptions} InitOptions
 * @typedef {import('../types.mjs').ProviderConfig} ProviderConfig
 * @typedef {import('../types.mjs').OpenAIConfig} OpenAIConfig
 * @typedef {import('../types.mjs').OutputConfig} OutputConfig
 * @typedef {import('../types.mjs').ScanConfig} ScanConfig
 */

/**
 * Default configuration
 * @type {TestifaiConfig}
 */
const DEFAULT_CONFIG = {
    provider: {
        openai: {
            apiKey: '<YOUR_OPENAI_API_KEY>',
        },
    },
    output: {
        testDir: null, // null means next to source file (handled as __tests__ by default in generator)
        extension: '.spec',
    },
    scan: {
        include: ['**/*.ts', '**/*.js'],
        exclude: ['node_modules/**', '**/*.spec.*', '**/*.test.*'],
    },
};

/**
 * Template for user-facing configuration file (excludes internal fields)
 */
const USER_CONFIG_TEMPLATE = {
    provider: {
        openai: {
            apiKey: DEFAULT_CONFIG.provider.openai.apiKey,
        },
    },
    output: {
        ...DEFAULT_CONFIG.output,
    },
    scan: {
        ...DEFAULT_CONFIG.scan,
    },
};

/**
 * Get config file path
 * @param {string|undefined} configPath - Custom config path
 * @returns {string} Path to config file
 */
function getConfigPath(configPath) {
    if (configPath) {
        return configPath;
    }

    // Try to detect project type and use appropriate config
    const cwd = process.cwd();

    // Check for package.json (JS/TS project)
    if (existsSync(join(cwd, 'package.json'))) {
        return join(cwd, 'testifai.json');
    }

    // Default to JSON
    return join(cwd, 'testifai.json');
}

/**
 * Load configuration from file
 * @param {string|undefined} configPath - Path to config file
 * @returns {Promise<TestifaiConfig>} Configuration object
 */
export async function loadConfig(configPath) {
    const path = getConfigPath(configPath);

    if (!existsSync(path)) {
        throw new Error(`Config file not found: ${path}. Run 'testifai init' to create one.`);
    }

    try {
        const content = readFileSync(path, 'utf8');

        if (path.endsWith('.json')) {
            return /** @type {TestifaiConfig} */ (JSON.parse(content));
        } else if (path.endsWith('.yaml') || path.endsWith('.yml')) {
            // For YAML support, we'd need a YAML parser
            // For now, just handle JSON
            throw new Error('YAML config not supported in JS CLI. Use testifai.json instead.');
        }

        throw new Error(`Unsupported config file format: ${path}`);
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        if (err.name === 'SyntaxError') {
            throw new Error(`Invalid JSON in config file: ${path}`);
        }

        throw err;
    }
}

/**
 * Initialize configuration file
 * @param {InitOptions} [options={}] - Options for initialization
 */
export async function initConfig(options = {}) {
    /** @type {InitOptions} */
    const initOptions = options;
    const configPath = getConfigPath(initOptions.config);

    if (existsSync(configPath) && !initOptions.force) {
        console.log(`Config file already exists: ${configPath}`);
        console.log('Use --force to overwrite');

        return;
    }

    try {
        const configContent = JSON.stringify(USER_CONFIG_TEMPLATE, null, 2);

        writeFileSync(configPath, configContent, 'utf8');

        console.log(`✅ Created config file: ${configPath}`);
        console.log('');
        console.log('Please update the configuration with your API key:');
        console.log('- Set provider.openai.apiKey to your OpenAI API key');
        console.log('- Adjust output or scan settings if you need custom paths or patterns');
        console.log('');
        console.log('Example:');
        console.log(
            JSON.stringify(
                {
                    provider: {
                        openai: {
                            apiKey: 'your-actual-api-key-here',
                        },
                    },
                },
                null,
                2,
            ),
        );
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        throw new Error(`Failed to create config file: ${err.message}`);
    }
}

/**
 * Validate configuration
 * @param {TestifaiConfig} config - Configuration to validate
 * @throws {Error} If configuration is invalid
 */
export function validateConfig(config) {
    if (!config) {
        throw new Error('Configuration is required');
    }

    if (!config.provider || !config.provider.openai || !config.provider.openai.apiKey) {
        throw new Error('OpenAI API key is required. Please set provider.openai.apiKey in your config.');
    }

    const { apiKey } = config.provider.openai;

    if (apiKey === '<YOUR_OPENAI_API_KEY>' || apiKey.length === 0) {
        throw new Error('Please set a valid OpenAI API key in your config file.');
    }
}

/**
 * Get effective configuration with defaults
 * @param {Partial<TestifaiConfig>} [config] - Base configuration
 * @returns {TestifaiConfig} Configuration with defaults applied
 */
export function getEffectiveConfig(config) {
    /** @type {Partial<TestifaiConfig>} */
    const baseConfig = config ?? {};

    const {
        provider: providerConfig = /** @type {Partial<ProviderConfig>} */ ({}),
        output: outputConfig = /** @type {Partial<OutputConfig>} */ ({}),
        scan: scanConfig = /** @type {Partial<ScanConfig>} */ ({}),
        ...additionalConfig
    } = baseConfig;

    const openaiConfig = /** @type {Partial<OpenAIConfig>} */ (providerConfig.openai ?? {});

    return {
        ...DEFAULT_CONFIG,
        ...additionalConfig,
        provider: {
            ...DEFAULT_CONFIG.provider,
            ...providerConfig,
            openai: {
                ...DEFAULT_CONFIG.provider.openai,
                ...openaiConfig,
            },
        },
        output: {
            ...DEFAULT_CONFIG.output,
            ...outputConfig,
        },
        scan: {
            ...DEFAULT_CONFIG.scan,
            ...scanConfig,
        },
    };
}
