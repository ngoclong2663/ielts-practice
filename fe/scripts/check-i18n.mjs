// Fails when a translation key exists in one locale catalogue but not in another.
import { readFileSync, readdirSync } from "node:fs";
import { join } from "node:path";

const dir = join(process.cwd(), "messages");
const files = readdirSync(dir).filter((f) => f.endsWith(".json"));

function flatten(obj, prefix = "") {
  return Object.entries(obj).flatMap(([key, value]) =>
    value && typeof value === "object" ? flatten(value, `${prefix}${key}.`) : [`${prefix}${key}`],
  );
}

const catalogues = Object.fromEntries(
  files.map((f) => [f, new Set(flatten(JSON.parse(readFileSync(join(dir, f), "utf8"))))]),
);
const allKeys = new Set(Object.values(catalogues).flatMap((s) => [...s]));

let missing = 0;
for (const [file, keys] of Object.entries(catalogues)) {
  for (const key of allKeys) {
    if (!keys.has(key)) {
      console.error(`✗ ${file} is missing "${key}"`);
      missing++;
    }
  }
}
if (missing > 0) {
  process.exit(1);
}
console.log(`✓ ${files.length} catalogues, ${allKeys.size} keys each`);
