<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { createQuery } from '@tanstack/svelte-query';
	import { api } from '$lib/api';
	import IdeaForm from '$lib/components/IdeaForm.svelte';
	import { Loader2, AlertCircle, ArrowLeft } from 'lucide-svelte';

	const id = $derived($page.params.id);

	const ideaQuery = createQuery({
		queryKey: () => ['idea', id],
		queryFn: () => api.ideas.get(id)
	});

	function handleSuccess() {
		goto(`/ideas/${id}`);
	}

	function handleCancel() {
		goto(`/ideas/${id}`);
	}
</script>

<svelte:head>
	<title>Edit Idea - Telos Idea Matrix</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center gap-4">
		<button onclick={handleCancel} class="btn btn-sm variant-ghost-surface">
			<ArrowLeft size={16} />
			Back
		</button>
		<h1 class="text-3xl font-bold">Edit Idea</h1>
	</div>

	{#if $ideaQuery.isPending}
		<div class="flex justify-center items-center py-12">
			<Loader2 size={32} class="animate-spin text-primary-500" />
		</div>
	{:else if $ideaQuery.isError}
		<div class="alert variant-filled-error">
			<AlertCircle size={20} />
			<p>Error loading idea: {$ideaQuery.error?.message || 'Unknown error'}</p>
		</div>
	{:else if $ideaQuery.data}
		<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token max-w-3xl">
			<IdeaForm idea={$ideaQuery.data} onSuccess={handleSuccess} onCancel={handleCancel} />
		</div>
	{/if}
</div>
