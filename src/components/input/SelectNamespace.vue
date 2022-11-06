<template>
	<multiselect
		v-model="selectedNamespaces"
		:search-results="foundNamespaces"
		:loading="namespaceService.loading"
		:multiple="true"
		:placeholder="$t('namespace.search')"
		label="namespace"
		@search="findNamespaces"
	/>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect, type PropType} from 'vue'

import Multiselect from '@/components/input/multiselect.vue'

import type {INamespace} from '@/modelTypes/INamespace'

import NamespaceService from '@/services/namespace'
import {includesById} from '@/helpers/utils'

const props = defineProps({
	modelValue: {
		type: Array as PropType<INamespace[]>,
		default: () => [],
	},
})
const emit = defineEmits<{
	(e: 'update:modelValue', value: INamespace[]): void
}>()

const namespaces = ref<INamespace[]>([])

watchEffect(() => {
	namespaces.value = props.modelValue
})

const selectedNamespaces = computed({
	get() {
		return namespaces.value
	},
	set: (value) => {
		namespaces.value = value
		emit('update:modelValue', value)
	},
})

const namespaceService = shallowReactive(new NamespaceService())
const foundNamespaces = ref<INamespace[]>([])

async function findNamespaces(query: string) {
	if (query === '') {
		foundNamespaces.value = []
		return
	}

	const response = await namespaceService.getAll({}, {s: query}) as INamespace[]

	// Filter selected items from the results
	foundNamespaces.value = response.filter(({id}) => !includesById(namespaces.value, id))
}
</script>