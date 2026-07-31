// @ts-check

const { nordDark } = require("./src/prism/nord");

const config = {
  title: "Zinc",
  tagline: "Fast, explicit Go APIs on net/http.",
  favicon: "img/zinc.png",

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
    image: "img/zinc.png",
    navbar: {
      title: "Zinc",
      logo: {
        alt: "Zinc",
        src: "img/zinc.png",
      },
      items: [
        { to: "/getting-started/quick-start", label: "Quick Start", position: "left" },
        { to: "/guide/routing", label: "Guides", position: "left" },
        { to: "/middleware/overview", label: "Middleware", position: "left" },
        { to: "/api/app", label: "Reference", position: "left" },
        {
          href: "https://pkg.go.dev/github.com/0mjs/zinc",
          label: "pkg.go.dev",
          position: "right",
          className: "navbar__link--package",
        },
        {
          href: "https://github.com/0mjs/zinc",
          label: "GitHub",
          position: "right",
          className: "navbar__link--github",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Learn",
          items: [
            { label: "Welcome", to: "/" },
            { label: "Quick Start", to: "/getting-started/quick-start" },
            { label: "Routing", to: "/guide/routing" },
            { label: "Binding", to: "/guide/binding" },
            { label: "Cookbook", to: "/cookbook" },
          ],
        },
        {
          title: "Reference",
          items: [
            { label: "App", to: "/api/app" },
            { label: "Context", to: "/api/context" },
            { label: "Configuration", to: "/api/config" },
            { label: "Errors", to: "/api/errors" },
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
      theme: nordDark,
      darkTheme: nordDark,
      additionalLanguages: ["bash", "go", "json", "toml", "yaml"],
    },
    colorMode: {
      defaultMode: "light",
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
  },
};

module.exports = config;
