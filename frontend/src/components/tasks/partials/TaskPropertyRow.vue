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

		// Slotted controls carry the parent's scope id, not this one.
		> :deep(*) {
			min-inline-size: 0;
		}
	}

	// Between tablet and desktop the column is too narrow for label and value side by side.
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

	&:not(.is-set) > dd :deep(select) {
		color: var(--grey-600);
		font-style: italic;
	}
}

.property > dd {
	:deep(.select) {
		inline-size: 100%;
		// Bulma sizes the wrapper to --input-height; the margin matches the date rows' SimpleButton.
		block-size: auto;
		margin-block: .1rem;

		select {
			inline-size: 100%;
			background: transparent;
			border-color: transparent;
			box-shadow: none;
			block-size: 2rem;
			padding-block: 0;
			padding-inline-start: .5rem;
			line-height: 2rem;
		}

		&:not(.has-defaults):after {
			opacity: 0;
		}

		&:hover:not(.has-defaults):after {
			opacity: 1;
		}

		select:disabled {
			color: var(--text);
			opacity: 1;
			cursor: default;
		}

		&:has(select:disabled)::after {
			display: none;
		}

		select:not(:disabled) {
			cursor: pointer;

			&:hover,
			&:focus {
				background: var(--scheme-main);
			}

			&:focus-visible {
				border-color: var(--primary);
				@include focus-ring;
			}
		}
	}

	:deep(.multiselect) {
		inline-size: 100%;
		min-inline-size: 0;

		input {
			flex: 1 1 5rem;
			min-inline-size: 5rem;
			inline-size: auto;
			text-overflow: ellipsis;
		}

		.tag {
			max-inline-size: 100%;

			> span {
				min-inline-size: 0;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}

		.input-wrapper:not(:focus-within) {
			// Beats Multiselect's !important hover border.
			border-color: transparent !important;

			&:not(:hover) {
				background: transparent;
			}
		}
	}

	// The ring goes on the field, not the borderless input inside it.
	.user-is-tabbing & :deep(.multiselect) .input-wrapper:focus-within {
		@include focus-ring;

		.input {
			box-shadow: none;
		}
	}

	:deep(.datepicker) {
		flex: 1;
		min-inline-size: 0;

		.show {
			color: var(--text);
			transition: background-color $transition;
			border-radius: $radius;
			display: block;
			inline-size: 100%;
			min-block-size: 2rem;
			text-align: start;

			i {
				color: var(--grey-600);
			}

			&:hover:not(:disabled) {
				background: var(--scheme-main);
			}
		}
	}

	// The selects' border, so keyboard focus looks the same without moving the text.
	:deep(.datepicker) .show,
	:deep(.reminders .simple-button:not(.option-button)),
	:deep(.property-value-button) {
		border: 1px solid transparent;
		padding: calc(.25rem - 1px) calc(.5rem - 1px);

		&:not(:disabled):focus-visible {
			border-color: var(--primary);
			background: var(--scheme-main);
		}
	}

	:deep(.reminders) {
		inline-size: 100%;

		.reminder-input:last-child {
			margin-block-end: 0;
		}
	}

	:deep(.color-picker-container) {
		justify-content: flex-start;
		min-block-size: 2rem;
		margin-block: .1rem;
		padding-inline-start: .5rem;

		.picker:has(:focus-visible) {
			border-color: var(--primary);
			@include focus-ring;
		}
	}

	:deep(.color-bubble) {
		inline-size: 1.25rem;
		block-size: 1.25rem;
		margin: calc(.375rem + .1rem) .5rem;
	}

	// 1px border + .25rem padding + 1.375rem avatar = the other rows' 2rem controls.
	:deep(.edit-assignees) {
		.input-wrapper {
			gap: .25rem;
			margin-block: .1rem;
			padding: .25rem .5rem;
		}

		input {
			block-size: 1.375rem;
			padding: 0;
		}
	}

	// The placeholder alone keeps the other rows' text inset.
	:deep(.edit-assignees.has-assignees) .input-wrapper {
		padding-inline-start: .25rem;
	}

	// Read-only list, without the Multiselect wrapper.
	> :deep(.assignees-list) {
		min-block-size: 2rem;
		margin-block: .1rem;
		// The editable field's border + padding.
		padding-inline: calc(.25rem + 1px);
		align-items: center;
	}

	:deep(.assignees-list) {
		flex-wrap: wrap;
		gap: .25rem;
		transition: margin-inline-start $transition;

		// Otherwise the wrapper's gap still pushes the input away.
		&:empty {
			display: none;
		}

		.assignee {
			// The task view's `*` opacity transition overrides AssigneeList's.
			transition: margin-inline-start $transition;

			&:not(:first-child) {
				margin-inline-start: 0;
			}
		}

		.avatar-wrapper {
			margin-inline-end: 0;
		}

		.avatar {
			inline-size: 1.375rem;
			block-size: 1.375rem;
		}

		.remove-assignee {
			inset-block-start: -.25rem;
			inset-inline: auto -.25rem;
			inline-size: .75rem;
			block-size: .75rem;
			padding: 0;
			font-size: .5rem;
			box-shadow: var(--shadow-sm);
		}

		.user {
			margin: 0;
		}

		.username {
			display: none;
		}
	}
}

@media (hover: hover) {
	// The list's margin offsets the avatars', so a wrapped line doesn't start outside the field.
	.property:not(:hover, :focus-within) > dd :deep(.assignees-list) {
		margin-inline-start: calc(.4375rem + .25rem);

		.assignee {
			margin-inline-start: calc(-.4375rem - .25rem);
		}
	}
}
</style>
