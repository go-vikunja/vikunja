<template>
	<div
		v-if="canWrite || isSet"
		class="property"
		:class="[`property-${name}`, {'is-set': isSet}]"
	>
		<dt :id="`task-property-${name}-label`">
			<Icon :icon="icon" />
			{{ label }}
		</dt>
		<dd>
			<slot />
			<BaseButton
				v-if="removeLabel && isSet && canWrite"
				v-tooltip="removeLabel"
				class="remove"
				:aria-label="removeLabel"
				@click="emit('remove')"
			>
				<Icon icon="times" />
			</BaseButton>
		</dd>
		<slot name="editor" />
	</div>
</template>

<script setup lang="ts">
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import BaseButton from '@/components/base/BaseButton.vue'

defineProps<{
	name: string,
	icon: IconProp,
	label: string,
	isSet: boolean,
	canWrite: boolean,
	removeLabel?: string,
}>()

const emit = defineEmits<{
	remove: [],
}>()
</script>

<style lang="scss" scoped>
.property {
	display: grid;
	grid-template-columns: minmax(6rem, 7.5rem) 1fr;
	column-gap: .5rem;
	align-items: start;
	padding-block: .375rem;
	border-block-end: 1px solid var(--grey-200);
	transition: background-color $transition;

	dt {
		display: flex;
		align-items: center;
		gap: .5rem;
		// Line up the label with the value's first line, whose control has its own vertical padding.
		min-block-size: 2rem;
		color: var(--grey-600);
		font-weight: 500;
		transition: color $transition;

		svg {
			inline-size: .875rem;
			flex-shrink: 0;
			color: var(--grey-400);
			transition: color $transition;
		}
	}

	dd {
		min-inline-size: 0;
		margin: 0;
		display: flex;
		align-items: center;
		gap: .25rem;

		// Not :slotted(*), which skips FormSelect's fragment root.
		> :deep(*) {
			min-inline-size: 0;
		}
	}

	@container task-sidebar (inline-size < 19rem) {
		grid-template-columns: 1fr;

		dt {
			min-block-size: 1.5rem;
		}
	}

	&.is-set dt {
		color: var(--grey-700);

		svg {
			color: var(--primary);
		}
	}

	.remove,
	:deep(.reminders .remove),
	:deep(.remove-assignee) {
		color: var(--grey-500);
		line-height: 1;
		padding: .25rem;
		border-radius: $radius;
		transition: opacity $transition, color $transition;

		&:hover,
		&:focus-visible {
			color: var(--danger);
		}
	}

	@media (hover: hover) {
		.remove,
		:deep(.reminders .remove),
		:deep(.remove-assignee) {
			opacity: 0;
		}

		&:hover,
		&:focus-within {
			.remove,
			:deep(.reminders .remove),
			:deep(.remove-assignee) {
				opacity: 1;
			}
		}
	}

	@media print {
		.remove {
			display: none;
		}
	}
}
</style>
