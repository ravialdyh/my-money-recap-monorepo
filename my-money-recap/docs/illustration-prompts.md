# My Money Recap · illustration prompts (batch 1: GPT Image 2; batch 2: Nano Banana Pro with the batch 1 renders as references)

_All paths below are relative to the `my-money-recap/` folder root, which is also where the `python3 scripts/...` commands are run from._

Use the STYLE BLOCK verbatim at the top of every prompt, then append one SUBJECT block.
Generate one image per subject. Square (1:1) or 4:3 works best; the subject is centred with generous margins so it can be placed over a UI card.

---

## STYLE BLOCK (paste first, every time)

Flat vector illustration in a playful, modern kids-app style. Chunky, rounded, simplified shapes with clean geometric construction. No outlines or line art: forms are defined purely by flat colour fills, with one slightly darker flat tone for shading on the lower third of each shape and small flat white highlight dots. Absolutely no gradients, no textures, no 3D rendering, no drop shadows, no photorealism. Colours are bold, saturated and cheerful, limited to this palette: violet #5B4FE8, sunny yellow #F7C948, gold #E8A317, coral red #F0665A, leaf green #3DAA6A, bubblegum pink #FF8FAB with darker pink #F06A8C for shading, deep navy ink #1E1B3A for eyes and small details, cream #FFF8EC and pale sky blue #C9E8FF for whites and glass. Composition: a single centred hero subject filling about 70% of the frame, generous empty margins on all sides, a plain solid single-colour background with no scenery, no floor line, no shadow on the ground. Crisp edges, sticker-like, suitable for a mobile app UI. No text or lettering anywhere unless specified.

---

## SUBJECT 1 · Cover screen: piggy bank with raining coins

Background colour: solid violet #5B4FE8.
Subject: a plump, friendly bubblegum-pink piggy bank seen from a three-quarter side view, facing right. Round barrel body, four short stubby legs, a small curled tail on the left, one pointed ear, a big round pink snout with two navy nostrils, one large navy eye with a white glint, a small darker-pink cheek blush. A navy coin slot on top of its back with a big gold coin half-inserted, standing upright. Eight to ten gold coins of different sizes rain down around it from above, some seen face-on as circles with a thin darker-gold inner ring, some seen edge-on as thin ovals. The piggy has a slightly darker pink flat shading tone across its lower belly and legs. Mood: cheerful, lucky, abundant.

---

## SUBJECT 2 · Earning screen: lemonade stand

Background colour: solid sunny yellow #F7C948.
Subject: a small kid-built lemonade stand seen straight-on. A coral-red-and-white vertically striped awning across the top with a scalloped bottom edge, held up by two thin wooden posts in warm brown. Below it, a coral red counter box with a white top edge. On the counter, from left to right: a tall glass pitcher of pink lemonade with a white handle and a lemon slice on the rim, two small pink paper cups, and a glass jar full of gold coins with a violet lid. A single white coin-shaped sun or gold coin floats top-right. The sign on the front of the counter reads exactly "LEMONADE 50p" in chunky rounded white capital letters, and this is the only text in the image. Because the background is yellow, use pink lemonade and cream or white glass so the drink reads clearly. Mood: sunny, entrepreneurial, homemade.

---

## SUBJECT 3 · Spending screen: vending machine

Background colour: solid coral red #F0665A.
Subject: a tall, slightly cartoon-proportioned vending machine seen straight-on, standing upright. Body in deep muted violet #3D3A6B with a bright violet #5B4FE8 rounded top band. A large pale sky-blue glass window on the left two thirds showing three shelves, each holding three simplified snack items in different flat colours: yellow crisp packets, pink chocolate bars, green and violet drink cans, coral candy boxes. Blocky, no labels. On the right a narrow violet control panel with a small navy display screen, six white round keypad buttons in two columns, a navy coin slot, and one gold coin about to drop into the slot. Near the bottom a wide navy dispensing flap with one coral snack packet tumbling out at a slight angle. Two short navy feet. Mood: tempting, a little mischievous, the place pocket money disappears.

---

## SUBJECT 4 · Money personality screen: "The Squirrel" (hero illustration, boldest of the set)

Background colour: solid leaf green #3DAA6A.
Subject: a cute rust-orange squirrel sitting proudly on top of a plump bubblegum-pink piggy bank, hugging one oversized gold coin to its chest with both paws. The squirrel has a round head, a small navy eye with a white glint, a dark brown button nose, a tiny smile, two upright ears with lighter peach inner ears, a cream oval belly, and a huge fluffy tail curling up and over its head in a big S-curve, the tail drawn as one thick rounded rust shape with a lighter peach stripe running through its centre. The piggy bank matches Subject 1: pink, darker-pink shading on the lower belly, navy coin slot on its back to the left of the squirrel, pink snout facing right, one navy eye, stubby legs. Scattered around the base on both sides: a few gold coins (face-on and edge-on) and two brown acorns with darker caps. Two small white four-point sparkles float in the upper corners. Mood: triumphant, cosy, proud saver. This is a shareable "personality badge" image, so it should feel like a mascot poster.

---

## Tips for consistency across the four

- Generate Subject 1 first. If the tool supports image references, feed it in for Subjects 2 to 4 with "match the exact illustration style of the reference".
- If a result has outlines or gradients, add: "Remove all outlines and gradients; every shape is a flat fill."
- For drop-in use on the UI cards, try appending "transparent background, PNG" instead of the solid colour if your tool supports transparency.
- Keep each subject to one image per prompt. Combining them dilutes the style.

---
---

# Batch 2 · onboarding, avatars, archetypes (added 2026-09-04, from docs/PRD.md §13)

Same rules as batch 1: paste the STYLE BLOCK first, then one SUBJECT block. Tool: Nano Banana Pro (cheaper than GPT Image 2). Attach all four `renders/*.png` files as reference images to every prompt and add "match the exact illustration style, shading and colour of the reference images". Request 2K output. Generate Subject 5 first as a style test; if it matches the cover piggy, continue.

Cost-saving exception to "one image per prompt": generate the six avatars (Subjects 6 to 11) as one sheet, "a 3 × 2 grid of six circular badges, evenly spaced, on a plain cream background", then crop. Siblings generated together match each other better. Do the same for the three archetype mascots (Subjects 12 to 14) as a row of three if the avatar sheet holds the style.

Two extra rules for this batch:

- **Avatars (Subjects 6 to 11) are circular badges.** They are shown at 68 px and 56 px, so the animal is a head-and-shoulders portrait filling about 85% of a solid coloured circle, facing the viewer, with a bold silhouette readable at thumbnail size. No props, no coins, no body below the shoulders. Generate on the circle's own colour so the circle can be cut out cleanly; the final asset is the circle itself with a transparent outside.
- **Archetype characters (Subjects 12 to 14) are full poster mascots** in the same register as Subject 4 (the Squirrel): whole body, one prop that tells the story, a few coins, two white four-point sparkles in the upper corners, on the green ground. They are shareable badges, so they should feel like a set with the Squirrel.

---

## SUBJECT 5 · Check-your-inbox screen: magic-link envelope

Background colour: solid violet #5B4FE8.
Subject: a cream #FFF8EC envelope seen straight-on and slightly tilted, flap open at the top. Out of the envelope rises a folded violet #5B4FE8 letter with a single large gold #E8A317 coin printed on it like a seal. The envelope flap has a darker cream flat shading tone on its underside. Around the envelope float three small white four-point sparkles and two tiny gold coins, suggesting something magic just arrived. No text, no address lines, no stamp. Mood: a good surprise in the post.

---

## SUBJECT 6 · Avatar: Fox

Circle colour: solid violet #5B4FE8.
Subject: a friendly fox portrait, head and shoulders, facing the viewer. Coral red #F0665A head with a cream #FFF8EC muzzle and cream inner ears, two tall pointed ears, a navy #1E1B3A button nose, two navy dot eyes each with a white glint, a small closed smile, a slightly darker coral flat shading tone across the lower jaw and shoulders. Mood: clever, cheerful.

---

## SUBJECT 7 · Avatar: Otter

Circle colour: solid sunny yellow #F7C948.
Subject: an otter portrait, head and shoulders, facing the viewer. Warm brown #8C5A3C head with a large cream #FFF8EC muzzle patch, small round ears, a navy button nose, two navy dot eyes with white glints, tiny whisker dots, a darker brown flat shading tone on the lower jaw. Mood: relaxed, playful.

---

## SUBJECT 8 · Avatar: Panda

Circle colour: solid coral red #F0665A.
Subject: a panda portrait, head and shoulders, facing the viewer. Cream #FFF8EC round head with navy #1E1B3A ears and navy oval eye patches, two white dot eyes with navy pupils inside the patches, a navy button nose, a small smile, a pale grey flat shading tone under the chin. Mood: calm, kind.

---

## SUBJECT 9 · Avatar: Frog

Circle colour: solid leaf green #3DAA6A.
Subject: a frog portrait, head and shoulders, facing the viewer. Bright green #6BD08A head with two big round eyes on top, each a cream circle with a navy pupil and white glint, a wide happy mouth drawn as a single navy curve, two small cheek blushes in pink #FF8FAB, a darker green flat shading tone across the lower face. Mood: excited.

---

## SUBJECT 10 · Avatar: Owl

Circle colour: solid pale sky blue #C9E8FF.
Subject: an owl portrait, head and shoulders, facing the viewer. Violet #5B4FE8 body with a cream #FFF8EC heart-shaped face disc, two large round navy eyes with white glints, a small gold #E8A317 triangular beak, two short ear tufts, a darker violet flat shading tone across the lower body. Mood: wise, a little serious.

---

## SUBJECT 11 · Avatar: Rabbit

Circle colour: solid gold #E8A317.
Subject: a rabbit portrait, head and shoulders, facing the viewer. Cream #FFF8EC head with two tall upright ears with pink #FF8FAB inner ears, two navy dot eyes with white glints, a small pink nose, two front teeth as a tiny white rectangle, a pale grey flat shading tone under the chin. Mood: bright, eager.

---

## SUBJECT 12 · Money personality: "The Hustler"

Background colour: solid leaf green #3DAA6A.
Subject: a busy honeybee character standing upright on two legs, body striped in sunny yellow #F7C948 and navy #1E1B3A, two pale sky-blue #C9E8FF translucent wings, two short antennae, one navy eye with a white glint, a determined little grin. It holds a gold coin overhead in one hand like a trophy and carries a small cream #FFF8EC sack overflowing with gold coins in the other. A few gold coins scattered at its feet. Two white four-point sparkles in the upper corners. Mood: energetic, on the move, hit the number while living a little.

---

## SUBJECT 13 · Money personality: "The Treat-Yourself"

Background colour: solid leaf green #3DAA6A.
Subject: a plump, content cat sitting on its haunches, bubblegum pink #FF8FAB fur with darker pink #F06A8C shading on the lower body, cream #FFF8EC belly, two pointed ears, closed happy eyes drawn as two navy curves, a tiny navy nose. It hugs an oversized ice-cream cone with three scoops in coral #F0665A, cream and violet #5B4FE8, with a gold coin pressed into the top scoop like a cherry. Two or three gold coins lie tipped over at its feet. Two white four-point sparkles in the upper corners. Mood: blissful, unbothered, one big treat tipped the month.

---

## SUBJECT 14 · Money personality: "The Slow Burner"

Background colour: solid leaf green #3DAA6A.
Subject: a smiling tortoise walking slowly to the right, seen from the side. Leaf-green #6BD08A skin with a darker green flat shading tone on the underside, a rounded violet #5B4FE8 shell with a few large gold #E8A317 hexagon plates, one navy eye with a white glint. Behind it trails a neat dotted line of small gold coins getting smaller into the distance, showing lots of little spends adding up. Two white four-point sparkles in the upper corners. Mood: steady, unhurried, sneaky little spends.

---

## SUBJECT 15 · Recap screen 3 hero: saving-target jar

Background colour: solid coral red #F0665A.
Subject: a tall glass jar with a violet #5B4FE8 lid, seen straight-on, drawn in pale sky blue #C9E8FF with a cream #FFF8EC vertical highlight stripe on its left side. The jar is about two-thirds full of gold coins stacked in loose layers. A small bubblegum-pink piggy bank (matching Subject 1) sits beside the jar on the right, looking up at it. One gold coin is mid-air above the jar, about to drop in. Nothing marks a measurement or percentage. Mood: nearly there, encouraging.

---

## SUBJECT 16 · App icon (optional, for the iOS build, not the artboards)

Background colour: solid violet #5B4FE8.
Subject: the piggy bank from Subject 1 as a front-facing head-and-snout portrait only, filling about 80% of the frame, one gold coin half-inserted in the slot on top, two navy eyes with white glints, big pink snout with two navy nostrils. Centred, symmetrical, no raining coins. Suitable as a 1024 × 1024 app icon; iOS applies the rounded mask.

---

## Batch 2 status (2026-09-04) — complete

All of Subjects 5 to 16 are generated, cut and wired in. Nothing in this file is outstanding.

| Subject | Render | Asset |
|---|---|---|
| 5 envelope | `renders/5-envelope.png` | `ill-envelope.webp` |
| 6 to 11 avatars | `renders/6-avatars-sheet.png` (one 3 × 2 sheet at 4K) | `ill-avatar-{fox,otter,panda,frog,owl,rabbit}.webp` |
| 12 Hustler | `renders/12-hustler-bee.png` | `ill-hustler.webp` |
| 13 Treat-Yourself | `renders/13-treat-cat.png` | `ill-treat.webp` |
| 14 Slow Burner | `renders/14-slowburner-tortoise.png` | `ill-slowburner.webp` |
| 15 saving-target jar | `renders/15-jar.png` | `ill-jar.webp` |
| 16 app icon | `renders/16-app-icon.png` | `assets/icons/app-icon-1024.png` (opaque, for the iOS build) |

### What to know before generating more

- **The Higgsfield CLI cannot be used on a trial plan** (`only_mcp_usage_on_trial_is_available`). Go through the MCP tools. Model `nano_banana_pro`, reference role `image_references`, up to 14 references. Cost is flat per image regardless of reference count: 2 credits at 1K or 2K, 4 at 4K.
- **References leak content, not just style.** This cost four attempts on Subject 14. Attaching `4-personality-squirrel.png` produced a squirrel riding the tortoise plus acorns, because that reference *is* a squirrel on a piggy bank with acorns. Swapping in `3-spending-vending-machine.png` put a ghosted vending machine under it; `1-cover-piggy-bank.png` put ghosted piggy legs under it. What worked was a single reference showing one clean character on a flat ground (`13-treat-cat.png`), plus an explicit "do not draw the cat from the reference image" and a list of the specific intruders to exclude.
- **Ask for the flat ground explicitly.** Left to itself the model adds a soft contact shadow under the character, which the style block forbids and which cuts out as a coloured blob. Spell out: no ground shadow, no ellipse under the feet, no floor, no horizon, no band of another colour along the bottom. `cut_flat_bg.py` also removes such a shadow if one survives, but a clean render is better.
- **Subject sheets work when the subjects are siblings.** The six avatars came out consistent in one job. The three archetype mascots each need their own frame, because the "two sparkles in the upper corners" framing is per-image.

## After generating: getting the files into the design system and Claude Design

1. Save each render as `renders/<n>-<name>.png` next to the first four.
2. Cut the background with the scripts in `scripts/`, which write the transparent WebP into `assets/illustrations/` as `ill-<key>.webp`. Keys: `envelope`, `avatar-fox`, `avatar-otter`, `avatar-panda`, `avatar-frog`, `avatar-owl`, `avatar-rabbit`, `hustler`, `treat`, `slowburner`, `jar`.
   - `python3 scripts/cut_flat_bg.py renders/<n>-<name>.png <key> <width>` for a hero on a flat ground. A plain colour key is not safe (the violet letter sits on the violet ground), so it keys only the background region connected to the frame border, erodes 2px, and re-feathers. Sizes to roughly 2× display width.
   - `python3 scripts/cut_avatars.py` for the avatar sheet: finds the six circles geometrically and masks each to its inscribed circle at 204px (3× the 68px grid).
   - Both need `Pillow` and `numpy` (`python3 -m pip install --user Pillow numpy`).
3. In `../my-money-recap-ds/scripts/build-assets.mjs`, add the new keys to the `names` map, then run `npm run build:assets` so `src/assets/illustrations.ts` gains them. Hyphenated keys are quoted automatically; read them as `ILLUSTRATIONS['avatar-fox']`.
4. Re-sync to Claude Design with `/design-sync` from the design-system folder. The `Illustration` card then offers the new keys.
5. In the Claude Design canvas, replace each "ILLUSTRATION TBD" placeholder with `Illustration` using the matching key. The avatar TBD circles become the six avatar assets at 68 px (A4) and 56 px (A8).
