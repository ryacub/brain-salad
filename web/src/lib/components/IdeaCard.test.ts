import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import IdeaCard from './IdeaCard.svelte';
import type { Idea } from '$lib/types';

describe('IdeaCard', () => {
	const mockIdea: Idea = {
		id: '1',
		content: 'Test Idea',
		raw_score: 8.0,
		final_score: 8.5,
		patterns: ['context-switching'],
		recommendation: 'This is a strong idea',
		created_at: '2025-01-01T00:00:00Z',
		status: 'active'
	};

	it('renders idea content correctly', () => {
		render(IdeaCard, { props: { idea: mockIdea } });
		expect(screen.getByText('Test Idea')).toBeInTheDocument();
	});

	it('displays the final score', () => {
		render(IdeaCard, { props: { idea: mockIdea } });
		expect(screen.getByText('8.5')).toBeInTheDocument();
	});

	it('shows patterns when present', () => {
		render(IdeaCard, { props: { idea: mockIdea } });
		expect(screen.getByText('context-switching')).toBeInTheDocument();
	});

	it('displays recommendation when provided', () => {
		render(IdeaCard, { props: { idea: mockIdea } });
		expect(screen.getByText('This is a strong idea')).toBeInTheDocument();
	});

	it('shows active status', () => {
		render(IdeaCard, { props: { idea: mockIdea } });
		expect(screen.getByText('active')).toBeInTheDocument();
	});

	it('calls onDelete when delete button is clicked', async () => {
		const onDelete = vi.fn();
		const { container } = render(IdeaCard, {
			props: { idea: mockIdea, onDelete }
		});

		const deleteButton = container.querySelector('button[aria-label="Delete idea"]');
		expect(deleteButton).toBeInTheDocument();
	});

	it('calls onEdit when edit button is clicked', async () => {
		const onEdit = vi.fn();
		const { container } = render(IdeaCard, {
			props: { idea: mockIdea, onEdit }
		});

		const editButton = container.querySelector('button[aria-label="Edit idea"]');
		expect(editButton).toBeInTheDocument();
	});
});
