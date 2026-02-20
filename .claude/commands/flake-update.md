<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Update the Nix flake and handle all downstream consequences. Run this
periodically before committing or creating PRs.

Steps:

1. `nix flake update`
2. `nix build '.#micasa'` -- if it fails with a hash mismatch, temporarily
   set `vendorHash = pkgs.lib.fakeHash;` (not `""`) in `flake.nix` to get
   the expected hash, then paste the real hash and rebuild
3. `go test -shuffle=on ./...` -- fix any breakage from updated packages
4. `nix run '.#osv-scanner'` -- a newer Go version may resolve previously
   ignored stdlib CVEs. Remove stale `[[IgnoredVulns]]` entries from
   `osv-scanner.toml` when their underlying vuln is fixed
5. `nix run '.#pre-commit'` -- verify formatting and linters still pass
6. `nix run '.#deadcode'` -- verify no new unreachable exports

If osv-scanner finds new vulnerabilities, try updating the dependency first
(`go get pkg@latest`). Only add `[[IgnoredVulns]]` if the vuln is genuinely
unreachable in micasa's code paths.
