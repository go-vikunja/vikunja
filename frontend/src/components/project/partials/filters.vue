<template>
	<card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
	>
		<FilterInput
			v-model="params.filter"
			:project-id="projectId"
			@blur="change()"
		/>

		<div class="field is-flex is-flex-direction-column">
			<Fancycheckbox
				v-model="params.filter_include_nulls"
				@blur="change()"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</Fancycheckbox>
		</div>

		<FilterInputDocs />

		<template
			v-if="hasFooter"
			#footer
		>
			<x-button
				variant="primary"
				@click.prevent.stop="changeAndEmitButton"
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
import FilterInput from '@/components/project/partials/FilterInput.vue'
import {useRoute} from 'vue-router'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {FILTER_OPERATORS, transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'
import FilterInputDocs from '@/components/project/partials/FilterInputDocs.vue'

const  {
	hasTitle= false,
	hasFooter = true,
	modelValue,
} = defineProps<{
	hasTitle?: boolean,
	hasFooter?: boolean,
	modelValue: TaskFilterParams,
}>()

const emit = defineEmits(['update:modelValue', 'showResultsButtonClicked'])

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
watchDebounced(
	() => modelValue,
	(value: TaskFilterParams) => {
		const val = {...value}
		val.filter = transformFilterStringFromApi(
			val?.filter || '',
			labelId => labelStore.getLabelById(labelId)?.title,
			projectId => projectStore.projects.value[projectId]?.title || null,
		)
		params.value = val
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

	let s = ''

	// When the filter does not contain any filter tokens, assume a simple search and redirect the input
	const hasFilterQueries = FILTER_OPERATORS.find(o => filter.includes(o)) || false
	if (!hasFilterQueries) {
		s = filter
	}

	const newParams = {
		...params.value,
		filter: s === '' ? filter : '',
		s,
	}

	if (JSON.stringify(modelValue) === JSON.stringify(newParams)) {
		return
	}

	emit('update:modelValue', newParams)
}

function changeAndEmitButton() {
	change()
	emit('showResultsButtonClicked')
}
</script>
