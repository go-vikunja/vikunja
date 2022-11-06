<template>
	<multiselect
		v-model="selectedLists"
		:search-results="foundLists"
		:loading="listService.loading"
		:multiple="true"
		:placeholder="$t('list.search')"
		label="title"
		@search="findLists"
	/>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect, type PropType} from 'vue'

import Multiselect from '@/components/input/multiselect.vue'

import type {IList} from '@/modelTypes/IList'

import ListService from '@/services/list'
import {includesById} from '@/helpers/utils'

const props = defineProps({
	modelValue: {
		type: Array as PropType<IList[]>,
		default: () => [],
	},
})
const emit = defineEmits<{
	(e: 'update:modelValue', value: IList[]): void
}>()

const lists = ref<IList[]>([])

watchEffect(() => {
	lists.value = props.modelValue
})

const selectedLists = computed({
	get() {
		return lists.value
	},
	set: (value) => {
		lists.value = value
		emit('update:modelValue', value)
	},
})

const listService = shallowReactive(new ListService())
const foundLists = ref<IList[]>([])

async function findLists(query: string) {
	if (query === '') {
		foundLists.value = []
		return
	}

	const response = await listService.getAll({}, {s: query}) as IList[]

	// Filter selected items from the results
	foundLists.value = response.filter(({id}) => !includesById(lists.value, id))
}
</script>