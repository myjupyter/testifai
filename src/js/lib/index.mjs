import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { loadConfig, initConfig } from './core/config.mjs';
import { scanFiles } from './core/parser.mjs';
import { sendToAPI } from './core/api.mjs';
import { generateTestFiles } from './core/generator.mjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const packagePath = join(__dirname, '..', 'package.json');
const packageJson = JSON.parse(readFileSync(packagePath, 'utf8'));

/**
 * @typedef {Object} ParsedArgs
 * @property {string} command
 * @property {Record<string, string | boolean>} options
 * @property {string[]} args
 */

/**
 * Parse command line arguments
 * @param {string[]} args - Process arguments
 * @returns {ParsedArgs} Parsed command and options
 */
function parseArgs(args = process.argv.slice(2)) {
    const command = args[0] || 'scan';
    /** @type {Record<string, string | boolean>} */
    const options = {};
    /** @type {string[]} */
    const positionalArgs = [];

    for (let i = 1; i < args.length; i++) {
        const arg = args[i];

        if (arg.startsWith('--')) {
            const [key, value] = arg.slice(2).split('=');

            if (value !== undefined) {
                options[key] = value;
            } else {
                options[key] = true;
            }
        } else if (arg.startsWith('-')) {
            const key = arg.slice(1);

            options[key] = true;
        } else {
            positionalArgs.push(arg);
        }
    }

    return {
        command,
        options,
        args: positionalArgs,
    };
}

/**
 * Display help information
 */
function showHelp() {
    console.log(`
testifai v${packageJson.version}
CLI tool for generating tests from comments in TypeScript and JavaScript code

Usage:
  testifai <command> [options] [path]

Commands:
  init                   Initialize testifai.json config file
  scan [path]           Scan files and generate tests (default)
  check [path]          Check files without generating tests
  version               Show version

Options:
  --config <path>       Path to config file (default: testifai.json)
  --dry-run            Show what would be generated without creating files
  --type <type>        Filter by test type (xunit, table, suite)
  --verbose            Show detailed output
  --help               Show this help

Examples:
  testifai init
  testifai scan ./src
  testifai scan --type=xunit ./src/utils.ts
  testifai check --verbose ./src
`);
}

/**
 * Show version information
 */
function showVersion() {
    console.log(`testifai v${packageJson.version}`);
}

/**
 * Main CLI function
 */
export async function main() {
    const { command, options, args } = parseArgs();
    let isVerbose = false;

    try {
        const helpOption = options.help;
        const shortHelpOption = options.h;
        const versionOption = options.version;
        const verboseOption = options.verbose;
        const typeOption = options.type;
        const dryRunOption = options['dry-run'];
        const configOption = options.config;
        const forceOption = options.force;

        const isHelp =
            helpOption === true || helpOption === 'true' || shortHelpOption === true || shortHelpOption === 'true';
        const isVersion = versionOption === true || versionOption === 'true';

        isVerbose = verboseOption === true || verboseOption === 'true';

        const isDryRun = dryRunOption === true || dryRunOption === 'true';
        const initConfigPath = typeof configOption === 'string' ? configOption : undefined;
        const targetType =
            typeof typeOption === 'string' && ['xunit', 'table', 'suite'].includes(typeOption)
                ? /** @type {import('./types.mjs').TestType} */ (typeOption)
                : undefined;
        const forceFlag = forceOption === true || forceOption === 'true';

        if (isHelp) {
            showHelp();

            return;
        }

        if (command === 'version' || isVersion) {
            showVersion();

            return;
        }

        if (command === 'init') {
            /** @type {import('./core/config.mjs').InitOptions} */
            const initOptions = {
                config: initConfigPath,
                force: forceFlag,
            };

            await initConfig(initOptions);

            return;
        }

        // Load config for scan/check commands
        const config = await loadConfig(initConfigPath);
        const targetPath = args[0] || process.cwd();

        if (command === 'check' || command === 'scan') {
            console.log(`🔍 Scanning ${targetPath} for testifai comments...`);

            const scanResults = await scanFiles(targetPath, {
                type: targetType,
                verbose: isVerbose,
            });

            if (scanResults.length === 0) {
                console.log('No testifai comments found.');

                return;
            }

            console.log(`Found ${scanResults.length} testifai comment(s)`);

            if (command === 'check' || isDryRun) {
                for (const result of scanResults) {
                    console.log(
                        `📝 ${result.file}:${result.line} - ${result.testType} test for ${result.functionName}`,
                    );
                }

                return;
            }

            console.log('🚀 Generating tests...');

            const generationTasks = scanResults.map(async (result) => {
                try {
                    if (isVerbose) {
                        console.log(`Processing ${result.file}:${result.line}`);
                    }

                    const response = await sendToAPI(config, result);

                    await generateTestFiles(result, response, { config });

                    console.log(`✅ Generated test for ${result.functionName} in ${result.file}`);
                } catch (error) {
                    const err = error instanceof Error ? error : new Error(String(error));

                    console.error(`❌ ${err.message}`);
                }
            });

            await Promise.all(generationTasks);

            console.log('🎉 Test generation completed!');
        } else {
            console.error(`Unknown command: ${command}`);
            console.error('Run "testifai --help" for usage information.');
            process.exit(1);
        }
    } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        if (isVerbose) {
            console.error('Stack trace:', err.stack);
        } else {
            console.error('Error:', err.message);
        }

        process.exit(1);
    }
}
