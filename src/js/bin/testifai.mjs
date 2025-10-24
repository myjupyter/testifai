#!/usr/bin/env node
import { main } from '../lib/index.mjs';

main().catch((e) => {
    console.error('[testifai] Error:', e?.stack || e?.message || e);
    process.exit(1);
});
