import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import PatternBadge from './PatternBadge.svelte';

describe('PatternBadge', () => {
	it('renders pattern name', () => {
		render(PatternBadge, { props: { pattern: 'context-switching' } });
		expect(screen.getByText('context-switching')).toBeInTheDocument();
	});

	it('renders perfectionism pattern', () => {
		render(PatternBadge, { props: { pattern: 'perfectionism' } });
		expect(screen.getByText('perfectionism')).toBeInTheDocument();
	});

	it('renders tutorial hell pattern', () => {
		render(PatternBadge, { props: { pattern: 'tutorial-hell' } });
		expect(screen.getByText('tutorial-hell')).toBeInTheDocument();
	});
});
