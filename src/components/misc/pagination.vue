<template>
	<nav
		aria-label="pagination"
		class="pagination is-centered p-4"
		role="navigation"
		v-if="totalPages > 1"
	>
		<router-link
			:disabled="currentPage === 1 || undefined"
			:to="getRouteForPagination(currentPage - 1)"
			class="pagination-previous">
			{{ $t('misc.previous') }}
		</router-link>
		<router-link
			:disabled="currentPage === totalPages || undefined"
			:to="getRouteForPagination(currentPage + 1)"
			class="pagination-next">
			{{ $t('misc.next') }}
		</router-link>
		<ul class="pagination-list">
			<li :key="`page-${i}`" v-for="(p, i) in pages">
				<span class="pagination-ellipsis" v-if="p.isEllipsis">&hellip;</span>
				<router-link
					v-else
					class="pagination-link"
					:aria-label="'Goto page ' + p.number"
					:class="{ 'is-current': p.number === currentPage }"
					:to="getRouteForPagination(p.number)"
				>
					{{ p.number }}
				</router-link>
			</li>
		</ul>
	</nav>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

function createPagination(totalPages: number, currentPage: number) {
	const pages = []
	for (let i = 0; i < totalPages; i++) {

		// Show ellipsis instead of all pages
		if (
			i > 0 && // Always at least the first page
			(i + 1) < totalPages && // And the last page
			(
				// And the current with current + 1 and current - 1
				(i + 1) > currentPage + 1 ||
				(i + 1) < currentPage - 1
			)
		) {
			// Only add an ellipsis if the last page isn't already one
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

function getRouteForPagination(page = 1, type = 'list') {
	return {
		name: 'list.' + type,
		params: {
			type: type,
		},
		query: {
			page: page,
		},
	}
}

const props = defineProps({
	totalPages: {
		type: Number,
		required: true,
	},
	currentPage: {
		type: Number,
		default: 0,
	},
})

const pages = computed(() => createPagination(props.totalPages, props.currentPage))
</script>

<style lang="scss" scoped>
.pagination {
	padding-bottom: 1rem;
}

.pagination-previous,
.pagination-next {
	&:not(:disabled):hover {
		background: $scheme-main;
		cursor: pointer;
	}
}
</style>