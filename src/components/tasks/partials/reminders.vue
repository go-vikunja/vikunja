<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r.reminder < new Date()}"
			class="reminder-input"
		>
			<ReminderDetail v-if="!!r.relativePeriod" v-model="reminders[index]" @close-on-change="() => addReminder(index)"
				/>
			<Datepicker
				v-if="!r.relativePeriod"
				v-model="reminders[index].reminder"
				:disabled="disabled"
				@close-on-change="() => addReminderDate(index)"
			/>
			<BaseButton @click="removeReminderByIndex(index)" v-if="!disabled" class="remove">
				<icon icon="times"></icon>
			</BaseButton>
		</div>
		<div class="reminder-input" v-if="!disabled">
			<Datepicker
				v-model="newReminder"
				@close-on-change="() => addReminderDate()"
				:choose-date-label="$t('task.addReminder')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch, type PropType } from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Datepicker from '@/components/input/datepicker.vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'
import TaskReminderModel from '@/models/taskReminder'
import type { ITaskReminder } from '@/modelTypes/ITaskReminder'

const props = defineProps({
	modelValue: {
		type: Array as PropType<ITaskReminder[]>,
		default: () => [],
		validator(prop) {
			// This allows arrays
			return prop instanceof Array
		},
	},
	disabled: {
		type: Boolean,
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
		for (const i in newVal) {
			if (typeof newVal[i].reminder === 'string') {
				newVal[i].reminder = new Date(newVal[i].reminder)
			}
		}
		reminders.value = newVal
	},
)


function updateData() {
	emit('update:modelValue', reminders.value)
}

const newReminder = ref(null)
function addReminder(index: number | null = null) {
	updateData()
}

function addReminderDate(index: number | null = null) {
	// New Date
	if (index === null) {
		if (newReminder.value === null) {
			return
		}
		reminders.value.push(new TaskReminderModel({reminder: new Date(newReminder.value)}))
		newReminder.value = null
	} else if(reminders.value[index].reminder === null) {
		return
	}

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
      padding-left: .5rem;
    }
  }
}
</style>