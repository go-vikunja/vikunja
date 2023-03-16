<template>
	<div class="reminders">
		<div v-for="(r, index) in reminders" :key="index" :class="{ 'overdue': r.reminder < new Date() }" class="reminder-input">
			<div class="reminder-detail">
				<ReminderDetail :disabled="disabled" v-model="reminders[index]" @update:modelValue="() => editReminder(index)"/>
			</div>
			<div>
				<BaseButton @click="removeReminderByIndex(index)" v-if="!disabled" class="remove">
					<icon icon="times"></icon>
				</BaseButton>
			</div>
		</div>
		<div class="reminder-input">
			<BaseButton @click.stop="toggleAddReminder"  v-if="!disabled">
				{{ $t('task.addReminder') }}
			</BaseButton>
		</div>
		<div class="reminder-input">
			<ReminderDetail v-if="isAddReminder" :disabled="disabled" @update:modelValue="addNewReminder"/>
		</div>
	</div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'
import TaskReminderModel from '@/models/taskReminder'
import type { ITaskReminder } from '@/modelTypes/ITaskReminder'
import { onMounted, reactive, ref, watch, type PropType } from 'vue'

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
	isAddReminder.value = false
	emit('update:modelValue', reminders.value)
}

function editReminder(index: number) {
	if (reminders.value[index] === null) {
		return
	}
	updateData()
}

function addNewReminder(newReminder : ITaskReminder) {
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

		.reminder-detail {
			width: 100%;
		}
		.remove {
			color: var(--danger);
			vertical-align: top;
			padding-left: .5rem;
			line-height: 1;
		}
	}
}
</style>