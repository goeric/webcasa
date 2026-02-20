<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Update the Nix vendorHash after adding or changing Go dependencies.

Steps:

1. `nix build '.#micasa'`
2. If it fails with a hash mismatch, temporarily set
   `vendorHash = pkgs.lib.fakeHash;` (not `""`) in `flake.nix` -- this
   avoids a noisy warning while giving you the expected hash in the error
3. Copy the correct hash from the error output and replace `fakeHash` with it
4. `nix build '.#micasa'` again to confirm it builds cleanly

Never use `vendorHash = "";` -- it produces a deprecation warning.
