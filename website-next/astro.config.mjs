import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

const page = (label, link) => ({ label, link });

export default defineConfig({
  site: "https://zinc.carbonsoft.sh",
  integrations: [
    starlight({
      title: "Zinc",
      logo: {
        src: "./src/assets/zinc.png",
        replacesTitle: true,
      },
      favicon: "/zinc.png",
      customCss: ["./src/styles/zinc.css"],
      editLink: {
        baseUrl: "https://github.com/0mjs/zinc/edit/dev/website-next/",
      },
      lastUpdated: true,
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/0mjs/zinc",
        },
      ],
      components: {
        Footer: "./src/components/Footer.astro",
      },
      head: [
        {
          tag: "script",
          content:
            "try{if(!localStorage.getItem('starlight-theme')){localStorage.setItem('starlight-theme','dark');document.documentElement.dataset.theme='dark';}}catch(e){document.documentElement.dataset.theme='dark';}",
        },
        {
          tag: "link",
          attrs: { rel: "preconnect", href: "https://fonts.googleapis.com" },
        },
        {
          tag: "link",
          attrs: {
            rel: "preconnect",
            href: "https://fonts.gstatic.com",
            crossorigin: true,
          },
        },
        {
          tag: "link",
          attrs: {
            rel: "stylesheet",
            href: "https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap",
          },
        },
      ],
      sidebar: [
        {
          label: "Guide",
          items: [
            page("Quickstart", "/guide/quickstart/"),
            page("Installation", "/guide/installation/"),
            page("Routing", "/guide/routing/"),
            page("Context", "/guide/context/"),
            page("Binding", "/guide/binding/"),
            page("Error Handling", "/guide/error-handling/"),
            page("Request", "/guide/request/"),
            page("Response", "/guide/response/"),
            page("Serving Static Files", "/guide/static-files/"),
            page("Templates", "/guide/templates/"),
            page("Cookies", "/guide/cookies/"),
            page("Customization", "/guide/customization/"),
            page("Testing", "/guide/testing/"),
            page("IP Address", "/guide/ip-address/"),
          ],
        },
        {
          label: "Middleware",
          items: [
            page("Basic Auth", "/middleware/basic-auth/"),
            page("Body Dump", "/middleware/body-dump/"),
            page("Body Limit", "/middleware/body-limit/"),
            page("Casbin Auth", "/middleware/casbin-auth/"),
            page("Context Timeout", "/middleware/context-timeout/"),
            page("CORS", "/middleware/cors/"),
            page("CSRF", "/middleware/csrf/"),
            page("Decompress", "/middleware/decompress/"),
            page("Gzip", "/middleware/gzip/"),
            page("JWT", "/middleware/jwt/"),
            page("Key Auth", "/middleware/key-auth/"),
            page("Request Logger", "/middleware/request-logger/"),
            page("Method Override", "/middleware/method-override/"),
            page("OpenTelemetry", "/middleware/open-telemetry/"),
            page("Prometheus", "/middleware/prometheus/"),
            page("Proxy", "/middleware/proxy/"),
            page("Rate Limiter", "/middleware/rate-limiter/"),
            page("Recover", "/middleware/recover/"),
            page("Redirect", "/middleware/redirect/"),
            page("Request ID", "/middleware/request-id/"),
            page("Rewrite", "/middleware/rewrite/"),
            page("Secure", "/middleware/secure/"),
            page("Session", "/middleware/session/"),
            page("Static", "/middleware/static/"),
            page("Trailing Slash", "/middleware/trailing-slash/"),
            page("Header Guards", "/middleware/header-guards/"),
            page("Jaeger", "/middleware/jaeger/"),
            page("pprof", "/middleware/pprof/"),
            page("Utility", "/middleware/utility/"),
          ],
        },
        {
          label: "Cookbook",
          items: [
            page("Hello World", "/cookbook/hello-world/"),
            page("CRUD", "/cookbook/crud/"),
            page("Automatic TLS", "/cookbook/auto-tls/"),
            page("CORS", "/cookbook/cors/"),
            page("Embed Resources", "/cookbook/embed-resources/"),
            page("File Download", "/cookbook/file-download/"),
            page("File Upload", "/cookbook/file-upload/"),
            page("Graceful Shutdown", "/cookbook/graceful-shutdown/"),
            page("HTTP/2 Server", "/cookbook/http2/"),
            page("HTTP/2 Server Push", "/cookbook/http2-server-push/"),
            page("JWT", "/cookbook/jwt/"),
            page("Custom Middleware", "/cookbook/middleware/"),
            page("JSONP", "/cookbook/jsonp/"),
            page("Server-Sent Events", "/cookbook/sse/"),
            page("Streaming Response", "/cookbook/streaming-response/"),
            page("WebSocket", "/cookbook/websocket/"),
            page("Subdomain", "/cookbook/subdomain/"),
            page("Timeout", "/cookbook/timeout/"),
            page("Reverse Proxy", "/cookbook/reverse-proxy/"),
            page("Load Balancing", "/cookbook/load-balancing/"),
            page("SQLite CRUD API", "/cookbook/sqlite-crud-api/"),
            page("Scheduled Jobs", "/cookbook/scheduled-jobs/"),
            page("Templ UI", "/cookbook/templ-ui/"),
            page("Templated HTML + JS", "/cookbook/templated-html-js-page/"),
          ],
        },
      ],
      expressiveCode: {
        themes: ["tokyo-night", "catppuccin-latte"],
      },
    }),
  ],
});
