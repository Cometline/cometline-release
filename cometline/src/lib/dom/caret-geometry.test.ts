// @vitest-environment jsdom

import { describe, expect, it } from 'vitest';
import { viewportDeltaToLocal } from './caret-geometry';

function mockWrap(opts: { offsetWidth: number; offsetHeight: number; wrapRect: DOMRect }) {
	const wrap = document.createElement('div');
	Object.defineProperty(wrap, 'offsetWidth', { value: opts.offsetWidth });
	Object.defineProperty(wrap, 'offsetHeight', { value: opts.offsetHeight });
	wrap.getBoundingClientRect = () => opts.wrapRect;
	return wrap;
}

describe('viewportDeltaToLocal', () => {
	it('keeps ratio ~1 when viewport size matches layout size', () => {
		const wrap = mockWrap({
			offsetWidth: 400,
			offsetHeight: 100,
			wrapRect: new DOMRect(10, 20, 400, 100)
		});
		const rect = new DOMRect(50, 30, 0, 22.5);

		const result = viewportDeltaToLocal(wrap, rect, 22.5);

		expect(result.x).toBeCloseTo(40);
		expect(result.y).toBeCloseTo(10);
		expect(result.h).toBeCloseTo(22.5);
	});

	it('applies inverse scale when viewport size is enlarged by transform', () => {
		const wrap = mockWrap({
			offsetWidth: 400,
			offsetHeight: 100,
			wrapRect: new DOMRect(10, 20, 404, 101)
		});
		const rect = new DOMRect(50, 30, 0, 0);

		const result = viewportDeltaToLocal(wrap, rect, 22.5);

		expect(result.x).toBeCloseTo(40 * (400 / 404), 5);
		expect(result.y).toBeCloseTo(10 * (100 / 101), 5);
		expect(result.h).toBeCloseTo(22.5 * (100 / 101), 5);
	});

	it('falls back to scale 1 when wrap viewport dimensions are zero', () => {
		const wrap = mockWrap({
			offsetWidth: 400,
			offsetHeight: 100,
			wrapRect: new DOMRect(10, 20, 0, 0)
		});
		const rect = new DOMRect(50, 30, 0, 18);

		const result = viewportDeltaToLocal(wrap, rect, 22.5);

		expect(result.x).toBeCloseTo(40);
		expect(result.y).toBeCloseTo(10);
		expect(result.h).toBeCloseTo(18);
	});
});
