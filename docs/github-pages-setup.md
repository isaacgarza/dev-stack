# GitHub Pages Setup Guide

This guide covers setting up and troubleshooting the GitHub Pages documentation site for dev-stack.

## Quick Setup

### 1. Enable GitHub Pages
1. Go to **Settings** → **Pages** in your repository
2. Set **Source** to "GitHub Actions" (not "Deploy from a branch")
3. Save the configuration

### 2. Required Files
- ✅ `hugo.toml` - Hugo configuration
- ✅ `content/` - Documentation content
- ✅ `themes/PaperMod/` - Hugo theme (git submodule)
- ✅ `.github/workflows/pages.yml` - Deployment workflow

### 3. Initialize Theme
```bash
git submodule update --init --recursive
```

### 4. Deploy
Push changes to `main` branch. Site will be available at:
```
https://[username].github.io/dev-stack
```

## How It Works

### Deployment Process
1. **Trigger**: Push to `main` branch (content changes)
2. **Build**: Hugo generates static site with PaperMod theme
3. **CLI Docs**: Auto-generates CLI reference (or uses placeholder)
4. **Deploy**: Uploads to GitHub Pages

### Content Structure
```
content/
├── _index.md           # Homepage
├── getting-started.md  # Installation guide
├── usage.md           # Usage documentation
├── services.md        # Services reference
├── contributing.md    # Contributing guide
└── cli-reference/     # CLI docs (auto-generated)
    └── index.md
```

### Workflow Features
- Hugo Extended v0.121.0+
- Automatic CLI documentation generation
- Content validation before build
- Build error handling with detailed logs

## Local Development

### Setup
```bash
# Install Hugo Extended
brew install hugo  # macOS
sudo snap install hugo --channel=extended  # Linux

# Clone with submodules
git clone --recursive https://github.com/[username]/dev-stack.git
cd dev-stack

# Start development server
hugo server --buildDrafts
# Site available at http://localhost:1313
```

### Test Build
```bash
# Build site (same as CI)
hugo --gc --minify

# Check output
ls -la public/
```

## Troubleshooting

### Site Not Deploying
**Check:**
- Repository Settings → Pages → Source = "GitHub Actions"
- Actions tab for workflow status
- Repository is public (or has GitHub Pro for private)

**Fix:**
1. Enable GitHub Pages in repository settings
2. Check workflow logs for errors
3. Manually trigger workflow from Actions tab

### Build Failures
**Theme Issues:**
```bash
# Re-initialize theme
git submodule update --init --recursive

# Verify theme exists
ls themes/PaperMod/

# If missing, re-add
git submodule add https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod
```

**Content Issues:**
- Check frontmatter syntax in `.md` files:
```yaml
---
title: "Page Title"
description: "Page description"
---
```

**Hugo Configuration:**
```bash
# Test config
hugo config

# Test build locally
hugo server --buildDrafts
```

### Content Not Updating
**Causes:**
- Browser cache (hard refresh: Ctrl+F5)
- GitHub Pages cache (wait 10-15 minutes)
- Changes not in monitored paths (`content/`, `themes/`, `hugo.toml`)
- Pushed to wrong branch (must be `main`)

### CLI Documentation
The workflow attempts to build the CLI and generate real documentation. If this fails, it creates placeholder content. This is normal during development.

## Manual Deployment

If automatic deployment fails:
1. Go to **Actions** tab
2. Select **"Deploy Documentation to GitHub Pages"**
3. Click **"Run workflow"**
4. Select `main` branch and run

## Configuration

### Hugo Configuration (`hugo.toml`)
Key settings:
```toml
baseURL = "/"
theme = "PaperMod"
title = "dev-stack Documentation"

[params]
  env = "production"
```

### Required Repository Settings
- **Pages**: Source = "GitHub Actions"
- **Actions**: Workflows enabled
- **Repository**: Public (or GitHub Pro for private)

## Performance

### Build Targets
- **Build Time**: <4 minutes
- **Hugo Generation**: <1 minute
- **Site Loading**: <2 seconds

### Optimization
- Minified CSS/JS (automatic)
- Image optimization (Hugo processes images)
- CDN distribution (GitHub Pages)

## Maintenance

### Weekly
- Monitor workflow execution
- Check for broken links

### Monthly
- Update Hugo version if needed
- Check theme updates
- Review content accuracy

### Updates
**Hugo Version:**
```yaml
# In .github/workflows/pages.yml
- name: Setup Hugo
  uses: peaceiris/actions-hugo@v3
  with:
    hugo-version: "0.121.0"  # Update as needed
```

**Theme Updates:**
```bash
cd themes/PaperMod
git pull origin master
cd ../..
git add themes/PaperMod
git commit -m "feat: update PaperMod theme"
git push origin main
```

## Getting Help

### Resources
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Hugo Documentation](https://gohugo.io/documentation/)
- [PaperMod Theme Guide](https://github.com/adityatelange/hugo-PaperMod/wiki)

### Support
- **Issues**: Report bugs in repository
- **Hugo Community**: [Hugo Discourse](https://discourse.gohugo.io/)
- **GitHub Support**: For Pages-specific issues
