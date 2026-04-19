// @ts-check

const { nordDark, nordLight } = require("./src/prism/nord");

const config = {
  title: "Zinc",
  tagline: "Fast, explicit Go APIs on net/http.",
  favicon: "img/z_logo.png",

  url: "https://zinc.carbonsoft.sh",
  baseUrl: "/",

  organizationName: "0mjs",
  projectName: "zinc",

  onBrokenLinks: "throw",
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: "warn",
    },
  },

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/0mjs/zinc/tree/main/website/",
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
        },
        blog: false,
        pages: false,
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],

  themes: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      {
        hashed: true,
        indexDocs: true,
        indexBlog: false,
        indexPages: false,
        docsRouteBasePath: "/",
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      },
    ],
  ],

  themeConfig: {
    image: "img/z_logo.png",
    navbar: {
      title: "Zinc",
      logo: {
        alt: "Zinc",
        src: "img/z_logo.png",
      },
      items: [
        { to: "/", label: "Docs", position: "left" },
        { to: "/getting-started/quick-start", label: "Quick Start", position: "left" },
        { to: "/guide/routing", label: "Guide", position: "left" },
        { href: "https://pkg.go.dev/github.com/0mjs/zinc", label: "pkg.go.dev", position: "right" },
        { href: "https://github.com/0mjs/zinc", label: "GitHub", position: "right" },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Docs",
          items: [
            { label: "Welcome", to: "/" },
            { label: "Quick Start", to: "/getting-started/quick-start" },
            { label: "Routing", to: "/guide/routing" },
            { label: "Binding", to: "/guide/binding" },
            { label: "App API", to: "/api/app" },
            { label: "Cookbook", to: "/cookbook" },
          ],
        },
        {
          title: "Project",
          items: [
            { label: "GitHub", href: "https://github.com/0mjs/zinc" },
            { label: "pkg.go.dev", href: "https://pkg.go.dev/github.com/0mjs/zinc" },
            { label: "Benchmarks", to: "/extra/benchmarks" },
          ],
        },
      ],
      copyright: `Copyright ${new Date().getFullYear()} Zinc.`,
    },
    prism: {
      theme: nordLight,
      darkTheme: nordDark,
      additionalLanguages: ["bash", "go", "json", "toml", "yaml"],
    },
    colorMode: {
      defaultMode: "light",
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    announcementBar: {
      id: "docs-preview",
      content: "Zinc is pre-1.0. The docs track the current API and call out behavior directly.",
      isCloseable: true,
    },
  },
};

module.exports = config;
