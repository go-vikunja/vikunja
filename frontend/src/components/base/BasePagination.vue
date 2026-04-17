<template>
	<nav
		v-if="totalPages > 1"
		aria-label="pagination"
		class="pagination is-centered p-4"
		role="navigation"
	>
		<slot
			name="previous"
			:disabled="currentPage === 1"
		>
			{{ $t('misc.previous') }}
		</slot>
		<slot
			name="next"
			:disabled="currentPage === totalPages"
		>
			{{ $t('misc.next') }}
		</slot>
		<ul class="pagination-list">
			<li
				v-for="(p, i) in pages"
				:key="`page-${i}`"
			>
				<span
					v-if="p.isEllipsis"
					class="pagination-ellipsis"
				>&hellip;</span>
				<slot
					v-else
					name="page-link"
					:page="p"
					:is-current="p.number === currentPage"
				>
					{{ p.number }}
				</slot>
			</li>
		</ul>
	</nav>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

interface PaginationPage {
	number: number
	isEllipsis: boolean
}

const props = defineProps<{
	totalPages: number,
	currentPage: number
}>()

function createPagination(totalPages: number, currentPage: number): PaginationPage[] {
	const pages: PaginationPage[] = []
	for (let i = 0; i < totalPages; i++) {
		if (
			i > 0 &&
			(i + 1) < totalPages &&
			((i + 1) > currentPage + 1 || (i + 1) < currentPage - 1)
		) {
			const prevPage = pages[i - 1]
			if (prevPage && !prevPage.isEllipsis) {
				pages.push({
					number: 0,
					isEllipsis: true,
				})
			}
			continue
		}
		
		pages.push({
			number: i + 1,
			isEllipsis: false,
		})
	}
	return pages
}

const pages = computed(() => createPagination(props.totalPages, props.currentPage))
</script>

<style lang="scss" scoped>
// Rules ported from bulma-css-variables/sass/components/pagination.sass.
// Slot content (.pagination-previous, .pagination-next, .pagination-link) is
// rendered by Pagination.vue / PaginationEmit.vue, so scoped attributes don't
// reach them — we use :deep() to style across the slot boundary.

.pagination {
	align-items: center;
	display: flex;
	font-size: $size-normal;
	justify-content: center;
	margin: -0.25rem;
	padding-block-end: 1rem;
	text-align: center;
}

.pagination-list {
	align-items: center;
	display: flex;
	flex-wrap: wrap;
	justify-content: center;
	text-align: center;

	&, & li {
		margin-block-start: 0;
	}

	li {
		list-style: none;
	}
}

:deep(.pagination-previous),
:deep(.pagination-next),
:deep(.pagination-link),
.pagination-ellipsis {
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
}

:deep(.pagination-previous),
:deep(.pagination-next),
:deep(.pagination-link) {
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

:deep(.pagination-previous),
:deep(.pagination-next) {
	padding-inline: 0.75em;
	white-space: nowrap;

	&:not(:disabled):hover {
		background: $scheme-main;
		cursor: pointer;
	}
}

:deep(.pagination-link.is-current) {
	background-color: var(--link);
	border-color: var(--link);
	color: var(--link-invert);
}

.pagination-ellipsis {
	color: var(--grey-light);
	pointer-events: none;
}

@media screen and (max-width: $tablet - 1px) {
	.pagination {
		flex-wrap: wrap;
	}

	:deep(.pagination-previous),
	:deep(.pagination-next) {
		flex-grow: 1;
		flex-shrink: 1;
	}

	.pagination-list li {
		flex-grow: 1;
		flex-shrink: 1;
	}
}

@media screen and (min-width: $tablet), print {
	.pagination-list {
		flex-grow: 1;
		flex-shrink: 1;
	}

	:deep(.pagination-previous),
	:deep(.pagination-next),
	:deep(.pagination-link),
	.pagination-ellipsis {
		margin-block: 0;
	}

	.pagination {
		margin-block: 0;

		&.is-centered {
			.pagination-list {
				justify-content: center;
			}
		}
	}
}
</style>

<style lang="scss">
// Unscoped: this rule relies on ancestors (.app-container.has-background /
// .link-share-container.has-background) that live outside BasePagination.
// Previously lived in styles/theme/background.scss.
.app-container.has-background .pagination-link:not(.is-current),
.link-share-container.has-background .pagination-link:not(.is-current) {
	background: var(--grey-100);
}
</style>

