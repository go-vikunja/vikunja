<template>
	<BasePagination
		:total-pages="totalPages"
		:current-page="currentPage"
	>
		<template #previous="{ disabled }">
			<RouterLink
				:disabled="disabled || undefined"
				:to="getRouteForPagination(currentPage - 1)"
				class="pagination-previous"
			>
				{{ $t('misc.previous') }}
			</RouterLink>
		</template>
		<template #next="{ disabled }">
			<RouterLink
				:disabled="disabled || undefined"
				:to="getRouteForPagination(currentPage + 1)"
				class="pagination-next"
			>
				{{ $t('misc.next') }}
			</RouterLink>
		</template>
		<template #page-link="{ page, isCurrent }">
			<RouterLink
				class="pagination-link"
				:aria-label="'Goto page ' + page.number"
				:class="{ 'is-current': isCurrent }"
				:to="getRouteForPagination(page.number)"
			>
				{{ page.number }}
			</RouterLink>
		</template>
	</BasePagination>
</template>

<script lang="ts" setup>
import BasePagination from '@/components/base/BasePagination.vue'

withDefaults(defineProps<{
	totalPages: number,
	currentPage?: number
}>(), {
	currentPage: 0,
})

function getRouteForPagination(page = 1, type = null) {
	return {
		name: type,
		params: {
			type: type,
		},
		query: {
			page: page,
		},
	}
}
</script>
