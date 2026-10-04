<template>
	<span
		class="risk-rating-badge"
		:class="`is-${rating}`"
	>
		{{ $t(`risks.ratings.${rating}`) }}
		<span
			v-if="score !== undefined"
			class="risk-rating-badge__score"
		>{{ score }}</span>
	</span>
</template>

<script setup lang="ts">
import type {RiskRating} from '@/helpers/riskRating'

defineProps<{
	rating: RiskRating
	score?: number
}>()
</script>

<style lang="scss" scoped>
.risk-rating-badge {
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	padding: .1rem .5rem;
	border-radius: 1rem;
	font-size: .8rem;
	font-weight: 600;
	white-space: nowrap;
	// The colours carry meaning, so they stay on paper.
	print-color-adjust: exact;
	-webkit-print-color-adjust: exact;

	&.is-low {
		background: var(--success);
		color: var(--white);
	}

	&.is-medium {
		background: var(--warning);
		color: var(--grey-900);
	}

	&.is-high {
		background: var(--orange, #f57c00);
		color: var(--white);
	}

	&.is-critical {
		background: var(--danger);
		color: var(--white);
	}
}

.risk-rating-badge__score {
	font-weight: 400;
	opacity: .85;
}
</style>
