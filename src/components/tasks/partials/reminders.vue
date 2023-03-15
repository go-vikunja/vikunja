<template>
	<div class="reminders">
		<div v-for="(r, index) in reminders" :key="index" :class="{ 'overdue': r.reminder < new Date() }" class="reminder-input">
			<div>
				<ReminderDetail :disabled="disabled" v-model="reminders[index]" mode="edit"
												@update:modelValue="() => editReminder(index)"/>
			</div>
			<div>
				<BaseButton @click="removeReminderByIndex(index)" v-if="!disabled" class="remove">
					<icon icon="times"></icon>
				</BaseButton>
			</div>
		</div>
		<div class="reminder-input" v-if="!disabled">
			<BaseButton @click.stop="toggleAddReminder" class="show" :disabled="disabled || undefined">
				{{ $t('task.addReminder') }}
			</BaseButton>

			<CustomTransition name="fade">
				<ReminderDetail v-if="isAddReminder" :disabled="disabled" mode="add" @update:modelValue="addNewReminder"/>
			</CustomTransition>
		</div>
	</div>
</template>

<script setup lang="ts">
import {onMounted, reactive, ref, watch, type PropType} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'
import TaskReminderModel from '@/models/taskReminder'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

const props = defineProps({
	modelValue: {
		type: Array as PropType<ITaskReminder[]>,
		default: () => [],
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const reminders = ref<ITaskReminder[]>([])

onMounted(() => {
	reminders.value = [...props.modelValue]
})

watch(
		() => props.modelValue,
		(newVal) => {
			reminders.value = newVal
		},
)

const isAddReminder = ref(false)
function toggleAddReminder() {
	isAddReminder.value = !isAddReminder.value
}

function updateData() {
	console.log('reminders.updateData', reminders.value)
	emit('update:modelValue', reminders.value)
	isAddReminder.value = false
}

function editReminder(index: number) {
	console.log('reminders.editReminder', reminders.value[index])
	if (reminders.value[index] === null) {
		return
	}
	updateData()
}

function addNewReminder(newReminder) {
	console.log('reminders.addNewReminder to old reminders', newReminder)
	if (newReminder == null) {
		return
	}
	reminders.value.push(newReminder)
	newReminder = reactive(new TaskReminderModel())
	updateData()
}


function removeReminderByIndex(index: number) {
	reminders.value.splice(index, 1)
	updateData()
}
</script>

<style lang="scss" scoped>
.reminders {
	.reminder-input {
		display: flex;
		align-items: center;

		&.overdue :deep(.datepicker .show) {
			color: var(--danger);
		}

		&:last-child {
			margin-bottom: 0.75rem;
		}

		.remove {
			color: var(--danger);
			padding-left: 2rem;
		}
	}
}

</style>