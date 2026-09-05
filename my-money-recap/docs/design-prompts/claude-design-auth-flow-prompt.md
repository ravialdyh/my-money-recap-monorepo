# Prompt for Claude Design: authentication and onboarding flow

Paste everything below the line into Claude Design, inside the "My Money Recap" project (the one the design system was synced to). Attach `docs/PRD.md` if the project lets you attach files; otherwise the prompt is self-contained.

---

Design the end-to-end authentication and onboarding flow for My Money Recap, from app launch to the first Home screen, as a multi-artboard canvas. Use the synced "My Money Recap" design system for every element. Do not invent a new visual language; if something is missing from the system, build it from the closest existing component and label it "NEW" in an annotation so it can be added to the system later.

## Product facts you must respect

- Users are adults (18+). The kids look is a visual theme only. Money words stay plain: "salary", "saving target", "spend". Never "pocket money" or "allowance".
- Native iOS app, iPhone 390 × 844 artboards, portrait. Fonts Fredoka (headlines, numbers) and Nunito (everything else). Two radii only: cards 28, pills 999. No shadows anywhere; depth comes from white cards on a coloured ground and from illustrations overlapping the card top.
- One ground colour per screen. Every screen in this flow uses the violet ground (`ground.recap`, #5B4FE8) so the flow rhymes with the recap cover. White text on the ground, dark ink (#1E1B3A) inside cards.
- Primary button: ink pill (#1E1B3A, white label). Secondary button: white pill with ink label. Pressed = scale 0.97. Disabled = 40% opacity.
- Every tappable element is at least 44 × 44 pt.
- Session rule: the app never persists a session, so every cold launch starts at Sign in.

## Screens to draw, in one row, left to right, 80 px gap, titled exactly as below

**A0 · Launch.** Violet ground, the cover illustration centred, app name "My Money Recap" in Fredoka 700 beneath it, nothing else. Annotate: "Shows for at most 800 ms, then A1."

**A1 · Sign in.** Cover illustration hanging over a white card. Headline (Fredoka, size l) "Your month in money." Sub copy "Sign in to start your recap." Then three full-width buttons stacked with 12 px gap, in this order:
1. Sign in with Apple. Use Apple's own black button with the white Apple logo and the label "Sign in with Apple", pill radius. Apple mandates this style; it is the one allowed exception to the system.
2. "Continue with Google", white pill, Google "G" mark at left, ink label.
3. "Continue with email", ink pill.
Helper text under the buttons in white at 85% opacity: "By continuing you agree to the Terms and Privacy Policy."

**A2 · Enter email.** Eyebrow chip "Email", headline "Where should we send your link?", a white card containing one text field labelled "Email address" (placeholder "you@example.com", email keyboard), and an ink button "Send me a link". Button disabled until the field holds a valid email. Back chevron top-left in white.

**A3 · Check your inbox.** Envelope illustration placeholder (mark "ILLUSTRATION TBD, same style as cover"), headline "Check your inbox.", sub copy "We sent a link to maya@example.com. It works for 15 minutes." Two helper links: "Resend" (shown disabled with a 30-second countdown "Resend in 0:24") and "Use a different email".

**A4 · Name and avatar.** Eyebrow "Step 1 of 4", headline "What should we call you?", white card with a text field "First name" (value "Maya"), a second white card titled "Pick your avatar" holding a 3 × 2 grid of six round avatars. Draw the avatars as placeholder circles in six palette hues (violet, yellow, coral, green, sky, gold) labelled Fox, Otter, Panda, Frog, Owl, Rabbit, each marked "ILLUSTRATION TBD". The selected avatar (Otter) gets a 2 px white ring on a violet halo, matching the selected archetype-pill treatment. Ink button "Next", enabled.

**A5 · Currency.** Eyebrow "Step 2 of 4", headline "Which currency?", a white card with a search field "Search by name or code" and a list of rows: "GBP · British Pound" (selected, violet check), "IDR · Indonesian Rupiah", "BHD · Bahraini Dinar", "USD · US Dollar", "EUR · Euro", "SGD · Singapore Dollar". Rows follow the DataRow anatomy without the icon well. Ink button "Next".

**A6 · Salary and payday.** Eyebrow "Step 3 of 4", headline "Your monthly salary", white card with an amount field showing a fixed "£" prefix and the value "2,400" in Fredoka, then a second white card titled "Payday" with a horizontal day picker 1 to 31 and "25" selected as an ink pill. Helper text under the picker: "If the month is shorter, we use its last day." Ink button "Next", and a white text link beneath it "Skip for now".

**A7 · Saving target.** Eyebrow "Step 4 of 4", headline "How much do you want to keep?", white card with a large "20%" in Fredoka, a slider from 0 to 100 in steps of 5 built from the ProgressBar track with a round ink knob, and the live preview "£480 of £2,400 each month" in secondary text. Ink button "Done", white link "Skip for now".

**A8 · Home, first run.** This is where the flow lands. Chip row: "September 2026" and a "Settings" chip. Greeting "Hi, Maya" (Fredoka, size m) with the Otter avatar at 56 pt on the right. Hero card with StatHero eyebrow "Salary lands in 21 days" and value "£0", plus stat chips "Saved £0" (green) and "Spent £0" (coral). Target card titled "Saving target", right-aligned "£0 of £480", ProgressBar at 0, helper "Nothing spent yet. Streak day 4." Two buttons: "Log a spend" (ink) and "See last month" (white, disabled). Below, an empty recap list with the helper "Your first recap arrives on 1 October."

## Second row: states and variants, same titles with a suffix

- **A1 · Sign in / loading**: the Google button shows a spinner and the other two are disabled.
- **A1 · Sign in / error**: coral toast at the bottom "Sign-in was cancelled. Try again." with buttons re-enabled.
- **A2 · Enter email / invalid**: field with a coral 2 px border and helper "That doesn't look like an email address."
- **A3 · Check your inbox / resend ready**: "Resend" enabled.
- **A4 · Name and avatar / empty**: field empty, no avatar selected, Next disabled.
- **A8 · Home / skipped**: the Member skipped salary and target. Hero card becomes a prompt card "Set your monthly salary" with a chevron; target card becomes "Set a saving target". Both cards keep the same size as the filled versions.
- **A8 · Home / after payday**: hero eyebrow "Money in so far", value "£2,400", chips "Saved £2,215" and "Spent £185", target card "£2,215 of £480" with the bar full, helper "Keep spending under £61 a day and you'll make it."

## Third row: one flow artboard

A single wide artboard titled "Flow · Launch to Home" showing thumbnails of A0 to A8 connected with arrows and labels: A0 → A1 "auto"; A1 → A2 "Continue with email"; A1 → A4 "Apple or Google, first sign-in"; A1 → A8 "Apple or Google, returning member"; A2 → A3 "Send me a link"; A3 → A4 "tap link, first sign-in"; A3 → A8 "tap link, returning member"; A4 → A5 → A6 → A7 → A8 "Next / Done"; A6 → A7 and A7 → A8 also carry "Skip for now". Note on the artboard: "Onboarding progress is saved per screen; a force-quit resumes at the next unanswered step."

## Annotations to add beside each screen

Short sticky notes, not inside the artboard: validation rules (name 1 to 30 characters, letters, spaces, hyphens, apostrophes; amount greater than 0; payday 1 to 31), which buttons are disabled and why, and which elements are placeholders awaiting illustrations.

## Do not

- Do not add a bottom tab bar, navigation bar, or a header title. Screens use the chip row and a back chevron only.
- Do not use shadows, gradients, or any radius other than 28 and 999.
- Do not add an age field, a password field, or a "remember me" toggle. None of these exist in the product.
- Do not draw the push-permission prompt; it appears later, after the first spend is logged.
