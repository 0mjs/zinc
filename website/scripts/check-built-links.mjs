import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const siteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const distRoot = path.join(siteRoot, "dist");

function filesBelow(root, accept) {
  const files = [];
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const filename = path.join(root, entry.name);
    if (entry.isDirectory()) files.push(...filesBelow(filename, accept));
    else if (accept(filename)) files.push(filename);
  }
  return files;
}

function normalizeRoute(route) {
  const normalized = path.posix.normalize(route.split(/[?#]/)[0]).replace(/\/index\.html$/, "").replace(/\.html$/, "").replace(/\/$/, "");
  return normalized || "/";
}

function routeForFile(filename) {
  return normalizeRoute(`/${path.relative(distRoot, filename).replace(/\\/g, "/")}`);
}

if (!fs.existsSync(distRoot)) {
  console.error("website/dist does not exist; run npm run build first");
  process.exit(1);
}

const htmlFiles = filesBelow(distRoot, (filename) => filename.endsWith(".html"));
const routes = new Set(htmlFiles.map(routeForFile));
const failures = [];
let checked = 0;

for (const filename of htmlFiles) {
  const sourceRoute = routeForFile(filename);
  const html = fs.readFileSync(filename, "utf8");

  if (path.basename(filename) !== "404.html" && /<title>404(?:\s|\|)/.test(html)) {
    failures.push(`${path.relative(distRoot, filename)}: generated the 404 page instead of its documentation entry`);
  }

  for (const match of html.matchAll(/\bhref=["']([^"']+)["']/g)) {
    const rawHref = match[1];
    if (/^(?:https?:|mailto:|tel:|data:|javascript:|#)/.test(rawHref)) continue;

    const cleanHref = rawHref.split(/[?#]/)[0];
    if (!cleanHref || /^\/(?:_astro|pagefind)\//.test(cleanHref)) continue;
    if (/\.[a-z0-9]+$/i.test(cleanHref) && !/\.html$/i.test(cleanHref)) continue;

    const target = cleanHref.startsWith("/")
      ? normalizeRoute(cleanHref)
      : normalizeRoute(path.posix.resolve(`${sourceRoute}/`, cleanHref));
    checked += 1;
    if (!routes.has(target)) {
      failures.push(`${path.relative(distRoot, filename)}: ${rawHref} resolves to missing ${target}`);
    }
  }
}

if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(`Checked ${checked} generated internal links across ${htmlFiles.length} HTML pages.`);
