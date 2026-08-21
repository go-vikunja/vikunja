<template>
	<header class="page-header">
		<div class="page-heading">
			<p
				v-if="eyebrow"
				class="page-eyebrow"
			>
				{{ eyebrow }}
			</p>
			<h1 class="page-title">
				{{ title }}
			</h1>
			<p
				v-if="description"
				class="page-description"
			>
				{{ description }}
			</p>
			<div
				v-if="$slots.meta"
				class="page-meta"
			>
				<slot name="meta" />
			</div>
		</div>
		<div
			v-if="$slots.actions"
			class="page-actions"
		>
			<slot name="actions" />
		</div>
	</header>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	title: string,
	description?: string,
	eyebrow?: string,
}>(), {
	description: '',
	eyebrow: '',
})
</script>

<style lang="scss" scoped>
.page-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1.5rem;
	margin-block-end: 1.5rem;
}

.page-heading {
	min-inline-size: 0;
}

.page-eyebrow {
	margin-block: 0 .25rem;
	color: var(--text-muted);
	font-size: .75rem;
	font-weight: 700;
	letter-spacing: .08em;
	text-transform: uppercase;
}

.page-title {
	margin: 0;
	color: var(--text-strong);
	font-family: $vikunja-font;
	font-size: clamp(1.75rem, 2vw, 2.25rem);
	font-weight: 700 !important;
	line-height: 1.15;
}

.page-description {
	max-inline-size: 48rem;
	margin: .5rem 0 0;
	color: var(--text-muted);
	font-size: .9375rem;
}

.page-meta {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
	margin-block-start: .75rem;
}

.page-actions {
	display: flex;
	flex: 0 0 auto;
	flex-wrap: wrap;
	justify-content: flex-end;
	gap: .625rem;
}

@media screen and (max-width: $tablet) {
	.page-header {
		flex-direction: column;
		gap: 1rem;
	}

	.page-actions {
		inline-size: 100%;
		justify-content: flex-start;
	}
}
</style>
