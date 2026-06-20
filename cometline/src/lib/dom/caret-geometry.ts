/** Convert a viewport-space caret rect to wrap-local coordinates. */
export function viewportDeltaToLocal(
	wrap: HTMLElement,
	rect: Pick<DOMRect, 'left' | 'top' | 'height'>,
	lineHeight: number
): { x: number; y: number; h: number } {
	const wrapRect = wrap.getBoundingClientRect();
	const scaleX = wrapRect.width > 0 ? wrap.offsetWidth / wrapRect.width : 1;
	const scaleY = wrapRect.height > 0 ? wrap.offsetHeight / wrapRect.height : 1;

	return {
		x: (rect.left - wrapRect.left) * scaleX,
		y: (rect.top - wrapRect.top) * scaleY,
		h: (rect.height || lineHeight) * scaleY
	};
}
