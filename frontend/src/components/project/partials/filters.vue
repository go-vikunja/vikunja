<template>
	<card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
	>
		<div class="field is-flex is-flex-direction-column">
			<Fancycheckbox
				v-model="params.filter_include_nulls"
				@update:modelValue="change()"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</Fancycheckbox>
		</div>
		
		<FilterInput 
			v-model="params.filter"
			:project-id="projectId"
		/>
		
		<template #footer>
			<x-button
				variant="primary"
				@click.prevent.stop="change()"
			>
				{{ $t('filters.showResults') }}
			</x-button>
		</template>
	</card>
</template>

<script lang="ts">
export const ALPHABETICAL_SORT = 'title'
</script>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {watchDebounced} from '@vueuse/core'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import {objectToSnakeCase} from '@/helpers/case'
import FilterInput from '@/components/project/partials/FilterInput.vue'
import {useRoute} from 'vue-router'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {transformFilterStringForApi} from '@/helpers/filters'

const props = defineProps({
	hasTitle: {
		type: Boolean,
		default: false,
	},
})

const modelValue = defineModel()

const route = useRoute()
const projectId = computed(() => {
	if (route.name?.startsWith('project.')) {
		return Number(route.params.projectId)
	}
	
	return undefined
})

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

// Using watchDebounced to prevent the filter re-triggering itself.
// FIXME: Only here until this whole component changes a lot with the new filter syntax.
watchDebounced(
	modelValue,
	(value) => {
		// FIXME: filters should only be converted to snake case in the last moment
		params.value = objectToSnakeCase(value)
	},
	{immediate: true, debounce: 500, maxWait: 1000},
)

const labelStore = useLabelStore()
const projectStore = useProjectStore()

function change() {
	const filter = transformFilterStringForApi(
		params.value.filter,
		labelTitle => labelStore.filterLabelsByQuery([], labelTitle)[0]?.id || null,
		projectTitle => projectStore.searchProject(projectTitle)[0]?.id || null,
	)
	
	modelValue.value = {
		...params.value,
		filter,
	}
}
</script>

<style lang="scss" scoped>
</style>
