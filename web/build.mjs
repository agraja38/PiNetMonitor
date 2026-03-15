import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const root = fileURLToPath(new URL(".", import.meta.url));
const srcDir = path.join(root, "src");
const distDir = path.join(root, "dist");

await rm(distDir, { recursive: true, force: true });
await mkdir(distDir, { recursive: true });

for (const file of ["index.html", "styles.css", "app.js"]) {
  const content = await readFile(path.join(srcDir, file), "utf8");
  await writeFile(path.join(distDir, file), content);
}

console.log("PiNetMonitor frontend build complete");
