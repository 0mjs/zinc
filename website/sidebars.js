module.exports = {
  docs: [
    {
      type: "doc",
      id: "intro",
      label: "👋 Welcome",
    },
    {
      type: "category",
      label: "🛠️ API",
      link: {
        type: "generated-index",
        title: "🛠️ API",
        description: "Reference pages for Zinc’s primary public types and runtime surface.",
      },
      items: [
        {
          type: "doc",
          id: "api/app",
          label: "🚀 App",
        },
        {
          type: "doc",
          id: "api/group",
          label: "👥 Group",
        },
        {
          type: "doc",
          id: "api/context",
          label: "🧠 Context",
        },
        {
          type: "doc",
          id: "api/binding",
          label: "📎 Bind",
        },
        {
          type: "doc",
          id: "api/config",
          label: "⚙️ Config",
        },
        {
          type: "doc",
          id: "api/errors",
          label: "🚨 Errors",
        },
      ],
    },
    {
      type: "category",
      label: "🧭 Guide",
      link: {
        type: "generated-index",
        title: "🧭 Guide",
        description: "Start here for Zinc concepts, patterns, and practical examples.",
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
      label: "🛡️ Middleware",
      link: {
        type: "generated-index",
        title: "🛡️ Middleware",
        description: "Detailed docs for Zinc’s first-party middleware package.",
      },
      items: [
        "middleware/overview",
        "middleware/cors",
        "middleware/csrf",
        "middleware/jwt",
        "middleware/basic-auth",
        "middleware/request-logger",
        "middleware/body-limit",
        "middleware/body-dump",
        "middleware/context-timeout",
        "middleware/rate-limiter",
      ],
    },
    {
      type: "category",
      label: "⚡ Extra",
      link: {
        type: "generated-index",
        title: "⚡ Extra",
        description: "Benchmarks, design notes, and practical FAQ material.",
      },
      items: ["extra/benchmarks", "extra/faq"],
    },
  ],
};
