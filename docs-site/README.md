# Documentation Site

This directory contains the Hugo-based documentation site for the dev-stack project.

## Structure

```
docs-site/
├── config/           # Hugo configuration files
│   └── _default/     # Default configuration
│       ├── hugo.toml # Main Hugo configuration
│       ├── params.toml # Site parameters
│       ├── module.toml # Hugo modules
│       └── menus/    # Navigation menus
├── content/          # Markdown content files
├── layouts/          # Custom Hugo layouts (optional)
├── static/           # Static assets (images, etc.)
├── assets/           # Hugo asset pipeline files
├── themes/           # Hugo themes (Doks theme)
├── resources/        # Hugo resource cache
├── public/           # Generated site output (ignored in git)
├── package.json      # Node.js dependencies for Hugo theme
└── node_modules/     # Node.js modules (ignored in git)
```

## Getting Started

### Prerequisites

- Hugo Extended (latest version)
- Node.js 18+
- npm

### Installation

From the project root directory:

```bash
# Install Node.js dependencies for the documentation site
cd docs-site
npm install

# Initialize Hugo themes (if needed)
npm run docs:setup
# Or manually: git submodule update --init --recursive
```

### Development

```bash
# From project root - build docs and start Hugo dev server
task hugo-serve

# Or manually from docs-site directory
cd docs-site
npm run docs:dev
```

### Building

```bash
# From project root - generate docs and build site
task hugo-build

# Or manually from docs-site directory
cd docs-site
npm run docs:build
```

## Available Commands

### From Project Root (using Taskfile)

- `task docs` - Generate CLI documentation only
- `task docs-hugo` - Generate docs and sync to Hugo content
- `task hugo-build` - Build the complete documentation site
- `task hugo-serve` - Start development server
- `task hugo-clean` - Clean build artifacts
- `task validate-docs` - Validate Hugo configuration and content

### From docs-site Directory (using npm)

- `npm run docs:build` - Build production site
- `npm run docs:serve` - Serve built site locally
- `npm run docs:dev` - Start development server with drafts
- `npm run docs:clean` - Clean build output
- `npm run docs:setup` - Initialize git submodules for themes

## Configuration

The main Hugo configuration is in `config/_default/hugo.toml`. Key settings:

- **Content Directory**: `content/` (contains all Markdown files)
- **Theme**: Doks (modern documentation theme)
- **Base URL**: Configured for deployment
- **Content Types**: Supports docs, blog posts, and CLI reference

## Content Organization

Content in the `content/` directory follows this structure:

- `_index.md` - Homepage content
- `*.md` - Individual documentation pages
- Subdirectories for organized sections

## Theme Customization

The Doks theme provides:

- Responsive design
- Search functionality
- Navigation menus
- Syntax highlighting
- SEO optimization

Custom layouts can be added in the `layouts/` directory to override theme defaults.

## Deployment

The site is built to the `public/` directory and can be deployed to:

- GitHub Pages
- Netlify
- Vercel
- Any static hosting service

## Troubleshooting

1. **Hugo build fails**: Check `hugo config` for configuration errors
2. **Theme issues**: Run `npm run docs:setup` to update theme submodules (located in `docs-site/themes/doks`)
3. **Missing content**: Ensure content files have proper front matter
4. **Development server issues**: Try `npm run docs:clean` then rebuild
5. **Submodule issues**: Run `git submodule update --init --recursive` from project root

For more help, see the main project documentation or run `task help`.
