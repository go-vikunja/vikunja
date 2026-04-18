<template>
	<BasePagination
		:total-pages="totalPages"
		:current-page="currentPage"
	>
		<template #previous="{ disabled }">
			<PaginationItem
				variant="previous"
				:disabled="disabled"
				@click="changePage(currentPage - 1)"
			>
				{{ $t('misc.previous') }}
			</PaginationItem>
		</template>
		<template #next="{ disabled }">
			<PaginationItem
				variant="next"
				:disabled="disabled"
				@click="changePage(currentPage + 1)"
			>
				{{ $t('misc.next') }}
			</PaginationItem>
		</template>
		<template #page-link="{ page, isCurrent }">
			<PaginationItem
				variant="link"
				:is-current="isCurrent"
				:aria-label="'Goto page ' + page.number"
				@click="changePage(page.number)"
			>
				{{ page.number }}
			</PaginationItem>
		</template>
	</BasePagination>
</template>

<script lang="ts" setup>
import BasePagination from '@/components/base/BasePagination.vue'
import PaginationItem from '@/components/misc/PaginationItem.vue'

const props = withDefaults(defineProps<{
	totalPages: number,
	currentPage?: number
}>(), {
	currentPage: 1,
})

const emit = defineEmits(['pageChanged'])

function changePage(page: number) {
	if (page >= 1 && page <= props.totalPages) {
		emit('pageChanged', page)
	}
}
</script>
