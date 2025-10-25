# Testifai TypeScript/JavaScript CLI

CLI tool for automatically generating tests from comments in TypeScript and JavaScript code.

## Installation

```bash
cd src/js
npm install
```

## Usage

### Initialize configuration

```bash
node bin/testifai.mjs init
```

This creates a `testifai.json` config file in your project root.

### Scan and generate tests

```bash
# Scan current directory
node bin/testifai.mjs scan

# Scan specific file or directory
node bin/testifai.mjs scan ./src/utils.ts

# Check without generating (dry run)
node bin/testifai.mjs check ./src

# Verbose output
node bin/testifai.mjs scan --verbose ./src
```

## Comment Format

Add comments before functions to generate tests:

```typescript
// testifai: -type=xunit
function calculateTotal(items: Item[]): number {
  return items.reduce((sum, item) => sum + item.price, 0);
}
```

Supported test types:

- `xunit` - XUnit style tests
- `table` - Table-driven tests
- `suite` - Test suite style

## Configuration

Edit `testifai.json` to configure:

```json
{
  "provider": {
    "openai": {
      "apiKey": "your-api-key-here"
    }
  },
  "output": {
    "testDir": null,
    "extension": ".spec"
  }
}
```

## Requirements

- Node.js >= 18
- Running testifai backend server
- OpenAI API key

The CLI connects to the bundled backend at `http://localhost:6667` automatically, so no additional API configuration is required. Generated tests are written to a `__tests__` folder next to each source file by default.
