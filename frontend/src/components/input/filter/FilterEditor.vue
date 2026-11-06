<script setup lang="ts">
import {
	computed,
	nextTick,
	ref,
	useId,
	watch,
} from 'vue'
import {useI18n} from 'vue-i18n'
import {emptySimpleFilter, fromQuery, toQuery, type SimpleFilter} from '@/helpers/simpleFilter'
import {useAuthStore} from '@/stores/auth'
import {useUpdateFrontendSettingsMutation} from '@/client/queries/account'
import FilterInput from './FilterInput.vue'
import FilterInputDocs from './FilterInputDocs.vue'
import SimpleFilterForm from './SimpleFilterForm.vue'
import XButton from '@/components/input/Button.vue'

const props = defineProps<{
	modelValue: string
	search?: string
	projectId?: number
	label?: string
}>()
const emit = defineEmits<{
	'update:modelValue': [value: string]
	'update:search': [value: string]
	'change': [mode: 'simple' | 'advanced']
}>()
const {t} = useI18n({useScope: 'global'})
const auth = useAuthStore()
const updateSettings = useUpdateFrontendSettingsMutation()
const parsed = computed(() => fromQuery(props.modelValue))
// eslint-disable-next-line vue/no-ref-object-reactivity-loss -- the watch on parsed takes over after setup
const initial = parsed.value
const mode = ref<'simple' | 'advanced'>(initial === null ? 'advanced' : auth.settings.frontend_settings.filter_editor_mode)
const simple = ref(initial ?? emptySimpleFilter())
const searchText = ref('')
const advancedInput = ref<InstanceType<typeof FilterInput>>()
const root = ref<HTMLElement>()
const id = useId()
watch(() => props.search, value => { searchText.value = value ?? '' }, {immediate: true})
watch(parsed, value => {
	if (value === null) {
		mode.value = 'advanced'
	} else if (toQuery(simple.value) !== props.modelValue) {
		simple.value = value
	}
})

function selectMode(value: 'simple' | 'advanced') {
	if (value === 'simple') {
		if (parsed.value === null) {
			return
		}
		simple.value = parsed.value
	}
	mode.value = value
	saveMode(value)
}

function saveMode(value: 'simple' | 'advanced') {
	const session = auth.session
	if (session && auth.settings.frontend_settings.filter_editor_mode !== value) {
		updateSettings.mutate({
			id: session.id,
			type: session.type,
			frontendSettings: {filter_editor_mode: value},
		})
	}
}

async function moveTab() {
	selectMode(mode.value === 'simple' ? 'advanced' : 'simple')
	await nextTick()
	document.getElementById(`${id}-${mode.value}`)?.focus()
}

function changeSimple(value: SimpleFilter) {
	simple.value = value
	emit('update:modelValue', toQuery(value))
	emit('change', 'simple')
}

function changeAdvanced(value: string) {
	emit('update:modelValue', value)
	emit('change', 'advanced')
}

async function clearAndUseSimple() {
	simple.value = emptySimpleFilter()
	emit('update:modelValue', '')
	emit('change', 'simple')
	mode.value = 'simple'
	saveMode('simple')
	await nextTick()
	document.getElementById(`${id}-simple`)?.focus()
}

function focus() {
	if (mode.value === 'advanced') {
		advancedInput.value?.focus()
	} else {
		root.value?.querySelector<HTMLInputElement>('input')?.focus()
	}
}

defineExpose({focus})
</script>

<template>
	<div
		ref="root"
		class="filter-editor"
	>
		<div
			class="is-flex mbe-3"
			role="tablist"
			:aria-label="label ?? t('filters.title')"
		>
			<XButton
				v-for="tab in (['simple', 'advanced'] as const)"
				:id="`${id}-${tab}`"
				:key="tab"
				:variant="mode === tab ? 'primary' : 'secondary'"
				:shadow="false"
				class="mie-2"
				role="tab"
				:tabindex="mode === tab ? 0 : -1"
				:aria-selected="mode === tab"
				:aria-controls="`${id}-panel`"
				:aria-disabled="tab === 'simple' && parsed === null"
				@click="selectMode(tab)"
				@keydown.left.prevent="moveTab"
				@keydown.right.prevent="moveTab"
			>
				{{ t(`filters.modes.${tab}`) }}
			</XButton>
		</div>
		<div
			v-if="parsed === null"
			class="message is-warning mbe-2"
			role="status"
		>
			<p>{{ t('filters.modes.tooAdvanced') }}</p>
			<XButton
				data-test="clear-simple"
				variant="secondary"
				@click="clearAndUseSimple"
			>
				{{ t('filters.modes.clear') }}
			</XButton>
		</div>
		<div
			:id="`${id}-panel`"
			role="tabpanel"
			:aria-labelledby="`${id}-${mode}`"
		>
			<template v-if="mode === 'simple'">
				<SimpleFilterForm
					:model-value="simple"
					:project-id="projectId"
					@update:modelValue="changeSimple"
				/>
				<label
					v-if="search !== undefined"
					class="field is-block"
				>
					<span class="label">{{ t('filters.simple.search') }}</span>
					<input
						v-model="searchText"
						type="search"
						class="input"
						@blur="emit('update:search', searchText)"
						@keydown.enter="emit('update:search', searchText)"
					>
				</label>
			</template>
			<template v-else>
				<FilterInput
					ref="advancedInput"
					:model-value="modelValue || search"
					:project-id="projectId"
					class="mbe-2"
					@update:modelValue="changeAdvanced"
				/>
				<FilterInputDocs />
			</template>
		</div>
	</div>
</template>
