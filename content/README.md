---
title: "Documentation Development Guide"
description: "Guide for working with Hugo-based documentation"
---

# Documentation Development Guide

This guide explains how to work with the dev-stack documentation, which is built using [Hugo](https://gohugo.io/) with the [PaperMod theme](https://github.com/adityatelange/hugo-PaperMod).

## Why Hugo?

We chose Hugo because:
- **Written in Go** - Aligns perfectly with our tech stack
- **Single binary** - No external language dependencies (Python, Ruby, Node.js)
- **Lightning fast** - Builds in milliseconds
- **GitHub Pages compatible** - Works seamlessly with our deployment
- **Rich ecosystem** - Beautiful themes and extensive features

## Quick Start

### Prerequisites

- **Hugo Extended** (for local development)
- **Git** (for content management)
- **Go** (already required for the main project)

### Local Development

1. **Install Hugo (optional for local development):**
   ```bash
   # macOS
   brew install hugo
   
   # Linux
   wget https://github.com/gohugoio/hugo/releases/latest/download/hugo_extended_Linux-64bit.tar.gz
   tar -xzf hugo_extended_Linux-64bit.tar.gz
   sudo mv hugo /usr/local/bin/
   
   # Windows
   choco install hugo-extended
   ```

2. **Initialize submodules (first time only):**
   ```bash
   git submodule update --init --recursive
   ```

3. **Start the development server:**
   ```bash
   hugo server --buildDrafts --buildFuture
   ```
   
   Documentation will be available at http://localhost:1313

4. **Build for production:**
   ```bash
   hugo --minify
   ```

## Project Structure

```
dev-stack/
├── hugo.toml                 # Hugo configuration
├── content/                  # Documentation content
│   ├── _index.md            # Homepage
│   ├── getting-started.md   # Installation guide
│   ├── usage.md             # Usage documentation
│   ├── cli-reference.md     # CLI commands (auto-generated)
│   ├── services.md          # Available services
│   └── contributing.md      # Contribution guide
├── themes/
│   └── PaperMod/           # Theme submodule
└── public/                 # Generated site (ignored in git)
```

## Writing Documentation

### Front Matter

All pages must include Hugo front matter:

```yaml
---
title: "Page Title"
description: "Brief description for SEO and navigation"
weight: 10              # Optional: controls ordering in menus
draft: false            # Optional: set to true for drafts
---
```

### Markdown Features

Hugo supports GitHub-flavored Markdown plus additional features:

#### Code Blocks with Syntax Highlighting
```bash
# This will be syntax highlighted
dev-stack init go --name my-app
```

#### Shortcodes for Cross-References
```markdown
# Link to other pages
[Getting Started]({{< ref "/getting-started" >}})

# Link to specific sections
[Troubleshooting]({{< ref "/usage#troubleshooting" >}})
```

#### Admonitions (using blockquotes)
```markdown
> **💡 Pro Tip**: Use `dev-stack doctor` to check your system health.

> **⚠️ Warning**: This operation will remove all data.

> **ℹ️ Note**: GitHub Actions will handle the Hugo build automatically.
```

#### Tables
```markdown
| Command | Description |
|---------|-------------|
| `up` | Start services |
| `down` | Stop services |
```

### Auto-Generated Content

Some content is automatically generated:

- **CLI Reference** (`content/cli-reference/index.md`) - Generated from CLI help
- **Service Documentation** - May be generated from service YAML files

These files are marked with Hugo front matter and updated during the CI/CD process.

## Theme Customization

The PaperMod theme is added as a Git submodule. Customizations are done through:

1. **Hugo configuration** (`hugo.toml`)
2. **Custom CSS** (if needed, add to `assets/css/`)
3. **Custom layouts** (if needed, add to `layouts/`)

### Common Customizations

**Change theme colors:**
```toml
[params.assets]
  disableFingerprinting = false

# Add custom CSS
[[params.assets.css]]
  src = "css/custom.css"
```

**Modify navigation:**
```toml
[menu]
  [[menu.main]]
    identifier = "docs"
    name = "Documentation"
    url = "/getting-started/"
    weight = 10
```

## npm Scripts

The project includes npm scripts for common tasks:

```bash
# Build documentation
npm run docs:build

# Serve documentation locally
npm run docs:serve

# Clean build artifacts
npm run docs:clean

# Setup theme submodules
npm run docs:setup

# Lint markdown
npm run lint:md

# Fix markdown issues
npm run lint:md:fix
```

## CI/CD Integration

Documentation is automatically built and deployed via GitHub Actions:

### Workflow Overview
1. **Trigger**: Push to `main` branch
2. **Build**: Generate CLI docs from Go binary
3. **Hugo**: Build static site with Hugo
4. **Deploy**: Deploy to GitHub Pages

### GitHub Actions Workflow
```yaml
- name: Setup Hugo
  uses: peaceiris/actions-hugo@v2
  with:
    hugo-version: 'latest'
    extended: true

- name: Build with Hugo
  run: hugo --minify

- name: Deploy to GitHub Pages
  uses: actions/deploy-pages@v2
```

## Content Guidelines

### Style Guide

1. **Use clear, concise language**
2. **Include practical examples** for all features
3. **Structure content** with descriptive headings
4. **Cross-reference** related documentation
5. **Keep it up-to-date** with the latest features

### File Naming

- Use lowercase with hyphens: `getting-started.md`
- Be descriptive but concise
- Match menu structure where possible

### Content Organization

1. **Start with overview** - What is this feature?
2. **Prerequisites** - What do users need?
3. **Step-by-step instructions** - Clear, numbered steps
4. **Examples** - Real-world usage scenarios
5. **Troubleshooting** - Common issues and solutions
6. **References** - Links to related docs

## Local Development Tips

### Live Reload
Hugo's dev server automatically reloads when you save changes:
```bash
hugo server --buildDrafts --navigateToChanged
```

### Draft Content
Mark pages as drafts during development:
```yaml
---
title: "Work in Progress"
draft: true
---
```

### Content Organization
Use Hugo's content organization features:
- **Sections**: Organize related pages
- **Taxonomies**: Add tags and categories
- **Menus**: Control navigation structure

## Troubleshooting

### Common Issues

**Hugo not found:**
```bash
# Check if Hugo is installed
hugo version

# Install if missing (see prerequisites)
```

**Theme not loading:**
```bash
# Update submodules
git submodule update --init --recursive

# Or setup via npm
npm run docs:setup
```

**Build fails:**
```bash
# Check for markdown errors
npm run lint:md

# Build with verbose output
hugo --verbose
```

**Links not working:**
```bash
# Use Hugo shortcodes for internal links
[Getting Started]({{< ref "/getting-started" >}})

# Not raw markdown links
[Getting Started](/getting-started/)
```

### Development Workflow

1. **Make changes** to content files
2. **Preview locally** with `hugo server`
3. **Test build** with `hugo --minify`
4. **Lint content** with `npm run lint:md`
5. **Commit and push** to trigger deployment

## Contributing to Documentation

### Making Changes

1. **Edit content** in the `content/` directory
2. **Test locally** with Hugo dev server
3. **Check links** and formatting
4. **Submit pull request**

### Adding New Pages

1. **Create markdown file** in appropriate section
2. **Add front matter** with title and description
3. **Update navigation** if needed (in `hugo.toml`)
4. **Test build** and navigation

### Updating Auto-Generated Content

Auto-generated content (like CLI reference) is updated automatically by CI/CD. To modify:

1. **Update source** (CLI help text, service definitions)
2. **Test generation** locally if possible
3. **Verify output** in CI/CD build

## Performance

Hugo builds are extremely fast:
- **Development**: ~50ms incremental builds
- **Production**: ~200ms full site builds
- **GitHub Pages**: ~30s including deployment

## SEO and Analytics

The site includes:
- **SEO meta tags** via Hugo's built-in SEO
- **Sitemap generation** for search engines
- **RSS feeds** for content updates
- **Social media** integration

To add analytics:
```toml
[params]
  googleAnalytics = "G-XXXXXXXXXX"
```

## Resources

- **[Hugo Documentation](https://gohugo.io/documentation/)**
- **[PaperMod Theme](https://github.com/adityatelange/hugo-PaperMod)**
- **[Markdown Guide](https://www.markdownguide.org/)**
- **[Hugo Shortcodes](https://gohugo.io/content-management/shortcodes/)**

---

> **Zero Dependencies**: This documentation system requires only Go and Hugo - no Python, Ruby, or Node.js runtime dependencies!