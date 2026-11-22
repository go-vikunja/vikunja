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
import FormField from '@/components/input/FormField.vue'

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
// Saved filters and favorites have negative pseudo project ids.
const realProjectId = computed(() => (props.projectId ?? 0) > 0 ? props.projectId : undefined)
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

// Focus lands on the disabled Simple tab too, so keyboard users get its tooltip.
async function moveTab(from: 'simple' | 'advanced') {
	const to = from === 'simple' ? 'advanced' : 'simple'
	selectMode(to)
	await nextTick()
	document.getElementById(`${id}-${to}`)?.focus()
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
		<div class="is-flex is-justify-content-space-between is-align-items-center mbe-3">
			<div
				class="is-flex"
				role="tablist"
				:aria-label="label ?? t('filters.title')"
			>
				<XButton
					v-for="tab in (['simple', 'advanced'] as const)"
					:id="`${id}-${tab}`"
					:key="tab"
					v-tooltip="tab === 'simple' && parsed === null ? t('filters.modes.tooAdvanced') : undefined"
					:variant="mode === tab ? 'primary' : 'secondary'"
					:shadow="false"
					class="mode-tab mie-2"
					role="tab"
					:tabindex="mode === tab ? 0 : -1"
					:aria-selected="mode === tab"
					:aria-controls="`${id}-panel`"
					:aria-disabled="tab === 'simple' && parsed === null"
					@click="selectMode(tab)"
					@keydown.left.prevent="moveTab(tab)"
					@keydown.right.prevent="moveTab(tab)"
				>
					{{ t(`filters.modes.${tab}`) }}
				</XButton>
			</div>
			<XButton
				v-if="parsed === null"
				data-test="clear-simple"
				variant="tertiary"
				:wrap="false"
				@click="clearAndUseSimple"
			>
				{{ t('filters.modes.clear') }}
			</XButton>
		</div>
		<slot name="hint" />
		<div
			:id="`${id}-panel`"
			role="tabpanel"
			:aria-labelledby="`${id}-${mode}`"
		>
			<template v-if="mode === 'simple'">
				<FormField
					v-if="search !== undefined"
					v-model="searchText"
					type="search"
					:label="t('filters.simple.search')"
					@blur="emit('update:search', searchText)"
					@keydown.enter="emit('update:search', searchText)"
				/>
				<SimpleFilterForm
					:model-value="simple"
					:project-id="realProjectId"
					@update:modelValue="changeSimple"
				/>
			</template>
			<template v-else>
				<FilterInput
					ref="advancedInput"
					:model-value="modelValue || search"
					:project-id="realProjectId"
					class="mbe-2"
					@update:modelValue="changeAdvanced"
				/>
				<FilterInputDocs />
			</template>
		</div>
	</div>
</template>

<style lang="scss" scoped>
// Button.vue disables pointer events on aria-disabled, which would block the hover tooltip.
.filter-editor .mode-tab[aria-disabled='true'] {
	pointer-events: auto;
}
</style>
