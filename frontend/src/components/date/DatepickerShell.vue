<template>
	<div
		ref="container"
		class="datepicker-with-range-container"
	>
		<Popup
			:open="open"
			placement="bottom-start"
			:anchor="anchor ?? container"
			sheet-on-mobile
			:sheet-title="sheetTitle"
			@update:open="isOpen => $emit('update:open', isOpen)"
		>
			<template #trigger="{toggle}">
				<slot
					name="trigger"
					:toggle="toggle"
				/>
			</template>
			<template #content="{isOpen}">
				<div
					class="datepicker-with-range"
					:class="{'is-open': isOpen}"
				>
					<div class="selections">
						<BaseButton
							v-for="selection in selections"
							:key="selection.key"
							:class="{'is-active': selection.active}"
							@click="selection.onSelect()"
						>
							{{ selection.label }}
						</BaseButton>
					</div>
					<div class="calendar-container input-group">
						<slot />

						<p>
							{{ $t('input.datemathHelp.canuse') }}
						</p>

						<BaseButton
							class="has-text-primary"
							@click="showHowItWorks = true"
						>
							{{ $t('input.datemathHelp.learnhow') }}
						</BaseButton>

						<Modal
							:enabled="showHowItWorks"
							:overflow="true"
							variant="hint-modal"
							@close="() => showHowItWorks = false"
						>
							<DatemathHelp />
						</Modal>
					</div>
				</div>
			</template>
		</Popup>
	</div>
</template>

<script lang="ts" setup>
import {ref} from 'vue'

import Popup from '@/components/misc/Popup.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import DatemathHelp from '@/components/date/DatemathHelp.vue'

interface DatepickerSelection {
	key: string
	label: string
	active: boolean
	onSelect: () => void
}

withDefaults(defineProps<{
	selections: DatepickerSelection[]
	sheetTitle: string
	open?: boolean
	anchor?: HTMLElement | null
}>(), {
	open: false,
	anchor: null,
})

defineEmits<{
	'update:open': [open: boolean],
}>()

defineSlots<{
	trigger(props: {toggle: () => boolean}): void
	default(): void
}>()

// Fallback anchor: the wrapper only ever wraps the trigger, so it is the same box.
const container = ref<HTMLElement | null>(null)

const showHowItWorks = ref(false)
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;
}

:deep(.popup) {
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	background-color: var(--white);
	box-shadow: $shadow;

	&.is-open {
		inline-size: 560px;
		max-inline-size: calc(100vw - 2rem);
	}
}

.datepicker-with-range {
	display: flex;
	inline-size: 100%;
}

.calendar-container {
	inline-size: 70%;
	border-inline-start: 1px solid var(--grey-200);
	padding: 1rem;
	font-size: .9rem;

	:deep(.label),
	:deep(.input) {
		font-size: .9rem;
	}
}

.selections {
	inline-size: 30%;
	display: flex;
	flex-direction: column;
	padding-block-start: .5rem;
	overflow-y: auto;
	max-block-size: 30rem;

	button {
		display: block;
		inline-size: 100%;
		text-align: start;
		padding: .5rem 1rem;
		transition: $transition;
		font-size: .9rem;
		color: var(--text);
		background: transparent;
		border: 0;
		cursor: pointer;

		&.is-active {
			color: var(--primary);
		}

		&:hover, &.is-active {
			background-color: var(--grey-100);
		}
	}
}

.bottom-sheet .datepicker-with-range {
	flex-direction: column;
}

.bottom-sheet .selections {
	flex-direction: row;
	flex-wrap: nowrap;
	inline-size: 100%;
	max-block-size: none;
	gap: .5rem;
	padding: .75rem 1rem;
	overflow-x: auto;
	overflow-y: hidden;
	border-block-end: 1px solid var(--grey-200);
	scrollbar-width: none;

	&::-webkit-scrollbar {
		display: none;
	}

	button {
		inline-size: auto;
		padding: .3125rem .6875rem;
		border: 1px solid var(--grey-200);
		border-radius: $radius-rounded;
		white-space: nowrap;
		font-weight: 600;

		&.is-active {
			background: var(--primary);
			border-color: var(--primary);
			color: var(--primary-invert);
		}
	}
}

.bottom-sheet .calendar-container {
	inline-size: 100%;
	border-inline-start: 0;
}
</style>
