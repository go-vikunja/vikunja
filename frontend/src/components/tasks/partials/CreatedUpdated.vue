<template>
	<p class="created">
		<time
			v-tooltip="formatDateLong(task.created)"
			:datetime="formatISO(task.created)"
		>
			<i18n-t
				keypath="task.detail.created"
				scope="global"
			>
				<span>{{ formatDisplayDate(task.created) }}</span>
				{{ getDisplayName(task.createdBy) }}
			</i18n-t>
		</time>
		<template v-if="+new Date(task.created) !== +new Date(task.updated)">
			<br>
			<time
				v-tooltip="updatedFormatted"
				:datetime="formatISO(task.updated)"
			>
				<i18n-t
					keypath="task.detail.updated"
					scope="global"
				>
					<span>{{ updatedSince }}</span>
				</i18n-t>
			</time>
		</template>
		<template v-if="task.done">
			<br>
			<time
				v-tooltip="doneFormatted"
				:datetime="formatISO(task.doneAt)"
			>
				<i18n-t
					keypath="task.detail.doneAt"
					scope="global"
				>
					<span>{{ doneSince }}</span>
				</i18n-t>
			</time>
		</template>
	</p>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import type {ITask} from '@/modelTypes/ITask'
import {formatISO, formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'

const props = defineProps<{
	task: ITask,
}>()

// Computed properties to show the actual date every time it gets updated
const updatedSince = computed(() => formatDisplayDate(props.task.updated))
const updatedFormatted = computed(() => formatDateLong(props.task.updated))
const doneSince = computed(() => formatDisplayDate(props.task.doneAt))
const doneFormatted = computed(() => formatDateLong(props.task.doneAt))
</script>

<style lang="scss" scoped>
.created {
	font-size: .75rem;
	color: var(--grey-500);
	text-align: end;
}
</style>
