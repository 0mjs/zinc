# Zinc documentation

The Zinc documentation site is built with Astro and Starlight.

```sh
npm install
npm run dev
```

The local site runs at <http://localhost:4321>. Documentation lives in
`src/content/docs`, the homepage lives in `src/components/HomeHero.astro`, and
the sidebar is configured in `astro.config.mjs`.

Before publishing:

```sh
npm run check:site
npm run check:examples
npm run check:templ
```
