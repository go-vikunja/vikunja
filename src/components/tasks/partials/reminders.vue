<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r < new Date()}"
			class="reminder-input"
		>
			<Datepicker
				v-model="reminders[index]"
				:disabled="disabled"
				@close-on-change="() => addReminderDate(index)"
			/>
			<a @click="removeReminderByIndex(index)" v-if="!disabled" class="remove">
				<icon icon="times"></icon>
			</a>
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
import {PropType, ref, onMounted, watch} from 'vue'

import Datepicker from '@/components/input/datepicker.vue'

type Reminder = Date | string

const props = defineProps({
	modelValue: {
		type: Array as PropType<Reminder[]>,
		default: () => [],
		validator(prop) {
			// This allows arrays of Dates and strings
			if (!(prop instanceof Array)) {
				return false
			}

			const isDate = (e: any) => e instanceof Date
			const isString = (e: any) => typeof e === 'string'

			for (const e of prop) {
				if (!isDate(e) && !isString(e)) {
					console.log('validation failed', e, e instanceof Date)
					return false
				}
			}

			return true
		},
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'change'])

const reminders = ref<Reminder[]>([])

onMounted(() => {
	reminders.value = props.modelValue
})

watch(
	() => props.modelValue,
	(newVal) => {
		for (const i in newVal) {
			if (typeof newVal[i] === 'string') {
				newVal[i] = new Date(newVal[i])
			}
		}
		reminders.value = newVal
	},
)


function updateData() {
	emit('update:modelValue', reminders.value)
	emit('change')
}

const newReminder = ref(null)
function addReminderDate(index : number | null = null) {
	// New Date
	if (index === null) {
		if (newReminder.value === null) {
			return
		}
		reminders.value.push(new Date(newReminder.value))
		newReminder.value = null
	} else if(reminders.value[index] === null) {
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

    &.overdue :deep(.datepicker a.show) {
      color: var(--danger);
    }

    &:last-child {
      margin-bottom: 0.75rem;
    }

    a.remove {
      color: var(--danger);
      padding-left: .5rem;
    }
  }
}
</style>