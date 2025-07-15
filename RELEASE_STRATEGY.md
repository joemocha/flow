# Release Strategy for Flow

This document outlines the release strategy, versioning scheme, and publication process for the Flow library.

## 🏷️ Semantic Versioning

Flow follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version (X.y.z): Incompatible API changes
- **MINOR** version (x.Y.z): New functionality in a backwards compatible manner
- **PATCH** version (x.y.Z): Backwards compatible bug fixes

### Version Format
- **Stable releases**: `v1.0.0`, `v1.1.0`, `v1.0.1`
- **Pre-releases**: `v1.0.0-rc.1`, `v1.0.0-beta.1`, `v1.0.0-alpha.1`

## 📋 Release Process

### 1. Pre-Release Checklist

Before creating any release:

- [ ] All tests pass (`go test -v -race ./...`)
- [ ] Code coverage is maintained or improved
- [ ] Documentation is updated
- [ ] Examples work correctly
- [ ] CHANGELOG.md is updated
- [ ] Version is bumped in relevant files

### 2. Release Types

#### Patch Release (v1.0.X)
- Bug fixes
- Performance improvements
- Documentation updates
- No API changes

#### Minor Release (v1.X.0)
- New adaptive behaviors
- New parameters
- New utility functions
- Backwards compatible API additions

#### Major Release (vX.0.0)
- Breaking API changes
- Architectural changes
- Removal of deprecated features

### 3. Release Commands

Use the automated release script in the `scripts/` directory:

```bash
# 1. Create a patch release (bug fixes)
./scripts/release.sh patch

# 2. Create a minor release (new features)
./scripts/release.sh minor

# 3. Create a major release (breaking changes)
./scripts/release.sh major

# 4. Create a pre-release
./scripts/release.sh minor --pre-release rc

# 5. Preview changes (recommended first)
./scripts/release.sh --dry-run patch

# 6. Run only pre-release checks
./scripts/release.sh --check-only
```

The script automatically:
- Calculates the next version from git tags
- Validates all pre-release requirements
- Updates CHANGELOG.md with git commit messages
- Creates and pushes the git tag
- Triggers GitHub Actions for release creation

## 🚀 Publication Channels

### 1. GitHub Releases
- Automated via GitHub Actions
- Includes built examples
- Generated changelog
- Release notes

### 2. Go Package Registry
- Automatic via go.mod and tags
- Available via `go get github.com/joemocha/flow`
- Documentation on pkg.go.dev

### 3. Community Announcements
- Go subreddit (/r/golang)
- Gopher Slack community
- Twitter/X with #golang hashtag
- Dev.to articles for major releases

## 📅 Release Schedule

### Initial Release (v1.0.0)
- **Target**: July 14, 2025 (ready now)
- **Focus**: Stable API, comprehensive documentation
- **Features**: All current adaptive behaviors

### Future Releases

#### v1.1.0 (Minor)
- Additional adaptive behaviors
- Performance optimizations
- Enhanced examples

#### v1.2.0 (Minor)
- Metrics and observability
- Additional utility functions
- Extended parameter support

#### v2.0.0 (Major - Future)
- Consider only if breaking changes are necessary
- Migration guide required
- Deprecation warnings in v1.x

## 🔄 Maintenance Strategy

### Long-term Support
- **v1.x**: Maintained for at least 2 years
- **Security patches**: Applied to all supported versions
- **Bug fixes**: Backported to latest minor version

### Deprecation Policy
- Features marked deprecated for at least one minor version
- Clear migration path provided
- Removal only in major versions

## 📊 Success Metrics

### Release Success Indicators
- [ ] CI/CD pipeline passes
- [ ] No critical issues reported within 48 hours
- [ ] Documentation is accessible
- [ ] Examples work as expected
- [ ] Community feedback is positive

### Long-term Success Metrics
- GitHub stars and forks
- Go package downloads
- Community contributions
- Issue resolution time
- Documentation quality scores

## 🛠️ Hotfix Process

For critical bugs in production:

1. Create hotfix branch from latest release tag
2. Apply minimal fix
3. Test thoroughly
4. Create patch release
5. Merge back to main

```bash
# Hotfix process
git checkout v1.0.0
git checkout -b hotfix/v1.0.1
# Apply fix
git commit -m "fix: critical bug description"
git tag v1.0.1
git push origin hotfix/v1.0.1
git push origin v1.0.1
```

## 📝 Changelog Format

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [1.0.1] - 2025-XX-XX
### Fixed
- Critical bug in retry logic

## [1.0.0] - 2025-07-14
### Added
- Initial release
- Adaptive node system
- Batch processing
- Retry logic
- Parallel execution
```

## 🎯 First Release Action Plan

Ready to publish v1.0.0:

1. **Final verification**: Run full test suite
2. **Create release**: Tag v1.0.0
3. **Monitor**: Watch for issues in first 48 hours
4. **Announce**: Share with Go community
5. **Support**: Respond to questions and issues

The Flow library is production-ready and follows all Go best practices for a successful v1.0.0 release!
