---
category: Components
---

Four 6px dots, the active one stretched to 22px. Omit on shareable screens (Recap screen 4) so screenshots
stay clean. Pass `onSelect` to make them controls (D46): every dot keeps its drawn size and gains a centred
44pt hit area, the same way a tappable `Chip` does. At accessibility text sizes the four Recap screens
become one vertical scroll and these dots are its section markers (D66).
