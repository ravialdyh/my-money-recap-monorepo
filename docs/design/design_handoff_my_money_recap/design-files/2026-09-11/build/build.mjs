import fs from 'node:fs';
import path from 'node:path';
import { DS, extraCss, a11yCss } from './lib.mjs';
const css = fs.readFileSync(path.join(DS, '_ds_bundle.css'), 'utf8');
const js = fs.readFileSync(path.join(DS, '_ds_bundle.js'), 'utf8');
const names = process.argv.slice(2);
for (const name of names) {
  const mod = await import(`./${name}.mjs?t=${Date.now()}`);
  const { body, props, logic } = mod.default;
  const html = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<script src="./support.js"></script>
</head>
<body>
<x-dc>
<helmet>
<style>
${css}
${extraCss}
${a11yCss}
</style>
<script>
${js}
</script>
</helmet>
${body}
</x-dc>
<script data-dc-script data-props='${JSON.stringify(props).replace(/&/g, '&amp;').replace(/'/g, '&#39;')}'>
class Component extends DCLogic {
${logic}
}
</script>
</body>
</html>
`;
  fs.mkdirSync('out', { recursive: true });
  fs.writeFileSync(path.join('out', `${name}.dc.html`), html);
  console.log('built', name, html.length);
}
