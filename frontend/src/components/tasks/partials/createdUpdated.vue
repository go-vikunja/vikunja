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
				<span>{{ formatDateSince(task.created) }}</span>
				{{ getDisplayName(task.createdBy) }}
			</i18n-t>
		</time>
		<template v-if="+new Date(task.created) !== +new Date(task.updated)">
			<br>
			<!-- Computed properties to show the actual date every time it gets updated -->
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
import {computed, toRefs, type PropType} from 'vue'
import type {ITask} from '@/modelTypes/ITask'
import {formatISO, formatDateLong, formatDateSince} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'

const props = defineProps({
	task: {
		type: Object as PropType<ITask>,
		required: true,
	},
})

const {task} = toRefs(props)

const updatedSince = computed(() => formatDateSince(task.value.updated))
const updatedFormatted = computed(() => formatDateLong(task.value.updated))
const doneSince = computed(() => formatDateSince(task.value.doneAt))
const doneFormatted = computed(() => formatDateLong(task.value.doneAt))
</script>

<style lang="scss" scoped>
.created {
	font-size: .75rem;
	color: var(--grey-500);
	text-align: right;
}
</style>