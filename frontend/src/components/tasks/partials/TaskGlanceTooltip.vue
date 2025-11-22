<template>
	<span
		ref="triggerRef"
		class="task-glance-trigger"
		@mouseenter="handleMouseEnter"
		@mouseleave="handleMouseLeave"
	>
		<slot />
	</span>

	<Teleport to="body">
		<CustomTransition name="fade">
			<div
				v-if="showTooltip"
				ref="tooltipRef"
				class="task-glance-tooltip"
				role="tooltip"
			>
				<div class="task-glance-content">
					<div class="task-glance-header">
						<div class="task-glance-title-section">
							<span class="task-identifier">{{ taskIdentifier }}</span>
							<span class="task-title">{{ task.title }}</span>
						</div>
						<div class="task-glance-indicators">
							<span
								v-if="task.attachments.length > 0"
								class="task-glance-icon"
							>
								<Icon icon="paperclip" />
							</span>
							<span
								v-if="!isEditorContentEmpty(task.description)"
								class="task-glance-icon is-mirrored-rtl"
							>
								<Icon icon="align-left" />
							</span>
							<CommentCount
								:task="task"
								class="task-glance-icon"
							/>
						</div>
					</div>
							
					<ChecklistSummary :task="task" />

					<Labels
						v-if="task.labels.length > 0"
						:labels="task.labels"
						class="task-glance-labels"
					/>

					<div
						v-if="task.dueDate"
						class="task-glance-due"
					>
						<Icon icon="calendar" />
						<span>{{ $t('task.detail.due') }}: {{ formatDisplayDate(task.dueDate) }}</span>
					</div>

					<div class="task-glance-meta">
						<div class="task-glance-created">
							<i18n-t
								keypath="task.detail.created"
								scope="global"
							>
								<span>{{ formatDisplayDate(task.created) }}</span>
								{{ getDisplayName(task.createdBy) }}
							</i18n-t>
						</div>
					</div>
				</div>
			</div>
		</CustomTransition>
	</Teleport>
</template>

<script setup lang="ts">
import {ref, computed, onUnmounted, nextTick} from 'vue'
import {computePosition, flip, offset, shift} from '@floating-ui/dom'

import type {ITask} from '@/modelTypes/ITask'
import {getTaskIdentifier} from '@/models/task'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {getDisplayName} from '@/models/user'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'

import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'
import CommentCount from '@/components/tasks/partials/CommentCount.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

const props = defineProps<{
	task: ITask
}>()

const HOVER_DELAY = 1000 // 1 second

const triggerRef = ref<HTMLElement | null>(null)
const tooltipRef = ref<HTMLElement | null>(null)
const showTooltip = ref(false)
let hoverTimeout: ReturnType<typeof setTimeout> | null = null



const taskIdentifier = computed(() => getTaskIdentifier(props.task))

async function updatePosition() {
	if (!triggerRef.value || !tooltipRef.value) {
		return
	}

	await nextTick()

	const {x, y} = await computePosition(triggerRef.value, tooltipRef.value, {
		strategy: 'absolute',
		placement: 'top',
		middleware: [
			offset(8),
			flip({
				fallbackPlacements: ['bottom', 'top-start', 'top-end', 'bottom-start', 'bottom-end'],
				padding: 8,
			}),
			shift({padding: 8}),
		],
	})

	// Set position directly on the element
	if (tooltipRef.value) {
		tooltipRef.value.style.left = `${x}px`
		tooltipRef.value.style.top = `${y}px`
	}
}

function handleMouseEnter() {
	// Clear any existing timeout
	if (hoverTimeout) {
		clearTimeout(hoverTimeout)
	}

	// Set timeout to show tooltip after 1 second
	hoverTimeout = setTimeout(async () => {
		showTooltip.value = true
		// Wait for the tooltip to be rendered in the DOM
		await nextTick()
		await updatePosition()
	}, HOVER_DELAY)
}

function handleMouseLeave() {
	// Clear timeout if user moves away before 1 second
	if (hoverTimeout) {
		clearTimeout(hoverTimeout)
		hoverTimeout = null
	}

	// Hide tooltip
	showTooltip.value = false
}

// Cleanup on unmount
onUnmounted(() => {
	if (hoverTimeout) {
		clearTimeout(hoverTimeout)
	}
})
</script>

<style lang="scss" scoped>
.task-glance-trigger {
	display: inline;
}

.task-glance-tooltip {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: 0;
	z-index: 9999;
	max-inline-size: 400px;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	box-shadow: var(--shadow-lg);
	pointer-events: none;
}

.task-glance-content {
	padding: 0.75rem;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	font-size: 0.875rem;
	color: var(--text);
}

.task-glance-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
}

.task-glance-title-section {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	flex: 1;
	min-inline-size: 0; /* Allow text to wrap */
}

.task-identifier {
	font-size: 0.75rem;
	color: var(--grey-500);
	font-weight: 600;
}

.task-title {
	font-weight: 600;
	color: var(--text);
	word-wrap: break-word;
}

.task-glance-labels {
	margin: 0;
}

.task-glance-due {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	color: var(--grey-700);

	.icon {
		inline-size: 1rem;
		block-size: 1rem;
	}
}

.task-glance-meta {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	padding-block-start: 0.25rem;
	border-block-start: 1px solid var(--grey-200);
	font-size: 0.8rem;
	color: var(--grey-600);
}

.task-glance-created {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.task-glance-indicators {
	display: flex;
	align-items: center;
	flex-shrink: 0; /* Prevent icons from shrinking */

	:deep(.checklist-summary) {
		padding-inline-end: 6px;

		&:not(:last-child) {
			padding-inline-end: 8px;
		}
	}
}

.task-glance-icon {
	color: var(--grey-500);
	font-size: 0.875rem;
	display: inline-flex;
	align-items: center;
	margin-inline-end: 6px;

	&:not(:last-of-type) {
		margin-inline-end: 8px;
	}
}
</style>
