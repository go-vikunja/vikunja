<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('list.search')"
		:search-results="foundLists"
		label="title"
		:select-placeholder="$t('list.searchSelect')"
		:model-value="list"
		@update:model-value="Object.assign(list, $event)"
		@select="select"
		@search="findLists"
	>
		<template #searchResult="{option}">
			<span class="list-namespace-title search-result">{{ namespace((option as IList).namespaceId) }} ></span>
			{{ (option as IList).title }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup>
import {reactive, ref, watch} from 'vue'
import type {PropType} from 'vue'
import {useI18n} from 'vue-i18n'

import type {IList} from '@/modelTypes/IList'
import type {INamespace} from '@/modelTypes/INamespace'

import {useListStore} from '@/stores/lists'
import {useNamespaceStore} from '@/stores/namespaces'

import ListModel from '@/models/list'

import Multiselect from '@/components/input/multiselect.vue'

const props = defineProps({
	modelValue: {
		type: Object as PropType<IList>,
		required: false,
	},
})
const emit = defineEmits(['update:modelValue'])

const {t} = useI18n({useScope: 'global'})

const list: IList = reactive(new ListModel())

watch(
	() => props.modelValue,
	(newList) => Object.assign(list, newList),
	{
		immediate: true,
		deep: true,
	},
)

const listStore = useListStore()
const namespaceStore = useNamespaceStore()
const foundLists = ref<IList[]>([])
function findLists(query: string) {
	if (query === '') {
		select(null)
	}
	foundLists.value = listStore.searchList(query)
}

function select(l: IList | null) {
	if (l === null) {
		return
	}
	Object.assign(list, l)
	emit('update:modelValue', list)
}

function namespace(namespaceId: INamespace['id']) {
	const namespace = namespaceStore.getNamespaceById(namespaceId)
	return namespace !== null
		? namespace.title
		: t('list.shared')
}
</script>

<style lang="scss" scoped>
.list-namespace-title {
	color: var(--grey-500);
}
</style>