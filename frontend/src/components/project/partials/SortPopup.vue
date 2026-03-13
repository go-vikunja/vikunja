<template>
	<Popup>
		<template #trigger="{toggle}">
			<XButton
				variant="secondary"
				icon="sort"
				@click.prevent.stop="toggle()"
			>
				{{ $t('project.list.sort') }}
			</XButton>
		</template>
		<template #content="{close}">
			<Card class="sort-popup">
				<p class="sort-description has-text-grey is-size-7">
					{{ $t('sorting.description') }}
				</p>
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
				<div
					v-if="!isManualSort"
					class="field"
				>
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
						@click="close()"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="primary"
						@click="applySort(close)"
					>
						{{ $t('misc.doit') }}
					</XButton>
				</div>
			</Card>
		</template>
	</Popup>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import XButton from '@/components/input/Button.vue'
import Popup from '@/components/misc/Popup.vue'
import Card from '@/components/misc/Card.vue'
import type {SortBy} from '@/composables/useTaskList'

const props = defineProps<{ modelValue: SortBy }>()
const emit = defineEmits<{ 'update:modelValue': [value: SortBy] }>()

const {t} = useI18n({useScope: 'global'})

const sortField = ref<string>('position')
const sortOrder = ref<'asc' | 'desc'>('asc')

const isManualSort = computed(() => sortField.value === 'position')

watch(() => props.modelValue, (val) => {
	const key = Object.keys(val)[0] || 'position'
	sortField.value = key
	sortOrder.value = (val as SortBy)[key as keyof SortBy] ?? 'asc'
}, {immediate: true})

const options = computed(() => {
	const manualOption = {value: 'position', label: t('sorting.manually')}
	const otherOptions = [
		{value: 'title', label: t('task.attributes.title')},
		{value: 'priority', label: t('task.attributes.priority')},
		{value: 'due_date', label: t('task.attributes.dueDate')},
		{value: 'start_date', label: t('task.attributes.startDate')},
		{value: 'end_date', label: t('task.attributes.endDate')},
		{value: 'percent_done', label: t('task.attributes.percentDone')},
		{value: 'created', label: t('task.attributes.created')},
		{value: 'updated', label: t('task.attributes.updated')},
	].sort((a, b) => a.label.localeCompare(b.label))

	return [manualOption, ...otherOptions]
})

function applySort(close: () => void) {
	const sort: SortBy = {} as SortBy
	const order = isManualSort.value ? 'asc' : sortOrder.value
	;(sort as Record<string, 'asc' | 'desc'>)[sortField.value] = order
	emit('update:modelValue', sort)
	close()
}
</script>

<style scoped lang="scss">
.sort-popup {
	margin: 0;
	min-inline-size: 18rem;

	:deep(.card-content .content) {
		display: flex;
		flex-direction: column;
	}

	.sort-description {
		margin-block-end: 1rem;
	}

	.field {
		margin-block-end: 1rem;
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: .5rem;
	}
}
</style>
