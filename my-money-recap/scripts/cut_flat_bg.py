"""Cuts a hero illustration off its flat single-colour ground into a transparent WebP.

A plain colour key is not safe here: several subjects use the ground hue inside the
artwork too (the violet letter on the violet ground, for example). So the key is
restricted to the background region actually connected to the frame border, and the
few anti-aliased pixels on that boundary get a soft alpha ramp instead of a hard cut.

Resizing happens in premultiplied space, otherwise PIL blends the ground colour of
fully transparent pixels back into the edge and leaves a coloured fringe.

    python3 scripts/cut_flat_bg.py renders/5-envelope.png envelope 640
"""
from PIL import Image, ImageDraw, ImageFilter
import numpy as np
import sys, os, json

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
ILL = os.path.join(ROOT, 'assets', 'illustrations')
T1 = 12.0   # colour distance under which a pixel counts as candidate ground
SHADOW_RESIDUAL, SHADOW_K_MIN = 12.0, 0.70   # how far off the ground ray a shadow may sit

src, key, out_w = sys.argv[1], sys.argv[2], int(sys.argv[3])
im = Image.open(os.path.join(ROOT, src)).convert('RGB')
rgb = np.asarray(im).astype(np.float32)
bg = rgb[4, 4].copy()
dist = np.sqrt(((rgb - bg) ** 2).sum(axis=2))

# The renderer sometimes drops a soft contact shadow on the ground despite the
# prompt. A shadow is the ground colour scaled down, so test each pixel against
# the ray k * bg: small residual and k just under 1 means ground-in-shadow, while
# real artwork of any other hue lands far off that ray and is kept.
k = (rgb @ bg) / float(bg @ bg)
residual = np.sqrt(((rgb - k[..., None] * bg) ** 2).sum(axis=2))
# k must stay below 1: a shadow only ever darkens the ground. Artwork that happens
# to be a lighter tint of it (the violet letter on the violet ground) has k just above
# 1 and must be kept.
shadow = (residual < SHADOW_RESIDUAL) & (k > SHADOW_K_MIN) & (k < 0.99)

# Background candidates, then keep only the region touching the frame border.
# .copy() detaches from the numpy buffer, otherwise floodfill's in-place
# writes are not visible when the array is read back.
# A median pass drops the speckle the renderer leaves in flat areas, which would
# otherwise survive as isolated opaque pixels and blow up the crop box.
cand = Image.fromarray(np.where((dist < T1) | shadow, 255, 0).astype(np.uint8)).filter(ImageFilter.MedianFilter(5)).copy()
ImageDraw.floodfill(cand, (0, 0), 128, thresh=0)
conn = np.array(cand) == 128

# The ground drifts by a few units across the frame, so a colour-distance ramp
# mis-scores its edges. Take the cut from the connected region instead, erode 1px
# to drop pixels that are part ground (2px, enough to clear the ground-blended rim
# so the art also sits cleanly on a white card), and re-feather to restore the anti-aliasing.
mask = Image.fromarray(np.where(conn, 0, 255).astype(np.uint8))
mask = mask.filter(ImageFilter.MinFilter(5)).filter(ImageFilter.GaussianBlur(1.0))
alpha = (np.asarray(mask).astype(np.float32) / 255.0)
bbox = mask.getbbox()

# Premultiply, resize, un-premultiply.
a = alpha[bbox[1]:bbox[3], bbox[0]:bbox[2], None]
c = rgb[bbox[1]:bbox[3], bbox[0]:bbox[2]]
pm = np.concatenate([c * a, a * 255], axis=2).astype(np.uint8)
h, w = pm.shape[:2]
out_h = max(1, round(h * out_w / w))
small = np.asarray(Image.fromarray(pm).resize((out_w, out_h), Image.LANCZOS)).astype(np.float32)
sa = np.clip(small[:, :, 3:4] / 255.0, 1e-4, 1.0)
out = np.concatenate([np.clip(small[:, :, :3] / sa, 0, 255), small[:, :, 3:4]], axis=2).astype(np.uint8)

dst = os.path.join(ILL, f'ill-{key}.webp')
Image.fromarray(out).save(dst, 'WEBP', quality=92, method=6, exact=True)
kb = os.path.getsize(dst) / 1024
print(json.dumps({key: {'file': f'ill-{key}.webp', 'bg': '#%02X%02X%02X' % tuple(int(v) for v in bg),
                        'w': out_w, 'h': out_h, 'kb': round(kb, 1), 'bbox': list(bbox)}}, indent=1))
