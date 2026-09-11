"""Cuts the 3x2 avatar sheet into six transparent circular WebPs.

The sheet is generated on a plain cream ground with one solid-coloured circle per
avatar, so the cut is geometric: find each circle's bounding box, mask to the
inscribed circle, inset a few px to drop the anti-aliased cream fringe.
"""
from PIL import Image, ImageDraw
import os, json

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
ILL = os.path.join(ROOT, 'assets', 'illustrations')
SHEET = os.path.join(ROOT, 'renders', '6-avatars-sheet.png')
OUT_PX = 204          # 3x the 68px grid display size
INSET = 4             # px of full-res radius dropped to kill the cream fringe

NAMES = [['fox', 'otter', 'panda'], ['frog', 'owl', 'rabbit']]

im = Image.open(SHEET).convert('RGB')
W, H = im.size
px = im.load()
bg = px[5, 5]

def isbg(c, tol=18):
    return all(abs(c[i] - bg[i]) <= tol for i in range(3))

def bands(vals, gap=12):
    out, s, p = [], vals[0], vals[0]
    for v in vals[1:]:
        if v - p > gap:
            out.append((s, p)); s = v
        p = v
    out.append((s, p))
    return out

rows = bands([y for y in range(0, H, 2) if any(not isbg(px[x, y]) for x in range(0, W, 4))])
cols = bands([x for x in range(0, W, 2) if any(not isbg(px[x, y]) for y in range(0, H, 4))])
assert len(rows) == 2 and len(cols) == 3, (rows, cols)

meta = {}
for r, (y0, y1) in enumerate(rows):
    for c, (x0, x1) in enumerate(cols):
        name = NAMES[r][c]
        d = min(x1 - x0, y1 - y0) + 1
        cx, cy = (x0 + x1) // 2, (y0 + y1) // 2
        box = (cx - d // 2, cy - d // 2, cx + d // 2, cy + d // 2)
        cell = im.crop(box).convert('RGBA')
        mask = Image.new('L', cell.size, 0)
        ImageDraw.Draw(mask).ellipse((INSET, INSET, cell.size[0] - 1 - INSET, cell.size[1] - 1 - INSET), fill=255)
        cell.putalpha(mask)
        cell.resize((OUT_PX, OUT_PX), Image.LANCZOS).save(os.path.join(ILL, f'ill-avatar-{name}.webp'), 'WEBP', quality=92, method=6)
        im.crop(box).save(os.path.join(ROOT, 'renders', f'6-avatar-{name}.png'))
        kb = os.path.getsize(os.path.join(ILL, f'ill-avatar-{name}.webp')) / 1024
        meta[f'avatar-{name}'] = {'file': f'ill-avatar-{name}.webp', 'bg': '#%02X%02X%02X' % px[cx, y0 + 6],
                                  'w': OUT_PX, 'h': OUT_PX, 'kb': round(kb, 1), 'source_d': d}
        print(f'{name:7s} d={d} circle_bg={meta["avatar-" + name]["bg"]} {kb:.1f} kB')

print(json.dumps(meta, indent=1))
