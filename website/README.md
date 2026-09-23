# Zinc documentation

The Zinc documentation site is built with Astro and Starlight.

```sh
npm install
npm run dev
```

The local site runs at <http://localhost:4321>. Documentation lives in
`src/content/docs`, the homepage lives in `src/components/HomeHero.astro`, and
the sidebar is configured in `astro.config.mjs`.

The theme lives in `src/styles/zinc.css`, and code blocks use the Nord-derived
`src/themes/zinc-frost.json` theme. The spangle texture and favicons in `public/`
are generated; rebuild them with `npm run build:brand`.

Before publishing:

```sh
npm run check:site
npm run check:examples
npm run check:templ
```
