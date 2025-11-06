#!/bin/bash

# October Backend - Git Conventional Commits Setup Script
# This script configures your local git repository to use conventional commits

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 Setting up Git Conventional Commits for October Backend"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if we're in a git repository
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    echo "❌ Error: Not a git repository"
    echo "Please run this script from within the October Backend repository"
    exit 1
fi

# Set commit message template
echo "📝 Configuring commit message template..."
git config commit.template .gitmessage
echo "✅ Commit template configured"
echo ""

# Configure commit message editor (if not already set)
if ! git config --get core.editor > /dev/null 2>&1; then
    echo "📝 Setting default git editor to vim..."
    git config core.editor "vim"
    echo "✅ Editor configured"
    echo ""
fi

# Optional: Install commit message hook
echo "Would you like to install a local commit-msg hook to validate commits? (y/n)"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "📝 Installing commit-msg hook..."
    
    # Create hooks directory if it doesn't exist
    mkdir -p .git/hooks
    
    # Create commit-msg hook
    cat > .git/hooks/commit-msg << 'EOF'
#!/bin/bash

# Conventional Commits validation hook
# This hook validates commit messages locally before they are committed

commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")

# Skip merge commits
if echo "$commit_msg" | grep -qE "^Merge "; then
    exit 0
fi

# Define valid commit types
VALID_TYPES="feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"

# Conventional commit regex pattern
PATTERN="^(${VALID_TYPES})(\(.+\))?!?: .{1,100}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get just the subject line (first line of commit message)
subject=$(echo "$commit_msg" | head -n1)

# Validate commit message
if ! echo "$subject" | grep -qE "$PATTERN"; then
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}❌ Invalid commit message format${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}Your commit message:${NC}"
    echo "$subject"
    echo ""
    echo -e "${YELLOW}Required format:${NC}"
    echo "  <type>(<scope>): <description>"
    echo ""
    echo -e "${YELLOW}Valid types:${NC}"
    echo "  feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  feat: add user authentication"
    echo "  feat(api): add market data endpoint"
    echo "  fix: resolve memory leak in RSS processor"
    echo "  docs: update API documentation"
    echo "  feat!: redesign API response format"
    echo ""
    
    # Provide specific feedback
    if ! echo "$subject" | grep -qE "^(${VALID_TYPES})"; then
        echo -e "${RED}Issue: Missing or invalid type prefix${NC}"
    elif ! echo "$subject" | grep -qE ": "; then
        echo -e "${RED}Issue: Missing colon and space after type/scope${NC}"
    elif [ ${#subject} -lt 10 ]; then
        echo -e "${RED}Issue: Description too short (minimum 10 characters)${NC}"
    elif [ ${#subject} -gt 100 ]; then
        echo -e "${RED}Issue: Subject line too long (maximum 100 characters)${NC}"
    fi
    
    echo ""
    echo "Please update your commit message and try again."
    echo ""
    exit 1
fi

echo -e "${GREEN}✅ Commit message format is valid${NC}"
exit 0
EOF
    
    # Make hook executable
    chmod +x .git/hooks/commit-msg
    
    echo "✅ Commit-msg hook installed"
    echo ""
else
    echo "⏭️  Skipping commit-msg hook installation"
    echo ""
fi

# Display summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Setup Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 What's been configured:"
echo "  ✓ Commit message template (.gitmessage)"
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "  ✓ Local commit-msg validation hook"
fi
echo ""
echo "📝 How to use:"
echo "  1. Create commits normally: git commit"
echo "  2. Your editor will open with the template"
echo "  3. Follow the format: <type>(<scope>): <description>"
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "  4. The hook will validate your message before committing"
fi
echo ""
echo "💡 Quick examples:"
echo "  git commit -m 'feat: add user authentication'"
echo "  git commit -m 'fix(api): resolve timeout issue'"
echo "  git commit -m 'docs: update README with new endpoints'"
echo ""
echo "📚 For more information:"
echo "  • Conventional Commits: https://www.conventionalcommits.org/"
echo "  • Project docs: QUICKSTART.md and RELEASES.md"
echo ""
echo "🚀 Happy coding!"
echo ""
