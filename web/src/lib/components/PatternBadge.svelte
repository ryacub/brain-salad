<script lang="ts">
	import { AlertTriangle } from 'lucide-svelte';

	interface Props {
		pattern: string;
	}

	let { pattern }: Props = $props();

	function getPatternColor(patternName: string): string {
		const name = patternName.toLowerCase();
		if (name.includes('context') || name.includes('switching')) {
			return 'variant-filled-error';
		}
		if (name.includes('perfectionism') || name.includes('over')) {
			return 'variant-filled-warning';
		}
		if (name.includes('tutorial') || name.includes('learning')) {
			return 'variant-filled-tertiary';
		}
		return 'variant-filled-surface';
	}
</script>

<span class="chip {getPatternColor(pattern)} text-xs flex items-center gap-1">
	<AlertTriangle size={12} />
	{pattern}
</span>
