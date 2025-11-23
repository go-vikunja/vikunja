<template>
	<span
		v-if="task.commentCount && task.commentCount > 0"
		v-tooltip="tooltip"
		class="comment-count"
		:class="{'is-unread': task.isUnread}"
	>
		<Icon :icon="['far', 'comments']" />
		<span class="comment-count-badge">{{ task.commentCount }}</span>
		<span
			v-if="task.isUnread"
			class="unread-indicator"
		/>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'

const props = defineProps<{
	task: ITask
}>()

const {t} = useI18n({useScope: 'global'})

const tooltip = computed(() => t('task.attributes.comment', props.task.commentCount))
</script>

<style scoped lang="scss">
.comment-count {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	font-size: 0.875rem;
	color: var(--grey-500);

	.comment-count-badge {
		font-weight: 600;
		font-size: 0.75rem;
		line-height: 1;
	}

	&:hover {
		color: var(--primary);
	}

	&.is-unread {
		font-weight: 600;
		color: var(--primary);

		.unread-indicator {
			display: inline-block;
			inline-size: 6px;
			block-size: 6px;
			background-color: var(--primary);
			border-radius: 50%;
			margin-inline-start: 2px;
			animation: pulse 2s infinite;
		}
	}
}

@keyframes pulse {
	0%, 100% {
		opacity: 1;
	}
	50% {
		opacity: 0.6;
	}
}
</style>

