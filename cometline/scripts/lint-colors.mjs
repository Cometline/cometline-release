import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const root = new URL('../src', import.meta.url).pathname;
const allowlist = new Set(['app.css']);
const forbidden = ['#b42318', '#15803d', '#B42318', '#15803D'];

/** @type {string[]} */
const violations = [];

function walk(dir) {
	for (const entry of readdirSync(dir)) {
		const path = join(dir, entry);
		const stat = statSync(path);
		if (stat.isDirectory()) {
			walk(path);
			continue;
		}
		if (!/\.(css|svelte)$/.test(entry)) continue;
		const rel = relative(root, path);
		if (allowlist.has(rel)) continue;
		const content = readFileSync(path, 'utf8');
		for (const hex of forbidden) {
			if (content.includes(hex)) {
				violations.push(`${rel}: ${hex}`);
			}
		}
	}
}

walk(root);

if (violations.length > 0) {
	console.error('Forbidden status hex colors found outside app.css:\n');
	for (const line of violations) console.error(`  ${line}`);
	process.exit(1);
}

console.log('No forbidden status hex colors outside app.css.');
