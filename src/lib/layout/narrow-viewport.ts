/** Matches `--sidebar-auto-collapse-max` (default 900px). */
export function narrowViewportQuery(): MediaQueryList {
	const max = getComputedStyle(document.documentElement)
		.getPropertyValue('--sidebar-auto-collapse-max')
		.trim();
	return window.matchMedia(`(max-width: ${max || '900px'})`);
}

export function isNarrowViewport(): boolean {
	return narrowViewportQuery().matches;
}
