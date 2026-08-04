import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const websiteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repoRoot = path.resolve(websiteRoot, "..");
const docsRoot = path.join(websiteRoot, "src/content/docs");
const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "zinc-doc-examples-"));

function filesBelow(root) {
  const files = [];
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const filename = path.join(root, entry.name);
    if (entry.isDirectory()) files.push(...filesBelow(filename));
    else if (/\.mdx?$/.test(filename)) files.push(filename);
  }
  return files;
}

function run(command, args, cwd) {
  execFileSync(command, args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
}

let checked = 0;
let skipped = 0;

try {
  for (const filename of filesBelow(docsRoot)) {
    const source = fs.readFileSync(filename, "utf8");
    const blocks = source.matchAll(/```go([^\n]*)\n(package main\n[\s\S]*?)```/g);

    let index = 0;
    for (const block of blocks) {
      index++;
      if (/\bcheck=false\b/.test(block[1])) {
        skipped++;
        continue;
      }

      const label = `${path.relative(docsRoot, filename)}#${index}`;
      const exampleRoot = path.join(tempRoot, `example-${checked + 1}`);
      fs.mkdirSync(exampleRoot, { recursive: true });
      fs.writeFileSync(path.join(exampleRoot, "main.go"), block[2]);

      if (block[2].includes("//go:embed public")) {
        fs.mkdirSync(path.join(exampleRoot, "public"));
        fs.writeFileSync(path.join(exampleRoot, "public/index.html"), "<!doctype html><title>Zinc</title>\n");
      }

      try {
        run("go", ["mod", "init", "docscheck"], exampleRoot);
        run("go", ["mod", "edit", "-go=1.25.0"], exampleRoot);
        run("go", ["mod", "edit", `-replace=github.com/0mjs/zinc=${repoRoot}`], exampleRoot);
        run("go", ["mod", "tidy"], exampleRoot);
        run("go", ["test", "."], exampleRoot);
      } catch (error) {
        const stderr = error.stderr?.toString().trim();
        const stdout = error.stdout?.toString().trim();
        console.error(`Failed to compile ${label}`);
        if (stderr) console.error(stderr);
        if (stdout) console.error(stdout);
        process.exit(1);
      }

      checked++;
    }
  }
} finally {
  fs.rmSync(tempRoot, { recursive: true, force: true });
}

console.log(`Compiled ${checked} complete Go documentation examples; skipped ${skipped} generated-code example.`);
