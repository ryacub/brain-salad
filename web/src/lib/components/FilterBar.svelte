<script lang="ts">
	import { Filter, X } from 'lucide-svelte';

	interface Props {
		status: string;
		minScore: number;
		onStatusChange: (status: string) => void;
		onMinScoreChange: (minScore: number) => void;
		onReset: () => void;
	}

	let { status, minScore, onStatusChange, onMinScoreChange, onReset }: Props = $props();

	const statuses = [
		{ value: '', label: 'All' },
		{ value: 'active', label: 'Active' },
		{ value: 'archived', label: 'Archived' },
		{ value: 'completed', label: 'Completed' }
	];

	const hasFilters = $derived(status !== '' || minScore > 0);
</script>

<div class="card p-4 bg-surface-100-800-token border border-surface-300-600-token">
	<div class="flex flex-wrap items-center gap-4">
		<div class="flex items-center gap-2 text-surface-600-300-token">
			<Filter size={20} />
			<span class="font-semibold">Filters:</span>
		</div>

		<div class="flex-1 flex flex-wrap gap-4">
			<!-- Status Filter -->
			<div class="form-group">
				<label for="status-filter" class="label text-sm">
					<span>Status</span>
				</label>
				<select
					id="status-filter"
					class="select w-36"
					value={status}
					onchange={(e) => onStatusChange(e.currentTarget.value)}
				>
					{#each statuses as statusOption}
						<option value={statusOption.value}>{statusOption.label}</option>
					{/each}
				</select>
			</div>

			<!-- Score Filter -->
			<div class="form-group">
				<label for="score-filter" class="label text-sm">
					<span>Min Score: {minScore.toFixed(1)}</span>
				</label>
				<input
					id="score-filter"
					type="range"
					min="0"
					max="10"
					step="0.5"
					class="range w-48"
					value={minScore}
					oninput={(e) => onMinScoreChange(parseFloat(e.currentTarget.value))}
				/>
			</div>
		</div>

		{#if hasFilters}
			<button onclick={onReset} class="btn btn-sm variant-ghost-surface" aria-label="Reset filters">
				<X size={16} />
				<span>Reset</span>
			</button>
		{/if}
	</div>
</div>
