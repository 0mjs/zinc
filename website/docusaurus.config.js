// @ts-check

const { themes } = require("prism-react-renderer");

const config = {
  title: "Zinc",
  tagline: "Fast, explicit API docs for a fast, explicit Go framework.",
  favicon: "img/logo.svg",

  url: "https://zinc.carbonsoft.sh",
  baseUrl: "/",

  organizationName: "0mjs",
  projectName: "zinc",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

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
        },
        blog: false,
        pages: false,
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],

  themeConfig: {
    navbar: {
      title: "Zinc",
      logo: {
        alt: "Zinc",
        src: "img/logo.svg",
      },
      items: [
        { to: "/", label: "Home", position: "left" },
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
      theme: themes.github,
      darkTheme: themes.dracula,
    },
    colorMode: {
      defaultMode: "dark",
      disableSwitch: false,
      respectPrefersColorScheme: false,
    },
    announcementBar: {
      id: "docs-preview",
      content: "Zinc docs are live in-repo. Expect rapid improvements while the API stabilizes toward 0.1.x.",
      backgroundColor: "#15161a",
      textColor: "#eceef0",
      isCloseable: true,
    },
  },
};

module.exports = config;
