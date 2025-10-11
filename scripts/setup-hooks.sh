#!/bin/bash
set -e

echo "🪝 Setting up Git hooks for dev-stack..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e
task pre-commit
EOF

# Pre-push hook  
cat > .git/hooks/pre-push << 'EOF'
#!/bin/bash
set -e
task pre-push
EOF

# Make hooks executable
chmod +x .git/hooks/pre-commit .git/hooks/pre-push

echo "✅ Git hooks installed successfully!"
echo ""
echo "Available commands:"
echo "  task pre-commit  - Run pre-commit checks manually"
echo "  task pre-push    - Run pre-push checks manually"
echo ""
echo "To skip hooks temporarily:"
echo "  git commit --no-verify"
echo "  git push --no-verify"
