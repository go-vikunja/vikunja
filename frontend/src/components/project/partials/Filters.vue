<template>
	<Card
		class="filters has-overflow"
		:title="hasTitle ? $t('filters.title') : ''"
		role="search"
		:show-close="showClose"
		@close="$emit('close')"
	>
		<FilterEditor
			ref="filterInputRef"
			v-model="params.filter"
			v-model:search="params.s"
			:project-id="projectId"
			class="mbe-2"
			@change="editorChanged"
			@update:search="change('always')"
		/>
		<div 
			v-if="filterFromView"
			class="tw:text-sm mbe-2"
		>
			{{ $t('filters.fromView') }}
			<code>{{ filterFromView }}</code><br>
			{{ $t('filters.fromViewBoth') }}
		</div>

		<div class="field is-flex is-flex-direction-column">
			<FancyCheckbox
				v-model="params.filter_include_nulls"
				@update:modelValue="change('always')"
			>
				{{ $t('filters.attributes.includeNulls') }}
			</FancyCheckbox>
		</div>

		<template
			v-if="hasFooter"
			#footer
		>
			<XButton
				variant="secondary"
				class="mie-2"
				:disabled="params.filter === '' && params.s === ''"
				@click.prevent.stop="clearFiltersAndEmit"
			>
				{{ $t('filters.clear') }}
			</XButton>
			<XButton
				variant="primary"
				@click.prevent.stop="changeAndEmitButton"
			>
				{{ $t('filters.showResults') }}
			</XButton>
		</template>
	</Card>
</template>

<script setup lang="ts" generic="T extends EditableTaskCollection">
import {computed, ref, watch} from 'vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {useRoute} from 'vue-router'
import type {EditableTaskCollection} from '@/types/EditableTaskCollection'
import {hasFilterQuery} from '@/helpers/filters'
import FilterEditor from '@/components/input/filter/FilterEditor.vue'

const props = withDefaults(defineProps<{
	modelValue: T,
	hasTitle?: boolean,
	hasFooter?: boolean,
	changeImmediately?: boolean,
	filterFromView?: string,
	showClose?: boolean,
}>(), {
	hasTitle: false,
	hasFooter: true,
	changeImmediately: false,
	filterFromView: undefined,
	showClose: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: T],
	'showResults': [],
	'close': [],
}>()

const route = useRoute()
const projectId = computed(() => {
	if (typeof route.name === 'string' && route.name.startsWith('project.')) {
		const id = Number(route.params.projectId)
		return id > 0 ? id : undefined
	}

	return undefined
})

const params = ref<EditableTaskCollection>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

const filterInputRef = ref()

watch(
	() => props.modelValue,
	(value: T) => {
		params.value = {...value}
	},
	{
		immediate: true,
		deep: true,
	},
)

function change(event: 'blur' | 'modelValue' | 'always') {
	if (event !== 'always') {
		if (props.changeImmediately && event === 'blur') {
			return
		}

		if (!props.changeImmediately && event === 'modelValue') {
			return
		}
	}

	let filter = params.value.filter ?? ''
	let s = params.value.s ?? ''
	if (filter !== '' && !hasFilterQuery(filter)) {
		s = filter
		filter = ''
	}

	const newParams = {
		...props.modelValue,
		...params.value,
		filter,
		s,
	}

	if (JSON.stringify(props.modelValue) === JSON.stringify(newParams)) {
		return
	}

	emit('update:modelValue', newParams)
}

// The Advanced input shows `filter || s`, so it owns both.
function editorChanged(mode: 'simple' | 'advanced') {
	if (mode === 'advanced') {
		params.value.s = ''
	}
	change(mode === 'simple' ? 'always' : 'modelValue')
}

function changeAndEmitButton() {
	change('always')
	emit('showResults')
}

function clearFiltersAndEmit() {
	params.value.filter = ''
	params.value.s = ''
	changeAndEmitButton()
}

function focusFilterInput() {
	filterInputRef.value?.focus()
}

defineExpose({
	focusFilterInput,
})
</script>
