# brewer

`brewer` is a small go program with one job: generate the text of a homebrew formula for `bbctl` using the latest gitlab release of `bbctl`.

## How does it work?

- hits the repo1 API to find out the latest release of `bbctl`
- fetches that release's tag `name`
- fetches that release's source tarball and generates a `sha256sum` of the release tarball
- Feeds `name` and `sha256sum` into a go template at `./scripts/brewer/templates` that then prints out the full text of a compliant homebrew formula for `bbctl`

## How to use it?

### Generate an updated `bbctl.rb`

Generate a new `bbctl.rb` formula in this directory: `make brew-regenerate > bbctl.rb `

### Open an MR branch on `homebrew-tools-public`

- Clone the BigBang [homebrew-tools-public](https://repo1.dso.mil/big-bang/homebrew-tools-public) repo: `git clone https://repo1.dso.mil/big-bang/homebrew-tools-public.git`

From within your local checkout of `homebrew-tools-public`: (`cd homebrew-tools-public`)

- Copy your newly generated `bbctl.rb` and overwrite the existing `Formula/bbctl.rb` (e.g. `cp ../bbctl/bbctl.rb ./Formula/bbctl.rb`)
- Open a local feature branch in `homebrew-tools-public` and commit this change:
 
```bash
git checkout -b myname-bump-bbctl-x_y_z
git add Formula/bbctl.rb
git commit -m "Bumps bbctl to version x.y.z"
```

Once committed locally, you can now test your updated formula:

 ```console
 homebrew-tools-public on  djp-adds-bbctl-formula
   ❯ make brew
   brew uninstall -f "bbctl"
   Uninstalling bbctl... (7 files, 91.6MB)
   brew untap -f "bigbang/tools-public" || echo "untap complete, swallowing return code..."
   Untapping bigbang/tools-public...
   Untapped 1 formula (34 files, 26.2KB).
   brew tap "bigbang/tools-public" .
   ==> Tapping bigbang/tools-public
   Cloning into '/opt/homebrew/Library/Taps/bigbang/homebrew-tools-public'...
   done.
   Tapped 1 formula (47 files, 33.4KB).
   brew reinstall -f "bigbang/tools-public/bbctl"
   ==> Fetching bigbang/tools-public/bbctl
   ==> Downloading https://repo1.dso.mil/big-bang/product/packages/bbctl/-/archive/0.7.6/bbctl-0.7.6.tar.gz
   Already downloaded: /Users/daniel/Library/Caches/Homebrew/downloads/49e83cc01d7821476f7806a0f0a8fbd2467882bbc632183cf1ba15eddcaf76b5--bbctl-0.7.6.tar.gz
   ==> Reinstalling bigbang/tools-public/bbctl
   ==> go build -ldflags=-s -w -X static.buildDate=2025-01-09T23:35:00Z
   🍺  /opt/homebrew/Cellar/bbctl/0.7.6: 7 files, 91.6MB, built in 6 seconds
   ==> Running `brew cleanup bbctl`...
   Disable this behaviour by setting HOMEBREW_NO_INSTALL_CLEANUP.
   Hide these hints with HOMEBREW_NO_ENV_HINTS (see `man brew`).
   brew audit --verbose --strict --formula "bigbang/tools-public/bbctl"
   /opt/homebrew/Library/Homebrew/vendor/bundle/ruby/3.3.0/bin/bundle clean
   brew test --verbose --debug "bigbang/tools-public/bbctl"
   /opt/homebrew/Library/Homebrew/vendor/bundle/ruby/3.3.0/bin/bundle clean
   /opt/homebrew/Library/Homebrew/brew.rb (Formulary::FromTapLoader): loading bigbang/tools-public/bbctl
   /opt/homebrew/Library/Homebrew/brew.rb (Formulary::FromAPILoader): loading go
   ==> Testing bigbang/tools-public/bbctl
   /opt/homebrew/Library/Homebrew/test.rb (Formulary::FromPathLoader): loading /opt/homebrew/Library/Taps/bigbang/homebrew-tools-public/Formula/bbctl.rb
   ==> /opt/homebrew/Cellar/bbctl/0.7.6/bin/bbctl
 ```

Assuming that passes, push your branch to repo1 (`git push origin`) and open a merge request for your change.