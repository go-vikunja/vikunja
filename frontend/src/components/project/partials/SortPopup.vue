<template>
	<XButton
		variant="secondary"
		icon="sort"
		@click="() => modalOpen = true"
	>
		{{ $t('project.list.sort') }}
	</XButton>
	<Modal
		:enabled="modalOpen"
		transition-name="fade"
		variant="hint-modal"
		@close="() => modalOpen = false"
	>
		<div class="sort-popup">
			<div class="field">
				<label class="label">{{ $t('sorting.sortBy') }}</label>
				<div class="select is-fullwidth">
					<select v-model="sortField">
						<option
							v-for="o in options"
							:key="o.value"
							:value="o.value"
						>
							{{ o.label }}
						</option>
					</select>
				</div>
			</div>
			<div class="field">
				<label class="label">{{ $t('sorting.order') }}</label>
				<div class="select is-fullwidth">
					<select v-model="sortOrder">
						<option value="asc">
							{{ $t('sorting.asc') }}
						</option>
						<option value="desc">
							{{ $t('sorting.desc') }}
						</option>
					</select>
				</div>
			</div>
			<div class="actions">
				<XButton
					variant="tertiary"
					class="has-text-danger"
					@click="modalOpen = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					@click="applySort"
				>
					{{ $t('misc.doit') }}
				</XButton>
			</div>
		</div>
	</Modal>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import XButton from '@/components/input/Button.vue'
import Modal from '@/components/misc/Modal.vue'
import type {SortBy} from '@/composables/useTaskList'

const props = defineProps<{ modelValue: SortBy }>()
const emit = defineEmits<{ 'update:modelValue': [value: SortBy] }>()

const {t} = useI18n({useScope: 'global'})

const modalOpen = ref(false)
const sortField = ref<string>('position')
const sortOrder = ref<'asc' | 'desc'>('asc')

watch(() => props.modelValue, (val) => {
	const key = Object.keys(val)[0] || 'position'
	sortField.value = key
	sortOrder.value = (val as SortBy)[key as keyof SortBy] ?? 'asc'
}, {immediate: true})

const options = [
	{value: 'position', label: t('sorting.position')},
	{value: 'title', label: t('task.attributes.title')},
	{value: 'priority', label: t('task.attributes.priority')},
	{value: 'due_date', label: t('task.attributes.dueDate')},
	{value: 'start_date', label: t('task.attributes.startDate')},
	{value: 'end_date', label: t('task.attributes.endDate')},
	{value: 'percent_done', label: t('task.attributes.percentDone')},
	{value: 'created', label: t('task.attributes.created')},
	{value: 'updated', label: t('task.attributes.updated')},
]

function applySort() {
	const sort: SortBy = {} as SortBy
        ;(sort as Record<string, 'asc' | 'desc'>)[sortField.value] = sortOrder.value
	emit('update:modelValue', sort)
	modalOpen.value = false
}
</script>

<style scoped lang="scss">
.sort-popup {
        display: flex;
        flex-direction: column;
        align-items: stretch;

        .field {
                margin-bottom: 1rem;
        }

        .actions {
                display: flex;
                justify-content: flex-end;
                gap: .5rem;
        }
}
</style>
