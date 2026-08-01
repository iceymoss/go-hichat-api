/**
 * Converts the source screenshots into web-sized WebP files.
 *
 *   bun run optimize:screenshots
 *
 * Reads the untouched originals from ../docs/screenshots (~3024px wide PNGs)
 * and writes public/screenshots/*.webp. Always reads from the originals so
 * repeated runs never re-compress an already-compressed file.
 *
 * This step is required because next.config.ts uses output: 'export', which
 * forces images.unoptimized — Next.js will not resize or convert anything at
 * build time, so whatever lands in public/ is what the browser downloads.
 */
import { mkdir, readdir, stat, unlink } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const __dirname = dirname(fileURLToPath(import.meta.url));
const SRC_DIR = resolve(__dirname, "../../docs/screenshots");
const OUT_DIR = resolve(__dirname, "../public/screenshots");

/** Only these are used by the site; see src/lib/content.ts. */
const KEEP = [
  "single-chat",
  "group-chat",
  "conversation-list",
  "create-group",
  "friend-list-and-settings",
  "friend-requests-received",
  "profile-home",
  "moments-feed-and-detail",
  "publish-moment",
  "moments",
  "call-incoming",
  "group-call-active",
];

/** 1600px covers a ~800px card at 2x DPR without shipping 3024px. */
const TARGET_WIDTH = 1600;
const QUALITY = 82;

const kb = (bytes) => Math.round(bytes / 1024);

async function main() {
  await mkdir(OUT_DIR, { recursive: true });

  let totalBefore = 0;
  let totalAfter = 0;
  const missing = [];

  for (const name of KEEP) {
    const src = join(SRC_DIR, `${name}.png`);
    const out = join(OUT_DIR, `${name}.webp`);

    let srcStat;
    try {
      srcStat = await stat(src);
    } catch {
      missing.push(`${name}.png`);
      continue;
    }

    const { width, height } = await sharp(src)
      .resize({ width: TARGET_WIDTH, withoutEnlargement: true })
      .webp({ quality: QUALITY })
      .toFile(out);

    const outStat = await stat(out);
    totalBefore += srcStat.size;
    totalAfter += outStat.size;

    const pct = Math.round((1 - outStat.size / srcStat.size) * 100);
    console.log(
      `${name.padEnd(30)} ${String(kb(srcStat.size)).padStart(5)} KB -> ` +
        `${String(kb(outStat.size)).padStart(4)} KB webp  ` +
        `${width}x${height}  (-${pct}%)`
    );
  }

  // Drop the superseded PNG copies so they are not deployed.
  let removed = 0;
  for (const file of await readdir(OUT_DIR)) {
    if (file.endsWith(".png")) {
      await unlink(join(OUT_DIR, file));
      removed += 1;
    }
  }

  console.log("");
  console.log(`total: ${kb(totalBefore)} KB -> ${kb(totalAfter)} KB webp`);
  console.log(
    `saved: ${kb(totalBefore - totalAfter)} KB ` +
      `(-${Math.round((1 - totalAfter / totalBefore) * 100)}%)`
  );
  if (removed) console.log(`removed ${removed} superseded .png file(s)`);
  if (missing.length) {
    console.warn(`\nWARNING missing sources: ${missing.join(", ")}`);
    process.exitCode = 1;
  }
}

await main();
