<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Check whether documentation needs updating after a feature or fix.

Review each surface and update any that are affected by the change:

1. **Hugo docs** (`docs/content/`) -- reference pages, guides, configuration
2. **README** -- features list, install instructions, keybindings, tech stack
3. **Website** (`docs/layouts/index.html`) -- landing page pitch copy,
   feature highlights
4. **Demo GIF** -- if UI/UX changed, run `/record-demo`
5. **Screenshot tapes** (`docs/tapes/`) -- if affected screens changed, re-
   capture with `nix run '.#capture-screenshots'`

Keep README and website in sync: when changing content on one (features,
install instructions, keybindings, tech stack, pitch copy), update the other
to match.
