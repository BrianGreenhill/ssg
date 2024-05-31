# Static Site Generator

Generate static site from markdown files. Use [Hugo](https://gohugo.io/) instead of this.

## Installation

```bash
go install github.com/briangreenhill/ssg@latest
```

## Usage

Choose a theme for your site. You can use the default theme or create your own.

```bash
# clone the default theme
git clone github.com/briangreenhill/ssg

# copy the default theme to your site
cp -r ../ssg/themes/default themes/<your-theme>
```

```bash
# create new site
ssg new

# theme: <your-theme>

# create new post
ssg post

# or use the archetype/post.md file to create a new post manually
```

Follow the prompts to create a new site. Once your new site is ready, you can start writing posts.

```bash
# generate the site files
ssg generate
```

This will generate the static site files in the `public` directory.

```bash
# watch for changes
ssg watch
```

This will watch for changes in the `posts` directory and regenerate the site files when changes are detected.

## Features
- [x] Markdown to HTML
- [x] Template rendering
- [x] Configuration
- [x] Watch mode
- [x] static assets
- [x] blog site generation
- [x] theme support
- [x] individual post pages
- [x] blog site links to posts
- [x] create initial user experience (theme, config, etc)
- [x] install themes
- [x] add new post command
- [ ] add draft mode for posts
