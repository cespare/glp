## To do

* In `Verify()`, remove any repos not needed by deps in the pinlist. Otherwise we're not really providing an
  environment isolated to the packages listed there.
  - Does this mean revisiting the behavior of using whatever rev is in an already-cached repo in `Sync()`?
  - In any case, `Sync()` should be printing out what revs it's using for new repos.
* Hijack `-h`, `help`, etc to provide useful information about glp.

## Things to test

* Do tests for each VCS
  - ~~git~~
  - hg
  - bzr
* Test when the cache dir isn't there
* When the cache dir is too recent
* When the cache dir is too old
* When the pinned version is not in the cache (pull? Or maybe a bad rev)
* When there is a rev conflict between two packages from the same repo
* When the whole cache dir doesn't exist
* When you have pinned deps that aren't used by the code anymore
* When you comment out a dep, sync, and uncomment it and sync again (shouldn't change the version)
* When there are dirty changes in a dep repo
* Running glp from not the project root
