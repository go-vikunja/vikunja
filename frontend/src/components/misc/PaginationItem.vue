<template>
	<RouterLink
		v-if="to !== undefined"
		:to="to"
		:disabled="disabled || undefined"
		:class="[`pagination-${variant}`, {'is-current': isCurrent}]"
	>
		<slot />
	</RouterLink>
	<BaseButton
		v-else
		:disabled="disabled"
		:class="[`pagination-${variant}`, {'is-current': isCurrent}]"
		@click="emit('click')"
	>
		<slot />
	</BaseButton>
</template>

<script lang="ts" setup>
import type {RouteLocationRaw} from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'

withDefaults(defineProps<{
	variant: 'previous' | 'next' | 'link',
	isCurrent?: boolean,
	disabled?: boolean,
	to?: RouteLocationRaw,
}>(), {
	isCurrent: false,
	disabled: false,
	to: undefined,
})

const emit = defineEmits<{
	(e: 'click'): void,
}>()
</script>

<style lang="scss" scoped>
// Rules ported from bulma-css-variables/sass/components/pagination.sass.
// PaginationItem owns the .pagination-previous / .pagination-next /
// .pagination-link markup, so scoped attributes attach directly to these
// classes — no :deep() necessary.

.pagination-previous,
.pagination-next,
.pagination-link {
	appearance: none;
	align-items: center;
	border: 1px solid transparent;
	border-radius: $radius;
	box-shadow: none;
	display: inline-flex;
	font-size: 1em;
	block-size: 2.5em;
	justify-content: center;
	line-height: 1.5;
	margin: 0.25rem;
	padding: calc(0.5em - 1px) 0.5em;
	position: relative;
	text-align: center;
	vertical-align: top;

	-webkit-touch-callout: none;
	user-select: none;

	&:focus,
	&:active {
		outline: none;
	}

	&[disabled],
	fieldset[disabled] & {
		cursor: not-allowed;
	}

	border-color: var(--border);
	color: var(--text-strong);
	min-inline-size: 2.5em;

	&:hover {
		border-color: var(--link-hover-border);
		color: var(--link-hover);
	}

	&:focus {
		border-color: var(--link-focus-border);
	}

	&:active {
		box-shadow: inset 0 1px 2px rgba($scheme-invert, 0.2);
	}

	&[disabled] {
		background-color: var(--border);
		border-color: var(--border);
		box-shadow: none;
		color: var(--text-light);
		opacity: 0.5;
	}
}

.pagination-previous,
.pagination-next {
	padding-inline: 0.75em;
	white-space: nowrap;

	&:not(:disabled):hover {
		background: $scheme-main;
		cursor: pointer;
	}
}

.pagination-link.is-current {
	background-color: var(--link);
	border-color: var(--link);
	color: var(--link-invert);
}

@media screen and (max-width: $tablet - 1px) {
	.pagination-previous,
	.pagination-next {
		flex-grow: 1;
		flex-shrink: 1;
	}
}

@media screen and (min-width: $tablet), print {
	.pagination-previous,
	.pagination-next,
	.pagination-link {
		margin-block: 0;
	}
}
</style>

<style lang="scss">
// Unscoped: this rule relies on ancestors (.app-container.has-background /
// .link-share-container.has-background) that live outside PaginationItem.
// Previously lived in styles/theme/background.scss, then BasePagination.vue.
.app-container.has-background .pagination-link:not(.is-current),
.link-share-container.has-background .pagination-link:not(.is-current) {
	background: var(--grey-100);
}
</style>
