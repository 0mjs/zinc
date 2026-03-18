module.exports = {
  docs: [
    "intro",
    {
      type: "category",
      label: "Guide",
      link: {
        type: "generated-index",
        title: "Guide",
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
      label: "API",
      link: {
        type: "generated-index",
        title: "API Reference",
        description: "High-signal reference pages for Zinc’s primary public types.",
      },
      items: [
        "api/app",
        "api/group",
        "api/context",
        "api/binding",
        "api/config",
        "api/errors",
      ],
    },
    {
      type: "category",
      label: "Middleware",
      link: {
        type: "generated-index",
        title: "Middleware",
        description: "Overview of Zinc’s first-party middleware package.",
      },
      items: ["middleware/overview"],
    },
    {
      type: "category",
      label: "Extra",
      link: {
        type: "generated-index",
        title: "Extra",
        description: "Benchmarks, design notes, and practical FAQ material.",
      },
      items: ["extra/benchmarks", "extra/faq"],
    },
  ],
};
