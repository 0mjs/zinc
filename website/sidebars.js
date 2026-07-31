module.exports = {
  docs: [
    {
      type: "doc",
      id: "intro",
      label: "Welcome",
    },
    {
      type: "category",
      label: "Getting Started",
      link: {
        type: "generated-index",
        title: "Getting Started",
        description: "Install Zinc, run a tiny app, and learn the shape of a normal handler.",
      },
      items: [
        "getting-started/installation",
        "getting-started/quick-start",
        "getting-started/first-route",
      ],
    },
    {
      type: "category",
      label: "Core Guides",
      link: {
        type: "generated-index",
        title: "Guide",
        description: "Learn Zinc concepts in the order you will use them in an application.",
      },
      items: [
        "guide/routing",
        "guide/groups-and-middleware",
        "guide/context",
        "guide/binding",
        "guide/responses-and-rendering",
        "guide/static-files",
        "guide/errors",
        "guide/configuration",
      ],
    },
    {
      type: "category",
      label: "Middleware",
      link: {
        type: "doc",
        id: "middleware/overview",
      },
      items: [
        "middleware/overview",
        {
          type: "category",
          label: "Security",
          items: [
            "middleware/cors",
            "middleware/csrf",
            "middleware/secure",
            "middleware/header-guards",
          ],
        },
        {
          type: "category",
          label: "Authentication",
          items: [
            "middleware/jwt",
            "middleware/basic-auth",
            "middleware/key-auth",
            "middleware/casbin-auth",
            "middleware/session",
          ],
        },
        {
          type: "category",
          label: "Observability",
          items: [
            "middleware/request-logger",
            "middleware/request-id",
            "middleware/prometheus",
            "middleware/jaeger",
            "middleware/pprof",
            "middleware/body-dump",
          ],
        },
        {
          type: "category",
          label: "Traffic Control",
          items: [
            "middleware/rate-limiter",
            "middleware/body-limit",
            "middleware/context-timeout",
            "middleware/recover",
          ],
        },
        {
          type: "category",
          label: "Transport",
          items: [
            "middleware/gzip",
            "middleware/decompress",
            "middleware/method-override",
            "middleware/trailing-slash",
            "middleware/rewrite",
            "middleware/redirect",
            "middleware/proxy",
          ],
        },
        {
          type: "category",
          label: "Assets and Utility",
          items: [
            "middleware/static",
            "middleware/utility",
          ],
        },
      ],
    },
    {
      type: "category",
      label: "Cookbook",
      link: {
        type: "doc",
        id: "cookbook/cookbook",
      },
      items: [
        "cookbook/templated-html-js-page",
        "cookbook/sqlite-crud-api",
        "cookbook/scheduled-jobs",
        "cookbook/templ-ui",
      ],
    },
    {
      type: "category",
      label: "API Reference",
      link: {
        type: "generated-index",
        title: "API Reference",
        description: "Reference pages for Zinc's main public types and extension points.",
      },
      items: [
        "api/app",
        "api/group",
        "api/context",
        "api/binding",
        "api/config",
        "api/errors",
        "api/response-writer",
      ],
    },
    {
      type: "category",
      label: "About Zinc",
      link: {
        type: "generated-index",
        title: "Project",
        description: "Benchmarks, FAQ, and project-level notes.",
      },
      items: ["extra/benchmarks", "extra/faq"],
    },
  ],
};
