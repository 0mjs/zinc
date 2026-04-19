const nordDark = {
  plain: {
    color: "#d8dee9",
    backgroundColor: "#2e3440",
  },
  styles: [
    {
      types: ["comment", "prolog", "doctype", "cdata"],
      style: { color: "#616e88", fontStyle: "italic" },
    },
    {
      types: ["punctuation"],
      style: { color: "#eceff4" },
    },
    {
      types: ["property", "tag", "constant", "symbol", "deleted"],
      style: { color: "#81a1c1" },
    },
    {
      types: ["boolean", "number"],
      style: { color: "#b48ead" },
    },
    {
      types: ["selector", "attr-name", "string", "char", "inserted"],
      style: { color: "#a3be8c" },
    },
    {
      types: ["builtin"],
      style: { color: "#8fbcbb" },
    },
    {
      types: ["operator", "entity", "url"],
      style: { color: "#81a1c1" },
    },
    {
      types: ["variable"],
      style: { color: "#d8dee9" },
    },
    {
      types: ["atrule", "attr-value", "keyword"],
      style: { color: "#81a1c1" },
    },
    {
      types: ["function", "class-name"],
      style: { color: "#88c0d0" },
    },
    {
      types: ["regex", "important"],
      style: { color: "#ebcb8b" },
    },
    {
      types: ["important", "bold"],
      style: { fontWeight: "bold" },
    },
    {
      types: ["italic"],
      style: { fontStyle: "italic" },
    },
  ],
};

const nordLight = {
  plain: {
    color: "#2e3440",
    backgroundColor: "#eceff4",
  },
  styles: [
    {
      types: ["comment", "prolog", "doctype", "cdata"],
      style: { color: "#60728a", fontStyle: "italic" },
    },
    {
      types: ["punctuation"],
      style: { color: "#3b4252" },
    },
    {
      types: ["property", "tag", "constant", "symbol", "deleted"],
      style: { color: "#5e81ac" },
    },
    {
      types: ["boolean", "number"],
      style: { color: "#b48ead" },
    },
    {
      types: ["selector", "attr-name", "string", "char", "inserted"],
      style: { color: "#8a9f6c" },
    },
    {
      types: ["builtin"],
      style: { color: "#5e9b9e" },
    },
    {
      types: ["operator", "entity", "url"],
      style: { color: "#5e81ac" },
    },
    {
      types: ["variable"],
      style: { color: "#2e3440" },
    },
    {
      types: ["atrule", "attr-value", "keyword"],
      style: { color: "#5e81ac" },
    },
    {
      types: ["function", "class-name"],
      style: { color: "#3b6f8c" },
    },
    {
      types: ["regex", "important"],
      style: { color: "#b58900" },
    },
    {
      types: ["bold"],
      style: { fontWeight: "bold" },
    },
    {
      types: ["italic"],
      style: { fontStyle: "italic" },
    },
  ],
};

module.exports = { nordDark, nordLight };
