<template>
	<BasePagination
		:total-pages="totalPages"
		:current-page="currentPage"
	>
		<template #previous="{ disabled }">
			<PaginationItem
				variant="previous"
				:to="getRouteForPagination(currentPage - 1)"
				:disabled="disabled"
			>
				{{ $t('misc.previous') }}
			</PaginationItem>
		</template>
		<template #next="{ disabled }">
			<PaginationItem
				variant="next"
				:to="getRouteForPagination(currentPage + 1)"
				:disabled="disabled"
			>
				{{ $t('misc.next') }}
			</PaginationItem>
		</template>
		<template #page-link="{ page, isCurrent }">
			<PaginationItem
				variant="link"
				:to="getRouteForPagination(page.number)"
				:is-current="isCurrent"
				:aria-label="'Goto page ' + page.number"
			>
				{{ page.number }}
			</PaginationItem>
		</template>
	</BasePagination>
</template>

<script lang="ts" setup>
import BasePagination from '@/components/base/BasePagination.vue'
import PaginationItem from '@/components/misc/PaginationItem.vue'
import { useRoute } from 'vue-router'

withDefaults(defineProps<{
	totalPages: number,
	currentPage?: number
}>(), {
	currentPage: 0,
})

const route = useRoute()
function getRouteForPagination(page = 1) {
	return {
		name: route.name,
		params: route.params,
		query: { ...route.query, page },
	}
}
</script>
