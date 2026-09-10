# Changelog Template

PayStream's `CHANGELOG.md` follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/). This document defines the format contributors and maintainers should use when adding an entry.

## Structure

Each release is its own heading, newest first, with an ISO date:

```markdown
## [1.4.0] - 2026-05-12

### Added
- Short, user-facing description of the new capability (#123)

### Changed
- Behavior that changed for existing users, and why it matters to them (#118)

### Deprecated
- Features still present but scheduled for removal, with a migration note (#101)

### Removed
- Features removed in this release (#95)

### Fixed
- Bug fixes, phrased as the symptom that's now gone (#130)

### Security
- Vulnerability fixes. Do not include exploit details — link the advisory instead (#140)
```

An `## [Unreleased]` section at the top of `CHANGELOG.md` collects entries as PRs merge, and is renamed to a version heading at release time.

## Category guidelines

Only include the categories that apply to a given release — omit empty ones.

| Category | Use for |
|---|---|
| `Added` | New features, endpoints, SDK methods, config options |
| `Changed` | Changes to existing behavior, including performance improvements |
| `Deprecated` | Soon-to-be-removed features |
| `Removed` | Features removed in this release |
| `Fixed` | Bug fixes |
| `Security` | Vulnerability fixes (coordinate disclosure timing with `security/bounty.md`) |

## Writing a good entry

- Write from the user's perspective, not the implementer's: "Batch payouts now accept up to 100 items per request" rather than "refactored batch validation loop."
- One line per change. If it needs more than one line, it's probably two changelog entries.
- Reference the PR or issue number in parentheses at the end of the line.
- Breaking changes must be called out explicitly, e.g. prefix the line with `**BREAKING:**`.

## When to add an entry

Add an entry under `## [Unreleased]` in the same PR that makes the change — not retroactively at release time. This keeps the changelog accurate even if a release is cut before every intended PR has landed.

## Example

```markdown
## [Unreleased]

### Added
- ACH bank transfer support as an alternate settlement rail alongside Stellar (#61)
- Recurring payment schedules with daily, weekly, and monthly intervals (#53)

### Changed
- Batch payout requests now validate each item independently, so one invalid
  recipient no longer rejects the entire batch (#57)
```
