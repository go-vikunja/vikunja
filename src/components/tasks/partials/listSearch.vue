<template>
	<Multiselect
		class="control is-expanded"
		:placeholder="$t('list.search')"
		@search="findLists"
		:search-results="foundLists"
		@select="select"
		label="title"
		v-model="list"
		:select-placeholder="$t('list.searchSelect')"
	>
		<template #searchResult="props">
			<span class="list-namespace-title search-result">{{ namespace(props.option.namespaceId) }} ></span>
			{{ props.option.title }}
		</template>
	</Multiselect>
</template>

<script lang="ts" setup>
import {reactive, ref, watch} from 'vue'
import type {PropType} from 'vue'
import {useStore} from '@/store'
import {useI18n} from 'vue-i18n'
import ListModel from '@/models/list'
import type {IList} from '@/modelTypes/IList'
import Multiselect from '@/components/input/multiselect.vue'
import {useListStore} from '@/stores/lists'

const props = defineProps({
	modelValue: {
		type: Object as PropType<IList>,
		required: false,
	},
})
const emit = defineEmits(['update:modelValue'])

const store = useStore()
const listStore = useListStore()
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

function namespace(namespaceId: number) {
	const namespace = store.getters['namespaces/getNamespaceById'](namespaceId)
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