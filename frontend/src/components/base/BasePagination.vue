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

const props = defineProps<{
	totalPages: number,
	currentPage: number
}>()

function createPagination(totalPages: number, currentPage: number) {
	const pages = []
	for (let i = 0; i < totalPages; i++) {
		if (
			i > 0 &&
			(i + 1) < totalPages &&
			((i + 1) > currentPage + 1 || (i + 1) < currentPage - 1)
		) {
			if (pages[i - 1] && !pages[i - 1].isEllipsis) {
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
.pagination {
	padding-block-end: 1rem;
}

.pagination-previous,
.pagination-next {
	&:not(:disabled):hover {
		background: $scheme-main;
		cursor: pointer;
	}
}

.pagination-list {
	&, & li {
		margin-block-start: 0;
	}
}
</style>
  
