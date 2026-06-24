import { cp, mkdir, rm } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const packageDir = path.dirname(fileURLToPath(new URL('../package.json', import.meta.url)));
const publicDir = path.join(packageDir, 'public');
const distDir = path.join(packageDir, 'dist');

await rm(distDir, { recursive: true, force: true });
await mkdir(distDir, { recursive: true });
await cp(publicDir, distDir, { recursive: true });

console.log(`Built BoltSight static assets to ${distDir}`);
