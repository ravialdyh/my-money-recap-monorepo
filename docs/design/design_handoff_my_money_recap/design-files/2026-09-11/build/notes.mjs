// Regenerates the 2026-09-11 sections of ../../../DESIGN_NOTES_VERBATIM.md from the built canvases, so the
// transcript can never drift from the stickies. Run after build.mjs: `node notes.mjs`.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const files = [['Main', 'IDR proofs'], ['MonthSoFar', 'Month so far and Partial Recap'], ['AuthAndOnboarding', 'Auth and Onboarding'], ['HomeAndDailyLoop', 'Home and Core Features'], ['RecapAndShare', 'Recap and Share'], ['SettingsAndAccount', 'Settings and Account'], ['CrossCutting', 'iPad']];
const strip = (h) => h.replace(/<[^>]+>/g, '').replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&amp;/g, '&').replace(/\s+/g, ' ').trim();
let out = '';
for (const [stem, name] of files) {
  const html = fs.readFileSync(path.join(here, 'out', `${stem}.dc.html`), 'utf8');
  out += `\n## 2026-09-11/${name}.dc.html\n`;
  const re = /<div data-screen-label="([^"]+)"/g; let m;
  const starts = []; while ((m = re.exec(html))) starts.push([m.index, m[1]]);
  starts.forEach(([start, label], i) => {
    const block = html.slice(start, starts[i + 1]?.[0] ?? html.length);
    for (const n of block.matchAll(/<div style="background:var\(--mmr-palette-yellow-100\)[^>]*>([\s\S]*?)<\/div>\n?<\/div><\/sc-if>|<div style="background:var\(--mmr-palette-yellow-100\)[^>]*>([\s\S]*?)<\/div>/g)) {
      const text = strip(n[1] ?? n[2] ?? '');
      if (text) out += `- [${label}] ${text}\n`;
    }
  });
}
const notesPath = path.join(here, '..', '..', '..', 'DESIGN_NOTES_VERBATIM.md');
let doc = fs.readFileSync(notesPath, 'utf8');
const marker = '\n## 2026-09-11/';
const cut = doc.indexOf(marker);
doc = (cut === -1 ? doc.trimEnd() : doc.slice(0, cut).trimEnd()) + '\n' + out;
fs.writeFileSync(notesPath, doc);
console.log('transcript updated:', (out.match(/^- \[/gm) || []).length, 'notes');
