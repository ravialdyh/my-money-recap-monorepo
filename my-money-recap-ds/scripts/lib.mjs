// What the build scripts share: the scripts folder, the parsed tokens.json, and the token-tree walk.
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const here = dirname(fileURLToPath(import.meta.url));
export const tokens = JSON.parse(readFileSync(resolve(here, '../src/tokens/tokens.json'), 'utf8'));
export const isToken = (o) => o && typeof o === 'object' && '$value' in o;

/** Calls visit(path, token) for every leaf under node; keys starting with $ are metadata, not tokens. */
export function walk(node, visit, path = []) {
  for (const [k, v] of Object.entries(node)) {
    if (k.startsWith('$')) continue;
    const p = [...path, k];
    if (isToken(v)) visit(p, v);
    else if (v && typeof v === 'object') walk(v, visit, p);
  }
}
