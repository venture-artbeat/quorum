# The Quorum Fork

Due to the lack of an existing extension mechanism able to add precompiles to the Quorum private
chain, it has been necessary to fork the codebase in order to add them.

While this has given us the functionality that we needed, it has introduced additional complexity
when it comes to keeping in sync with upstream changes. This document exists in order to describe
the intended workflows for performing this sync.

## Repository Structure

Where the [upstream](https://github.com/Consensys/quorum) repository uses `master` as its default
branch, this fork has been changed to use a branch called `ec-precompiles` instead. It is intended
that the `master` branch in this fork remains identical to the `master` branch in the upstream
repository.

All following instructions for the workflows assume that this is the case.

## Updating from Upstream

So the upstream repository has some new features or critical security fixes that you want to
integrate into this fork. How does one do that?

1. Start by making sure that you have the upstream as a remote in your git configuration. You can
   check this by running `git remote -v`, and you should see the following:

   ```
   upstream        git@github.com:Consensys/quorum.git (fetch)
   ```

   If you do not see this listed, add the remote by running
   `git remote add upstream git@github.com:Consensys/quorum.git`.

2. Next, you want to pull the new changes from the upstream repository. First, swap to the `master`
   branch in the fork by running `git checkout master`, and then run the following to fetch and
   merge any new changes.

   ```sh
   git pull upstream master
   ```

   If this completes successfully, the master branch in the fork will be up to date with the master
   branch in the upstream repository.

3. Next, we need to find the hash of the latest commit on the `master` branch. Run `git log master`
   and record the hash corresponding to the `master` branch head. We will use `xxxxxx` as an example
   here.

4. Swap back to the `ec-precompiles` branch by running `git checkout ec-precompiles` and run
   `git pull` to make sure you have all of the latest changes from the remote.

5. Next, we put all of our additional changes on top of the new ones using a rebase. Run the
   following command:

   ```sh
   git rebase xxxxxx
   ```

   This will effectively "replay" all of our enhancement commits on top of the new `master`. Please
   note that **this may cause conflicts**. If any conflicts occur, fix them and continue the rebase
   as instructed by the git prompt.

6. Once the rebase has completed, all of the changes in the fork will now exist on top of the new
   changes from upstream. All that remains is to update the remote repository.

   ```sh
   git push --force
   ```

## History Rewriting

You will note that this approach requires the ability to rewrite history in the repository. This
unfortunately will require engineers to re-sync with the ec-precompiles branch after the update
process. This is the trade-off for having a relatively simple branch architecture and making the
update process relatively simple.

In order to allow this, though the `ec-precompiles` branch is protected, that protection explicitly
allows force pushing by any engineer with the appropriate access.
