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
					<div class="select is-fullwidth">
						<select v-model="selected">
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
						{{ $t('sorting.apply') }}
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

const MANUAL = 'position:asc'
const selected = ref<string>(MANUAL)

watch(() => props.modelValue, (val) => {
	const key = Object.keys(val)[0]
	if (!key || key === 'position') {
		selected.value = MANUAL
		return
	}
	const order = (val as Record<string, 'asc' | 'desc'>)[key] ?? 'asc'
	selected.value = `${key}:${order}`
}, {immediate: true})

const options = computed(() => {
	const manual = {value: MANUAL, label: t('sorting.manually')}
	const rest = [
		{value: 'title:asc', label: t('sorting.options.titleAsc')},
		{value: 'title:desc', label: t('sorting.options.titleDesc')},
		{value: 'priority:desc', label: t('sorting.options.priorityDesc')},
		{value: 'priority:asc', label: t('sorting.options.priorityAsc')},
		{value: 'due_date:asc', label: t('sorting.options.dueDateAsc')},
		{value: 'due_date:desc', label: t('sorting.options.dueDateDesc')},
		{value: 'start_date:asc', label: t('sorting.options.startDateAsc')},
		{value: 'start_date:desc', label: t('sorting.options.startDateDesc')},
		{value: 'end_date:asc', label: t('sorting.options.endDateAsc')},
		{value: 'end_date:desc', label: t('sorting.options.endDateDesc')},
		{value: 'percent_done:desc', label: t('sorting.options.percentDoneDesc')},
		{value: 'percent_done:asc', label: t('sorting.options.percentDoneAsc')},
		{value: 'created:desc', label: t('sorting.options.createdDesc')},
		{value: 'created:asc', label: t('sorting.options.createdAsc')},
		{value: 'updated:desc', label: t('sorting.options.updatedDesc')},
		{value: 'updated:asc', label: t('sorting.options.updatedAsc')},
	].sort((a, b) => a.label.localeCompare(b.label))

	return [manual, ...rest]
})

function applySort(close: () => void) {
	const [field, order] = selected.value.split(':') as [string, 'asc' | 'desc']
	const sort: SortBy = {} as SortBy
	;(sort as Record<string, 'asc' | 'desc'>)[field] = order
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
