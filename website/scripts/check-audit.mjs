// Runs npm audit and fails on any advisory, at any severity, except those
// listed below. Each exception says why it can't affect this site and when to
// look at it again, so it can't quietly outlive its reason.
import { execFileSync } from "node:child_process";

const exceptions = {
  // postcss-selector-parser: quadratic parsing of a long flat selector can
  // exhaust CPU. Starlight pulls it in through postcss-nested, which accepts
  // only version 6, and the fix is 7.1.6. It runs at build time on the site's
  // own CSS, never on a visitor's input.
  "GHSA-rj75-hqrm-r3gf": { review: "2026-12-01" },
};

let report;
try {
  report = execFileSync("npm", ["audit", "--json"], { encoding: "utf8", stdio: ["ignore", "pipe", "inherit"] });
} catch (err) {
  report = err.stdout; // npm audit exits non-zero when it finds anything
}
const { vulnerabilities = {} } = JSON.parse(report);

const found = new Map(); // advisory id -> package and severity
for (const [name, v] of Object.entries(vulnerabilities)) {
  for (const via of v.via) {
    if (typeof via === "string") continue; // a dependency path; the advisory is listed under its own package
    const id = via.url?.split("/").pop() ?? `${name}: ${via.title}`;
    found.set(id, `${name} (${via.severity}): ${via.title}`);
  }
}

const today = new Date().toISOString().slice(0, 10);
const problems = [];
for (const [id, what] of found) {
  const exception = exceptions[id];
  if (!exception) problems.push(`${id} ${what}`);
  else if (exception.review < today) problems.push(`${id} ${what}: its exception was due for review on ${exception.review}`);
}
for (const id of Object.keys(exceptions)) {
  if (!found.has(id)) problems.push(`${id} no longer appears in npm audit; remove its exception`);
}

if (problems.length > 0) {
  console.error(`npm audit:\n  ${problems.join("\n  ")}`);
  process.exit(1);
}
console.log(`npm audit: no advisories beyond ${Object.keys(exceptions).length} reviewed exception(s).`);
