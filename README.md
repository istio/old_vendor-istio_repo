# vendor-istio
(former) Vendored dependencies for the main istio repo

We tried vendor as a submodule and while it has a lot of benefits it also proved too much complexity ([#4746](https://github.com/istio/istio/issues/4746)) so we switched to fully checked-in vendor.

---
Historical FAQ:
---
~Istio uses vendored dependencies through a submodule. This avoids having large unrelated commits mixed with our actual work while having the benefits of hermetic reproducible builds, tracked changes and fast dependencies download.~

## How do I get setup?
One time / first time setup

### Existing clone/fork:

If you have a pre Feb 15 2018 fork or clone:

```
rm -rf ./vendor/ # only do that if vendor/ is a plain directory without a .git
make init
( cd vendor && git checkout master && git pull )
```

### New repo
```
git clone --recurse-submodules https://github.com/istio/istio.git
cd istio/vendor && git checkout master && git pull
```

or even better (as it will place the source tree in the right location for go):
```
go get istio.io/istio
```

## How do I stay in sync?

1 time setup that helps if you will be changing the dependencies:
```
cd vendor && git checkout master && git pull
```
then regularly
```
make pull
```

or manually, when you notice that [Gopkg.lock](https://github.com/istio/istio/blob/master/Gopkg.lock) changed:

```
git submodule sync && git submodule update
```

## How do I check my setup is ok vendor wise?

```
# in ~/go/src/istio.io/istio
cd vendor; git status; git remote -v
```
should output
```
On branch master
Your branch is up to date with 'origin/master'.

nothing to commit, working tree clean
origin	https://github.com/istio/vendor-istio.git (fetch)
origin	https://github.com/istio/vendor-istio.git (push)
```

If you get something else, see setup above
try:
```
cd vendor
git reset --hard origin/master
```

## How big is vendor, how much of a download?
```
# during clone, submodule download is ~8Mb
Receiving objects: 100% (4280/4280), 7.96 MiB | 6.93 MiB/s, done.
# size once expanded
$ du -hs vendor/
 54M	vendor/
# it used to be 407Mb before
```
This is thanks to the [pruning](https://github.com/istio/istio/pull/3348/files#diff-836546cc53507f6b2d581088903b1785R39) setup in go dep.

## How do I add / change a dependency?

See the next question for an overview

1. You need to be able to create a branch on https://github.com/istio/vendor-istio (be an org contributor, if you are not, ask someone who is). This is because the PR you make in istio/istio to validate the vendor change will need a valid vendor SHA for the CI to pull it successfully (and thus need to be reachable, alternatively could also change the submodule pointer but that's dangerous/not recommended)

1. Make sure your `go` and `dep` are recent enough (understands pruning etc...): 
   ```shell
   # make sure your go version is current, as of this writing at least go1.10.1
   $ go version
   go version go1.10.1 darwin/amd64
   ```
   ```shell
   go get -u github.com/golang/dep/cmd/dep
   ls -l `which dep` # should show now's timestamp
   ```

1. Use your new dependency in new code (most cases) or edit the [Gopkg.toml](https://github.com/istio/istio/blob/master/Gopkg.toml) (special cases).

1. Run `make depend.update` (or `make depend.update DEPARGS="--update some.package/to.be.updated"` for instance to update only 1 package - but that doesn't seem to always succeed so the target without DEPARGS is better for now. or manually `dep ensure --update <name.of.the.package.added.or.updated>`)

1. If getting errors you may need to do a `make depend.update.full` and/or delete directories in your ~/go/dep and/or ask for help

1. If you ran dep by hand (don't!), make sure to copy the Gopkg.* to vendor/

1. Note that `Gopkg.*` and `vendor` are in `.gitignore` to avoid accidental changes so you will need to manually add the changes

1. `cd vendor/` and `git status` / check `git diff`, make a PR for the changes. **DO NOT PUSH YOUR PR IF THE CHANGES ARE UNEXPECTED - For instance if you see 20k files changed, something is wrong...**

1. Do not forget to add all "untracked" files/directories to your vendor PR, that's how new files become available.

1. make sure `make vendor.check depend.status` doesn't error out before submitting your PRs

1. put [vendor-change] for clarity in your istio/istio PR title and cross reference the 2 PRs

Once approved/merged (or on a branch of the vendor repo): submit your PR once the submodule change doesn't show `-dirty`

You will need to change the istio/istio PR again after the istio/vendor-istio one is merged (as the SHA changes from a branch SHA to master SHA)

Make sure both PR get merged within 1h or each other so master of the 2 repos don't stay out of sync for long and so other vendor changes can proceed. All vendor changes by nature must be serialized.

## Check in Sequence for Vendor (7 steps, 2 PRs dance)

1. make your changes as above (`make depend.update`)
1. istio/vendor-istio PR, must be in a branch (not a fork)
1. istio/istio PR with your changes (can be a fork) and referencing sha from previous PR
1. after all tests pass
1. get the first (vendor) PR merged
1. update the second (istio) PR to use the master vendor sha
1. tests should stay green, merge second PR

## Should I merge changes in `vendor/`?

**No**, If you see in your `git status`
```
	modified:   vendor (untracked content)
```
Do not add this to your PR. Unlike you explicitly wanted to make dependency changes, like in the previous question.

Try
```
make submodule-sync
```
to remove that

## I did something and now my PR has a vendor change, how do I fix it ?

```
# Setup a remote that points for sure to istio's master (you could use upstream if it's setup and fetched)
git remote remove vendor_fix
git remote add -t master -m master -f vendor_fix https://github.com/istio/istio.git
# This gets Gopkg.* and vendor sha from istio's master:
git checkout vendor_fix -- Gopkg.lock Gopkg.toml vendor
# This syncs vendor file and vendor directory - if you get error; rm -rf vendor; make submodule-sync
git submodule update
```

The add, commit, push to your PR to fix it

## How do we auto update dependencies?

A periodic job can run `dep ensure` and commit a vendor submodule PR first and if approved and passing the test move the `vendor/` in istio/istio

## Can I just edit the lock file manually or run dep by hand ?

**NO** if you do it'll mess up vendor for the next person and your change will be lost at next update.
