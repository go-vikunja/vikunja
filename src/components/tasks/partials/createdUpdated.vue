<template>
	<p class="created">
		<time :datetime="formatISO(task.created)" v-tooltip="formatDate(task.created)">
			<i18n-t keypath="task.detail.created">
				<span>{{ formatDateSince(task.created) }}</span>
				{{ task.createdBy.getDisplayName() }}
			</i18n-t>
		</time>
		<template v-if="+new Date(task.created) !== +new Date(task.updated)">
			<br/>
			<!-- Computed properties to show the actual date every time it gets updated -->
			<time :datetime="formatISO(task.updated)" v-tooltip="updatedFormatted">
				<i18n-t keypath="task.detail.updated">
					<span>{{ updatedSince }}</span>
				</i18n-t>
			</time>
		</template>
		<template v-if="task.done">
			<br/>
			<time :datetime="formatISO(task.doneAt)" v-tooltip="doneFormatted">
				<i18n-t keypath="task.detail.doneAt">
					<span>{{ doneSince }}</span>
				</i18n-t>
			</time>
		</template>
	</p>
</template>

<script lang="ts" setup>
import {computed, toRefs} from 'vue'
import TaskModel from '@/models/task'
import {formatDateLong, formatDateSince} from '@/helpers/time/formatDate'

const props = defineProps({
	task: {
		type: TaskModel,
		required: true,
	},
})

const {task} = toRefs(props)

const updatedSince = computed(() => formatDateSince(task.value.updated))
const updatedFormatted = computed(() => formatDateLong(task.value.updated))
const doneSince = computed(() => formatDateSince(task.value.doneAt))
const doneFormatted = computed(() => formatDateLong(task.value.doneAt))
</script>
