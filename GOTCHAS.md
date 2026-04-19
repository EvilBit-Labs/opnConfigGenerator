# Development Gotchas & Pitfalls

This document tracks non-obvious behaviors, common pitfalls, and architectural "gotchas" in the opnConfigGenerator codebase to assist future maintainers and contributors.

It is also worthwhile to check the [opnDossier GOTCHAS](https://raw.githubusercontent.com/EvilBit-Labs/opnDossier/refs/heads/main/GOTCHAS.md) for related information.

## 1. Testing & Concurrency

## 2. Plugin Architecture

## 3. Data Processing

## 4. Generator Engine

## 5. CLI Flag Wiring

## 6. Validator

## 7. CommonDevice to Device Serializer

## 8. Git Tagging

### 8.1 Tag the Squash-Merge Commit on Main

When tagging a release after a squash-merge PR, always tag the resulting commit **on `main`**, not the PR branch head. Squash-merge creates a new commit on `main` that is not an ancestor of the branch commits. If you tag the branch head instead, the tag points to an orphaned commit that `git log main` and `git describe` will never reach.

- **Symptom:** `git tag --merged main` does not list the release tag; `git describe` on `main` skips the version.
- **Fix:** `git checkout main && git pull && git tag vX.Y.Z && git push origin vX.Y.Z`
- **Prevention:** Always switch to `main` and pull before tagging. Never tag from the feature branch after merge.
