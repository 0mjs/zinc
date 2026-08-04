import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const websiteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repoRoot = path.resolve(websiteRoot, "..");
const recipe = path.join(websiteRoot, "src/content/docs/cookbook/templ-ui.md");
const source = fs.readFileSync(recipe, "utf8");
const templ = source.match(/```templ[^\n]*\n([\s\S]*?)```/)?.[1];
const app = source.match(/```go[^\n]*\bcheck=false\b[^\n]*\n(package main\n[\s\S]*?)```/)?.[1];

if (!templ || !app) {
  console.error("Templ UI recipe must contain its complete Templ component and check=false Go application.");
  process.exit(1);
}

const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "zinc-templ-example-"));

function run(command, args) {
  execFileSync(command, args, {
    cwd: tempRoot,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
}

try {
  fs.mkdirSync(path.join(tempRoot, "views"));
  fs.writeFileSync(path.join(tempRoot, "views/home.templ"), templ);
  fs.writeFileSync(path.join(tempRoot, "main.go"), app);
  fs.mkdirSync(path.join(tempRoot, "public"));

  run("go", ["mod", "init", "zinc-templ"]);
  run("go", ["mod", "edit", "-go=1.25.0"]);
  run("go", ["mod", "edit", `-replace=github.com/0mjs/zinc=${repoRoot}`]);
  run("go", ["get", "github.com/0mjs/zinc"]);
  run("go", ["get", "github.com/a-h/templ"]);
  run("go", ["get", "github.com/templui/templui@latest"]);
  run("go", ["get", "-tool", "github.com/a-h/templ/cmd/templ@latest"]);
  run("go", ["tool", "templ", "generate"]);
  run("go", ["mod", "tidy"]);
  run("go", ["test", "."]);
} catch (error) {
  console.error("Templ UI cookbook workflow failed.");
  if (error.stderr?.length) console.error(error.stderr.toString().trim());
  if (error.stdout?.length) console.error(error.stdout.toString().trim());
  process.exit(1);
} finally {
  fs.rmSync(tempRoot, { recursive: true, force: true });
}

console.log("Generated and compiled the Templ UI cookbook application successfully.");
