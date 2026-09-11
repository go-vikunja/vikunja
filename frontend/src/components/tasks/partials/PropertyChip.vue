<template>
	<div
		class="property-chip"
		:class="{'is-set': isSet, 'is-open': open}"
	>
		<Popup
			:open="open"
			@update:open="v => open = v"
		>
			<template #trigger="{toggle}">
				<BaseButton
					class="chip"
					:aria-expanded="open"
					:aria-label="label"
					:disabled="disabled"
					@click="open ? toggle() : openAndFocus()"
				>
					<Icon
						:icon="icon"
						class="chip-icon"
					/>
					<span class="chip-value">
						<slot
							v-if="isSet"
							name="value"
						/>
						<template v-else>
							{{ label }}
						</template>
					</span>
				</BaseButton>
			</template>
			<template #content="{close}">
				<div
					ref="popover"
					class="chip-popover"
					role="dialog"
					:aria-label="label"
					@keydown.esc.stop="close()"
				>
					<div class="chip-popover-header">
						<span>{{ label }}</span>
						<BaseButton
							v-if="clearable && isSet"
							class="chip-clear"
							@click="() => { emit('clear'); close() }"
						>
							{{ $t('misc.clear') }}
						</BaseButton>
					</div>
					<div class="chip-popover-body">
						<slot />
					</div>
				</div>
			</template>
		</Popup>
	</div>
</template>

<script setup lang="ts">
import {nextTick, ref} from 'vue'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import BaseButton from '@/components/base/BaseButton.vue'
import Popup from '@/components/misc/Popup.vue'

withDefaults(defineProps<{
	icon: IconProp
	label: string
	isSet: boolean
	clearable?: boolean
	disabled?: boolean
}>(), {
	clearable: false,
	disabled: false,
})

const emit = defineEmits<{
	'clear': []
}>()

const open = ref(false)
const popover = ref<HTMLElement | null>(null)

const FOCUSABLE = 'input, select, textarea, button, [tabindex]:not([tabindex="-1"])'

async function openAndFocus() {
	open.value = true
	await nextTick()
	const el = popover.value?.querySelector('.chip-popover-body')?.querySelector<HTMLElement>(FOCUSABLE)
	el?.focus()
	// A date field's only control is the button that opens the calendar; open it right away.
	// Deferred so the click that opened the chip has finished bubbling (the calendar closes on outside clicks).
	if (el?.matches('.datepicker .show')) {
		setTimeout(() => el.click())
	}
}

defineExpose({open: openAndFocus})
</script>

<style scoped lang="scss">
.property-chip {
	position: relative;
	display: inline-flex;
}

.chip {
	display: inline-flex;
	align-items: center;
	gap: .375rem;
	max-inline-size: 100%;
	min-block-size: 1.875rem;
	padding: .25rem .625rem;
	border: 1px solid var(--grey-200);
	border-radius: 999px;
	background: var(--scheme-main);
	color: var(--grey-500);
	font-size: .875rem;
	line-height: 1.25;
	white-space: nowrap;
	transition: color $transition, border-color $transition, background-color $transition, box-shadow $transition;

	.chip-icon {
		inline-size: .8125rem;
		flex-shrink: 0;
		color: var(--grey-400);
		transition: color $transition;
	}

	&:hover,
	&:focus-visible {
		color: var(--text);
		border-color: var(--grey-300);
	}

	.is-set & {
		color: var(--text);

		.chip-icon {
			color: var(--primary);
		}
	}

	.is-open & {
		border-color: var(--primary);
		box-shadow: 0 0 0 3px hsla(var(--primary-h), var(--primary-s), var(--primary-l), .15);
	}
}

.chip-value {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	min-inline-size: 0;
	overflow: hidden;
	text-overflow: ellipsis;
}

.chip-popover {
	inline-size: 20rem;
	max-inline-size: calc(100vw - 2rem);
	margin-block-start: .25rem;
	border-radius: $radius;
	background: var(--scheme-main);
	box-shadow: var(--shadow-md);
	border: 1px solid var(--grey-200);
}

.chip-popover-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: .5rem .75rem;
	border-block-end: 1px solid var(--grey-200);
	font-size: .75rem;
	font-weight: 600;
	color: var(--grey-500);
}

.chip-clear {
	font-size: .75rem;
	font-weight: 500;
	color: var(--grey-500);

	&:hover {
		color: var(--danger);
	}
}

.chip-popover-body {
	padding: .5rem .75rem;
}
</style>
