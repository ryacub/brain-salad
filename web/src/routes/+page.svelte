<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { api } from '$lib/api';
	import IdeaCard from '$lib/components/IdeaCard.svelte';
	import IdeaForm from '$lib/components/IdeaForm.svelte';
	import FilterBar from '$lib/components/FilterBar.svelte';
	import { Loader2, AlertCircle } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	let status = $state('');
	let minScore = $state(0);

	const queryClient = useQueryClient();

	const ideasQuery = createQuery({
		queryKey: () => ['ideas', status],
		queryFn: () => api.ideas.list({ status: status || undefined })
	});

	const deleteMutation = createMutation({
		mutationFn: (id: string) => api.ideas.delete(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['ideas'] });
		}
	});

	const filteredIdeas = $derived(
		$ideasQuery.data?.ideas.filter((idea) => idea.final_score >= minScore) || []
	);

	function handleEdit(id: string) {
		goto(`/ideas/${id}/edit`);
	}

	function handleDelete(id: string) {
		if (confirm('Are you sure you want to delete this idea?')) {
			deleteMutation.mutate(id);
		}
	}

	function resetFilters() {
		status = '';
		minScore = 0;
	}
</script>

<svelte:head>
	<title>Dashboard - Telos Idea Matrix</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-4xl font-bold mb-2">Idea Dashboard</h1>
		<p class="text-surface-600-300-token">
			Capture and evaluate ideas against your personal goals and values
		</p>
	</div>

	<!-- Idea Form -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<IdeaForm />
	</div>

	<!-- Filters -->
	<FilterBar
		{status}
		{minScore}
		onStatusChange={(s) => (status = s)}
		onMinScoreChange={(score) => (minScore = score)}
		onReset={resetFilters}
	/>

	<!-- Ideas List -->
	<div>
		<div class="flex justify-between items-center mb-4">
			<h2 class="text-2xl font-bold">
				Your Ideas
				{#if $ideasQuery.data}
					<span class="text-surface-600-300-token text-lg">({filteredIdeas.length})</span>
				{/if}
			</h2>
		</div>

		{#if $ideasQuery.isPending}
			<div class="flex justify-center items-center py-12">
				<Loader2 size={32} class="animate-spin text-primary-500" />
			</div>
		{:else if $ideasQuery.isError}
			<div class="alert variant-filled-error">
				<AlertCircle size={20} />
				<p>Error loading ideas: {$ideasQuery.error?.message || 'Unknown error'}</p>
			</div>
		{:else if filteredIdeas.length === 0}
			<div class="card p-12 text-center bg-surface-100-800-token border border-surface-300-600-token">
				<p class="text-surface-600-300-token text-lg">
					{$ideasQuery.data?.ideas.length === 0
						? 'No ideas yet. Start by capturing your first idea above!'
						: 'No ideas match your filters.'}
				</p>
			</div>
		{:else}
			<div class="space-y-4">
				{#each filteredIdeas as idea (idea.id)}
					<IdeaCard {idea} onDelete={handleDelete} onEdit={handleEdit} />
				{/each}
			</div>
		{/if}
	</div>
</div>
