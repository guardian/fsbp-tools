# Commit messages

Use Conventional Commits for every commit message:

`<type>: <short imperative summary>`

An optional scope goes after the type, as in `fix(cli): handle empty input`. Add `!` before the colon only for a breaking change, as in `feat(cli)!: remove the old flag`. Choose the type based on the effect of the resulting change for consumers of the released program, not on which files were edited:

- Use `feat` for a new user-facing capability (minor release).
- Use `fix` for a user-facing correction, or `perf` for a performance improvement (patch release).
- Mark an incompatible change with `!` and explain the migration or incompatibility in a `BREAKING CHANGE:` footer (major release). Use the type that best describes the change as well.
- Use `docs`, `test`, `build`, `ci`, `refactor`, or `chore` for changes that do not change released behaviour. Use `chore(deps)` for dependency maintenance that fits this category.

Keep the subject specific and describe the change in the body when the impact is not clear from the subject. A release is run on demand, and its version bump is determined by the commit messages included since the previous release. Use a maintenance type only when the change has no effect on consumers of the released program.
