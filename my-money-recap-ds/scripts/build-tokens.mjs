// Generates src/tokens/tokens.css from src/tokens/tokens.json.
// Aliases like "{palette.violet.500}" become var(--mmr-palette-violet-500).
import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { here, tokens, walk } from './lib.mjs';

const vars = [];

const aliasToVar = (v) => v.replace(/\{([^}]+)\}/g, (_, p) => `var(--mmr-${p.replace(/\./g, '-')})`);
const flatRaw = {};
walk(tokens, (p, v) => {
  flatRaw[p.join('.')] = v.$value;
});

for (const [p, raw] of Object.entries(flatRaw)) {
  const name = `--mmr-${p.replace(/\./g, '-')}`;
  let css;
  if (typeof raw === 'string') css = aliasToVar(raw);
  else if (typeof raw === 'number') css = String(raw);
  else if (Array.isArray(raw))
    css = p.startsWith('font.family')
      ? raw.map((f) => (f.includes(' ') ? `'${f}'` : f)).join(', ')
      : `cubic-bezier(${raw.join(', ')})`;
  else if (raw && typeof raw === 'object') {
    if (p.startsWith('type.')) {
      for (const [k2, v2] of Object.entries(raw)) {
        const part = k2.replace(/[A-Z]/g, (m) => `-${m.toLowerCase()}`);
        vars.push([`${name}-${part}`, typeof v2 === 'string' ? aliasToVar(v2) : String(v2)]);
      }
      continue;
    }
    if ('color' in raw && 'spread' in raw)
      css = `${raw.offsetX} ${raw.offsetY} ${raw.blur} ${raw.spread} ${aliasToVar(raw.color)}`;
    else if ('color' in raw && 'width' in raw) css = `${raw.width} ${raw.style} ${aliasToVar(raw.color)}`;
    else continue;
  } else continue;
  vars.push([name, css]);
}

const css = `/* Generated from tokens.json by scripts/build-tokens.mjs. Do not edit by hand. */\n:root {\n${vars.map(([n, v]) => `  ${n}: ${v};`).join('\n')}\n}\n`;
writeFileSync(resolve(here, '../src/tokens/tokens.css'), css);

console.log(`tokens.css: ${vars.length} custom properties`);
