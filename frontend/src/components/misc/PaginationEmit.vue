<template>
	<BasePagination
		:total-pages="totalPages"
		:current-page="currentPage"
	>
		<template #previous="{ disabled }">
			<BaseButton
				:disabled="disabled"
				class="pagination-previous"
				@click="changePage(currentPage - 1)"
			>
				{{ $t('misc.previous') }}
			</BaseButton>
		</template>
		<template #next="{ disabled }">
			<BaseButton
				:disabled="disabled"
				class="pagination-next"
				@click="changePage(currentPage + 1)"
			>
				{{ $t('misc.next') }}
			</BaseButton>
		</template>
		<template #page-link="{ page, isCurrent }">
			<BaseButton
				class="pagination-link"
				:aria-label="'Goto page ' + page.number"
				:class="{ 'is-current': isCurrent }"
				@click="changePage(page.number)"
			>
				{{ page.number }}
			</BaseButton>
		</template>
	</BasePagination>
</template>

<script lang="ts" setup>
import BasePagination from '@/components/base/BasePagination.vue'
import BaseButton from '@/components/base/BaseButton.vue'

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
