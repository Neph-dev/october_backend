# Conventional Commits Guide

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification to ensure consistent, meaningful commit messages that enable automated versioning and changelog generation.

## Why Conventional Commits?

✅ **Automated Versioning**: Commit types determine version bumps (major/minor/patch)  
✅ **Clear History**: Easy to understand what changed and why  
✅ **Auto-Generated Releases**: Release notes are automatically categorized  
✅ **Better Collaboration**: Team members can quickly scan commit history  
✅ **Enforced Standards**: CI/CD validates all commit messages  

## Quick Setup

Run this once to set up your local environment:

```bash
./scripts/setup-commit-rules.sh
```

This configures:
- Commit message template (opens in your editor)
- Optional local validation hook
- Git editor preferences

## Commit Message Format

```
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

### Components

#### Type (required)
Describes the kind of change:

| Type | Description | Version Impact | Example |
|------|-------------|----------------|---------|
| `feat` | New feature | Minor (1.0.0 → 1.1.0) | `feat: add user authentication` |
| `fix` | Bug fix | Patch (1.0.0 → 1.0.1) | `fix: resolve memory leak` |
| `docs` | Documentation only | Patch | `docs: update API guide` |
| `style` | Code style/formatting | Patch | `style: format imports` |
| `refactor` | Code restructuring | Patch | `refactor: simplify handler` |
| `perf` | Performance improvement | Patch | `perf: optimize database query` |
| `test` | Add/update tests | Patch | `test: add integration tests` |
| `build` | Build system changes | Patch | `build: update dependencies` |
| `ci` | CI/CD changes | Patch | `ci: add deployment step` |
| `chore` | Maintenance tasks | Patch | `chore: update .gitignore` |
| `revert` | Revert previous commit | Patch | `revert: undo feature X` |

#### Scope (optional)
Describes what part of the codebase is affected:

**Examples:**
- `api` - API endpoints
- `auth` - Authentication
- `db` - Database
- `cache` - Caching system
- `ui` - User interface
- `config` - Configuration
- `deploy` - Deployment scripts

**Usage:**
```bash
git commit -m "feat(api): add market data endpoint"
git commit -m "fix(auth): correct JWT validation"
```

#### Subject (required)
Short description of the change:

**Rules:**
- ✅ Max 100 characters
- ✅ Use imperative mood: "add" not "added" or "adds"
- ✅ Don't capitalize first letter
- ✅ No period at the end

**Good:**
```
feat: add user authentication
fix: resolve memory leak in RSS processor
docs: update deployment guide
```

**Bad:**
```
feat: Added user authentication.    # ❌ Past tense, period
Fix memory leak                      # ❌ Missing type
feat: ADDING USER AUTHENTICATION     # ❌ All caps
```

#### Body (optional)
Detailed explanation:

**When to use:**
- Complex changes requiring explanation
- Breaking changes with migration notes
- Context about why the change was made

**Format:**
- Wrap at 72 characters
- Separate from subject with blank line
- Explain what and why, not how

**Example:**
```
feat: redesign authentication system

The previous JWT implementation had security vulnerabilities.
This new approach uses refresh tokens and implements rate limiting.

Users will need to re-authenticate after this update.
```

#### Footer (optional)
Additional metadata:

**Breaking Changes:**
```
BREAKING CHANGE: API response format changed
```

**Issue References:**
```
Closes #123
Fixes #456, #789
See also: #321
```

**Complete Example:**
```
feat(api)!: redesign response format

All API endpoints now return data in a standardized envelope
format with metadata, improving consistency across the API.

BREAKING CHANGE: Clients must update to handle new response structure.
Previously: { ...data }
Now: { success: true, data: {...}, metadata: {...} }

Closes #123
```

## Breaking Changes

Breaking changes can be indicated two ways:

### Method 1: Exclamation Mark
```bash
git commit -m "feat!: redesign API response format"
git commit -m "fix(auth)!: change token validation logic"
```

### Method 2: Footer
```bash
git commit -m "feat: redesign authentication system

BREAKING CHANGE: Old auth tokens are no longer valid"
```

Both trigger a **major version bump** (1.0.0 → 2.0.0)

## Examples by Scenario

### Adding a Feature
```bash
# Simple feature
git commit -m "feat: add user authentication"

# Feature with scope
git commit -m "feat(api): add market data endpoint"

# Feature with body
git commit -m "feat(cache): implement Redis caching

Adds Redis support for caching API responses, reducing
database load and improving response times significantly."
```

### Fixing a Bug
```bash
# Simple fix
git commit -m "fix: resolve memory leak in RSS processor"

# Fix with scope and issue reference
git commit -m "fix(auth): correct JWT token validation

Closes #456"

# Fix with explanation
git commit -m "fix(db): prevent connection pool exhaustion

The connection pool was not releasing connections properly
under high load, causing timeouts. This implements proper
connection lifecycle management."
```

### Documentation
```bash
git commit -m "docs: update README with deployment steps"
git commit -m "docs(api): add examples for market endpoints"
git commit -m "docs: fix typo in CONTRIBUTING guide"
```

### Refactoring
```bash
git commit -m "refactor: simplify error handling logic"
git commit -m "refactor(handlers): extract validation into middleware"
```

### Tests
```bash
git commit -m "test: add integration tests for AI service"
git commit -m "test(api): add market endpoint tests"
git commit -m "test: increase coverage for auth module"
```

### Chores and Maintenance
```bash
git commit -m "chore: update dependencies"
git commit -m "chore: clean up unused imports"
git commit -m "build: upgrade Go to 1.21"
git commit -m "ci: add caching to GitHub Actions"
```

### Multiple Changes (Use Multiple Commits)
```bash
# ❌ Bad: Multiple unrelated changes in one commit
git commit -m "fix bugs and add features"

# ✅ Good: Separate commits for each logical change
git commit -m "fix: resolve memory leak"
git commit -m "feat: add caching layer"
git commit -m "docs: update performance guide"
```

## Validation

### Local Validation (if hook installed)
When you commit, your message is validated locally:

```bash
git commit -m "invalid message"
# ❌ Invalid commit message format
# Your commit message: invalid message
# Required format: <type>(<scope>): <description>
```

### CI/CD Validation
All commits are validated in GitHub Actions:

- ✅ Pull requests are checked before merge
- ✅ Pushes to main are validated
- ✅ Invalid commits block deployment
- ✅ Detailed feedback provided in PR comments

## Fixing Invalid Commits

### Last Commit (Not Pushed)
```bash
# Amend the commit message
git commit --amend -m "feat: proper commit message"
```

### Last Commit (Already Pushed)
```bash
# Amend and force push
git commit --amend -m "feat: proper commit message"
git push --force-with-lease
```

### Multiple Commits
```bash
# Interactive rebase
git rebase -i HEAD~3  # Last 3 commits

# In the editor, change 'pick' to 'reword' for commits to fix
# Save and update each commit message
```

### In a Pull Request
```bash
# Fix commits locally
git rebase -i HEAD~N  # N = number of commits

# Force push to update PR
git push --force-with-lease
```

## Best Practices

### DO ✅

- **Be specific**: "fix: resolve timeout in RSS fetcher" not "fix: bug fix"
- **Use imperative mood**: "add" not "added" or "adds"
- **Keep subject short**: Max 100 characters
- **Add body for complex changes**: Explain the "why"
- **Reference issues**: "Closes #123"
- **One logical change per commit**: Small, focused commits
- **Test before committing**: Ensure code works

### DON'T ❌

- **Vague messages**: "fix stuff", "updates", "wip"
- **Past tense**: "added feature" should be "feat: add feature"
- **Multiple unrelated changes**: Split into separate commits
- **Skip the type**: "update API" should be "feat: update API"
- **Overly long subjects**: Keep it concise
- **Commit broken code**: Each commit should be functional

## Tools and Commands

### Check Commit History
```bash
# View recent commits
git log --oneline -10

# View commits with full messages
git log -5

# Check if messages follow convention
git log --oneline | grep -v -E "^[a-f0-9]+ (feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?:"
```

### Amend Last Commit
```bash
# Change message only
git commit --amend -m "feat: corrected message"

# Add more changes to last commit
git add .
git commit --amend --no-edit
```

### View Commit Template
```bash
# See current template
git config --get commit.template

# Use template when committing
git commit  # Opens editor with template
```

## IDE Integration

### VS Code
Install extensions:
- **Conventional Commits** by vivaxy
- **Git Commit Message Editor** by Adam Boczek

### IntelliJ IDEA
Install plugins:
- **Conventional Commit** by lppedd
- **Git Commit Template** by MobileTribe

### Vim
Template is automatically loaded when you run `git commit`

## FAQ

**Q: What if I forget the format?**  
A: Run `git commit` without `-m` and the template will guide you

**Q: Can I use multiple types?**  
A: No, each commit should have one type. Split into multiple commits if needed.

**Q: What about merge commits?**  
A: Merge commits are automatically skipped by validation

**Q: How strict is the validation?**  
A: Very strict in CI/CD, but the local hook is optional for flexibility

**Q: What if I need to bypass validation?**  
A: Not recommended, but you can skip the local hook with `git commit --no-verify`

**Q: How do I see which commits are invalid?**  
A: Check the GitHub Actions "Validate Commit Messages" workflow output

**Q: What about commits from dependabot or other bots?**  
A: Bot commits are typically in branches and validated before merging to main

## Resources

- **Conventional Commits Spec**: https://www.conventionalcommits.org/
- **Semantic Versioning**: https://semver.org/
- **Project Quick Start**: [QUICKSTART.md](QUICKSTART.md)
- **Release Guide**: [RELEASES.md](RELEASES.md)
- **GitHub Actions Workflow**: [.github/workflows/commit-lint.yml](.github/workflows/commit-lint.yml)

---

**Remember**: Good commit messages are a gift to your future self and your teammates! 🎁

