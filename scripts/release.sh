#!/bin/bash

# Flow Release Script - Idempotent Implementation of RELEASE_STRATEGY.md
# This script implements the complete release process for the Flow library
# following semantic versioning and the documented release strategy.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHANGELOG_FILE="$REPO_ROOT/CHANGELOG.md"
GO_MOD_FILE="$REPO_ROOT/go.mod"
EXAMPLES_DIR="$REPO_ROOT/examples"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Flow Release Script

USAGE:
    ./scripts/release.sh [OPTIONS] <RELEASE_TYPE>

ARGUMENTS:
    RELEASE_TYPE    One of: patch, minor, major
                   - patch: Bug fixes, backwards compatible (x.y.Z)
                   - minor: New features, backwards compatible (x.Y.0)
                   - major: Breaking changes (X.0.0)

OPTIONS:
    -h, --help          Show this help message
    -d, --dry-run       Show what would be done without making changes
    -f, --force         Skip interactive confirmations
    -c, --check-only    Only run pre-release checks without creating release
    --skip-tests        Skip running tests (not recommended)
    --skip-examples     Skip building examples
    --pre-release TYPE  Create pre-release with suffix (alpha, beta, rc)

EXAMPLES:
    ./scripts/release.sh patch              # Create next patch version (e.g., v1.0.1)
    ./scripts/release.sh minor              # Create next minor version (e.g., v1.1.0)
    ./scripts/release.sh major              # Create next major version (e.g., v2.0.0)
    ./scripts/release.sh --dry-run patch    # Preview what would happen
    ./scripts/release.sh --check-only       # Run pre-release checks only
    ./scripts/release.sh minor --pre-release rc  # Create v1.1.0-rc.1

RELEASE TYPES:
    - patch: Bug fixes, performance improvements, documentation
    - minor: New features, backwards compatible API additions
    - major: Breaking API changes, architectural changes

EOF
}

# Parse command line arguments
DRY_RUN=false
FORCE=false
CHECK_ONLY=false
SKIP_TESTS=false
SKIP_EXAMPLES=false
RELEASE_TYPE=""
PRE_RELEASE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -c|--check-only)
            CHECK_ONLY=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --skip-examples)
            SKIP_EXAMPLES=true
            shift
            ;;
        --pre-release)
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                PRE_RELEASE="$2"
                shift 2
            else
                log_error "--pre-release requires a type (alpha, beta, rc)"
                exit 1
            fi
            ;;
        -*)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
        patch|minor|major)
            if [[ -z "$RELEASE_TYPE" ]]; then
                RELEASE_TYPE="$1"
            else
                log_error "Multiple release types specified: $RELEASE_TYPE and $1"
                exit 1
            fi
            shift
            ;;
        *)
            log_error "Invalid release type: $1"
            log_error "Valid types: patch, minor, major"
            show_help
            exit 1
            ;;
    esac
done

# Get current version from git tags
get_current_version() {
    local current_tag
    current_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

    if [[ -z "$current_tag" ]]; then
        echo "v0.0.0"  # Default for first release
    else
        echo "$current_tag"
    fi
}

# Calculate next version based on release type
calculate_next_version() {
    local release_type="$1"
    local pre_release="$2"
    local current_version
    current_version=$(get_current_version)

    # Remove 'v' prefix and any pre-release suffix
    local version_core="${current_version#v}"
    version_core="${version_core%%-*}"

    # Parse version components
    local major minor patch
    IFS='.' read -r major minor patch <<< "$version_core"

    # Calculate next version based on type
    case "$release_type" in
        patch)
            patch=$((patch + 1))
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        *)
            log_error "Invalid release type: $release_type"
            exit 1
            ;;
    esac

    # Construct new version
    local new_version="v${major}.${minor}.${patch}"

    # Add pre-release suffix if specified
    if [[ -n "$pre_release" ]]; then
        # Check if there's already a pre-release with this version
        local existing_pre
        existing_pre=$(git tag -l "${new_version}-${pre_release}.*" | sort -V | tail -1)

        if [[ -n "$existing_pre" ]]; then
            # Extract the number and increment
            local pre_num
            pre_num=$(echo "$existing_pre" | sed "s/.*${pre_release}\.//")
            pre_num=$((pre_num + 1))
            new_version="${new_version}-${pre_release}.${pre_num}"
        else
            new_version="${new_version}-${pre_release}.1"
        fi
    fi

    echo "$new_version"
}

# Validate release type
validate_release_type() {
    local release_type="$1"

    case "$release_type" in
        patch|minor|major)
            return 0
            ;;
        *)
            log_error "Invalid release type: $release_type"
            log_error "Valid types: patch, minor, major"
            exit 1
            ;;
    esac
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
}

# Check if working directory is clean
check_working_directory() {
    log_info "Checking working directory status..."

    if [[ -n "$(git status --porcelain)" ]]; then
        log_error "Working directory is not clean. Please commit or stash changes:"
        git status --short
        exit 1
    fi

    log_success "Working directory is clean"
}

# Check if we're on the main branch
check_main_branch() {
    local current_branch
    current_branch=$(git branch --show-current)

    if [[ "$current_branch" != "main" ]]; then
        log_warning "Not on main branch (currently on: $current_branch)"
        if [[ "$FORCE" != "true" ]]; then
            read -p "Continue anyway? [y/N]: " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "Aborted by user"
                exit 1
            fi
        fi
    fi
}

# Check if tag already exists
check_tag_exists() {
    local version="$1"

    if git tag -l | grep -q "^$version$"; then
        log_error "Tag $version already exists"
        if [[ "$FORCE" != "true" ]]; then
            log_error "Use --force to overwrite existing tag (not recommended)"
            exit 1
        else
            log_warning "Will overwrite existing tag $version"
        fi
    fi
}

# Validate Go module
validate_go_module() {
    log_info "Validating Go module..."

    # Check if go.mod exists
    if [[ ! -f "$GO_MOD_FILE" ]]; then
        log_error "go.mod file not found: $GO_MOD_FILE"
        exit 1
    fi

    # Tidy dependencies
    if ! go mod tidy; then
        log_error "Failed to tidy Go module"
        exit 1
    fi

    # Verify dependencies
    if ! go mod verify; then
        log_error "Go module verification failed"
        exit 1
    fi

    # Run go vet
    if ! go vet ./...; then
        log_error "Go vet failed"
        exit 1
    fi

    log_success "Go module validation passed"
}

# Run tests
run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_warning "Skipping tests (--skip-tests specified)"
        return
    fi

    log_info "Running test suite..."

    # Run tests with race detection
    if ! go test -v -race ./...; then
        log_error "Tests failed"
        exit 1
    fi

    log_success "All tests passed"
}

# Build examples
build_examples() {
    if [[ "$SKIP_EXAMPLES" == "true" ]]; then
        log_warning "Skipping example builds (--skip-examples specified)"
        return
    fi

    log_info "Building examples..."

    if [[ ! -d "$EXAMPLES_DIR" ]]; then
        log_warning "Examples directory not found: $EXAMPLES_DIR"
        return
    fi

    local dist_dir="$REPO_ROOT/dist/examples"
    mkdir -p "$dist_dir"

    for dir in "$EXAMPLES_DIR"/*/; do
        if [[ -d "$dir" ]]; then
            local example_name
            example_name=$(basename "$dir")
            log_info "Building example: $example_name"

            if ! (cd "$dir" && go build -o "$dist_dir/$example_name" .); then
                log_error "Failed to build example: $example_name"
                exit 1
            fi
        fi
    done

    log_success "All examples built successfully"
}

# Update changelog
update_changelog() {
    local version="$1"
    local changelog_temp
    changelog_temp=$(mktemp)

    log_info "Updating changelog for $version..."

    # Check if version already exists in changelog
    if grep -q "## \[$version\]" "$CHANGELOG_FILE"; then
        log_warning "Version $version already exists in changelog"
        return
    fi

    # Get the date
    local release_date
    release_date=$(date +%Y-%m-%d)

    # Generate changelog entries from git commits
    local changelog_entries
    local previous_tag
    previous_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

    if [[ -n "$previous_tag" ]]; then
        log_info "Generating changelog from commits since $previous_tag..."
        changelog_entries=$(git log --pretty=format:"- %s" "$previous_tag"..HEAD | grep -v "^- chore:" | head -20)
    else
        log_info "No previous tags found, using recent commits..."
        changelog_entries=$(git log --pretty=format:"- %s" --max-count=10 | grep -v "^- chore:")
    fi

    # If no meaningful commits, use a default entry
    if [[ -z "$changelog_entries" ]]; then
        changelog_entries="- Release $version"
    fi

    # Create updated changelog
    {
        # Copy everything up to [Unreleased]
        sed '/## \[Unreleased\]/q' "$CHANGELOG_FILE"

        # Add new version entry
        echo ""
        echo "## [$version] - $release_date"
        echo ""
        echo "### Added"
        echo "$changelog_entries"
        echo ""

        # Copy the rest, skipping the [Unreleased] line
        sed '1,/## \[Unreleased\]/d' "$CHANGELOG_FILE"
    } > "$changelog_temp"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Would update changelog with:"
        echo "$changelog_entries"
        rm "$changelog_temp"
    else
        mv "$changelog_temp" "$CHANGELOG_FILE"
        log_success "Changelog updated with generated entries"
    fi
}

# Create git tag and push
create_and_push_tag() {
    local version="$1"

    log_info "Creating and pushing tag $version..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Would create tag $version (dry run mode)"
        log_info "Would push to origin main (dry run mode)"
        log_info "Would push tag $version (dry run mode)"
        return
    fi

    # Commit changelog changes
    git add "$CHANGELOG_FILE"
    git commit -m "chore: prepare for $version release"

    # Create tag
    git tag "$version" -m "Release $version"

    # Push main branch and tag
    git push origin main
    git push origin "$version"

    log_success "Tag $version created and pushed"
}

# Main execution
main() {
    log_info "Flow Release Script Starting..."

    # Change to repository root
    cd "$REPO_ROOT"

    # Validate inputs
    if [[ "$CHECK_ONLY" != "true" && -z "$RELEASE_TYPE" ]]; then
        log_error "Release type is required unless using --check-only"
        show_help
        exit 1
    fi

    # Validate release type and calculate version
    if [[ -n "$RELEASE_TYPE" ]]; then
        validate_release_type "$RELEASE_TYPE"
        local current_version
        current_version=$(get_current_version)
        VERSION=$(calculate_next_version "$RELEASE_TYPE" "$PRE_RELEASE")
        log_info "Current version: $current_version"
        log_info "Target version: $VERSION ($RELEASE_TYPE release)"
        if [[ -n "$PRE_RELEASE" ]]; then
            log_info "Pre-release type: $PRE_RELEASE"
        fi
    fi

    # Pre-release checks
    log_info "Running pre-release checks..."
    check_git_repo
    check_working_directory
    check_main_branch

    if [[ -n "$VERSION" ]]; then
        check_tag_exists "$VERSION"
    fi

    # Validate Go module and run tests
    validate_go_module
    run_tests
    build_examples

    # Exit if check-only mode
    if [[ "$CHECK_ONLY" == "true" ]]; then
        log_success "All pre-release checks passed"
        exit 0
    fi

    # Show summary and confirm
    echo
    log_info "Release Summary:"
    echo "  Release Type: $RELEASE_TYPE"
    echo "  Version: $VERSION"
    if [[ -n "$PRE_RELEASE" ]]; then
        echo "  Pre-release: $PRE_RELEASE"
    fi
    echo "  Dry Run: $DRY_RUN"
    echo "  Force: $FORCE"
    echo

    if [[ "$FORCE" != "true" && "$DRY_RUN" != "true" ]]; then
        read -p "Proceed with release? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Release aborted by user"
            exit 0
        fi
    fi

    # Execute release
    update_changelog "$VERSION"
    create_and_push_tag "$VERSION"

    echo
    log_success "Release $VERSION completed successfully!"
    log_info "GitHub Actions will now:"
    log_info "  - Run full test suite"
    log_info "  - Create GitHub release"
    log_info "  - Build and attach examples"
    log_info "  - Generate release notes"
    log_info ""
    log_info "Monitor the release at: https://github.com/joemocha/flow/actions"
}

# Execute main function
main "$@"
