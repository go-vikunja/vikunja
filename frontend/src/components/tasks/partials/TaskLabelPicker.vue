<template>
	<Popup
		v-model:open="open"
		placement="bottom-end"
		:anchor="triggerElement"
		sheet-on-mobile
		:sheet-title="$t('task.label.edit')"
	>
		<template #trigger="{toggle, isOpen}">
			<BaseButton
				ref="trigger"
				v-tooltip="$t('task.label.edit')"
				class="label-trigger"
				:aria-label="$t('task.label.edit')"
				:aria-expanded="isOpen"
				@click.stop="toggle()"
			>
				<Icon icon="tags" />
			</BaseButton>
		</template>

		<template #content="{isOpen}">
			<div
				v-if="isOpen"
				class="label-picker"
				@keyup.enter.stop
				@click.stop
			>
				<p v-if="isPending">
					{{ $t('misc.loading') }}
				</p>
				<p v-else-if="isError && labels.length === 0">
					{{ $t('task.label.loadError') }}
				</p>
				<p v-else-if="labels.length === 0">
					{{ $t('label.newCTA') }}
				</p>
				<BaseButton
					v-for="label in labels"
					:key="label.id"
					class="label-option"
					:aria-pressed="assignedIds.has(label.id)"
					:aria-disabled="saving"
					@click="toggleLabel(label)"
				>
					<span class="checkmark">
						<Icon
							v-if="assignedIds.has(label.id)"
							icon="check"
						/>
					</span>
					<LabelChip :label="label" />
				</BaseButton>
			</div>
		</template>
	</Popup>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import BaseButton from '@/components/base/BaseButton.vue'
import Popup from '@/components/misc/Popup.vue'
import LabelChip from '@/components/tasks/partials/Label.vue'
import type {Label} from '@/client/generated'
import {labelsQuery} from '@/client/queries/labels'
import {useAddTaskLabelMutation, useRemoveTaskLabelMutation} from '@/client/queries/taskMutations'

const props = defineProps<{
	taskId: number
	assignedLabels: Label[]
}>()

const emit = defineEmits<{
	updated: []
}>()

const open = ref(false)
const trigger = ref<InstanceType<typeof BaseButton> | null>(null)
const triggerElement = computed<HTMLElement | null>(() => trigger.value?.$el ?? null)
const {data, isPending, isError} = useQuery({...labelsQuery(), enabled: open})
const labels = computed(() => (data.value ?? []).filter(
	(label): label is Label & {id: number} => label.id !== undefined,
))
const assignedIds = computed(() => new Set(props.assignedLabels.map(label => label.id)))

const addLabel = useAddTaskLabelMutation()
const removeLabel = useRemoveTaskLabelMutation()
const saving = computed(() => addLabel.isPending.value || removeLabel.isPending.value)

async function toggleLabel(label: Label & {id: number}) {
	const mutation = assignedIds.value.has(label.id) ? removeLabel : addLabel
	try {
		await mutation.mutateAsync({taskId: props.taskId, label})
		emit('updated')
	} catch {
		// The mutation reports the error and leaves the assigned labels intact.
	}
}
</script>

<style lang="scss" scoped>
.label-trigger {
	flex: none;
	inline-size: 27px;
	border-radius: $radius;
	transition: opacity $transition, color $transition;

	&:hover,
	&[aria-expanded='true'] {
		color: var(--primary);
	}
}

.label-picker {
	min-inline-size: min(12rem, 100%);
	max-inline-size: min(22rem, calc(100vw - 16px));
	padding: .5rem;
	background: var(--scheme-main);
	border: 1px solid var(--border);
	border-radius: $radius;
}

.label-option {
	display: flex;
	align-items: center;
	gap: .5rem;
	inline-size: 100%;
	padding: .375rem;
	border-radius: $radius;

	&:hover {
		background: var(--grey-100);
	}

	:deep(.tag) {
		block-size: auto;
		white-space: normal;
		overflow-wrap: anywhere;
		text-align: start;
	}
}

.checkmark {
	flex: none;
	inline-size: 1rem;
}
</style>
