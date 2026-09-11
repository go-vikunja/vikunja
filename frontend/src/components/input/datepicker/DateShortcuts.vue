<template>
	<div
		class="date-shortcuts"
		:class="`date-shortcuts--${layout}`"
	>
		<div
			v-if="layout === 'list'"
			class="date-shortcuts__heading"
		>
			{{ $t('input.datepicker.quickSelect') }}
		</div>
		<BaseButton
			v-for="shortcut in shortcuts"
			:key="shortcut.key"
			class="datepicker__quick-select-date date-shortcuts__item"
			:class="{'is-active': isSameDay(shortcut.date, active)}"
			@click.stop="emit('select', shortcut.date)"
		>
			<span class="date-shortcuts__icon">
				<Icon :icon="shortcut.icon" />
			</span>
			<span class="date-shortcuts__label">{{ $t(`input.datepicker.${shortcut.key}`) }}</span>
			<span
				v-if="layout === 'list'"
				class="date-shortcuts__weekday"
			>{{ shortcut.weekday }}</span>
		</BaseButton>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'

import {formatDate} from '@/helpers/time/formatDate'
import {isSameDay} from '@/helpers/time/dateMath'
import {buildDateShortcuts, type DateShortcutKey} from '@/helpers/time/dateShortcuts'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

withDefaults(defineProps<{
	layout?: 'list' | 'chips'
	active?: Date | null
}>(), {
	layout: 'list',
	active: null,
})

const emit = defineEmits<{
	select: [date: Date]
}>()

const ICONS: Record<DateShortcutKey, IconProp> = {
	today: ['far', 'calendar-alt'],
	tomorrow: ['far', 'sun'],
	nextMonday: 'coffee',
	thisWeekend: 'cocktail',
	laterThisWeek: 'chess-knight',
	nextWeek: 'forward',
	nextWeekend: 'umbrella-beach',
	endOfMonth: 'flag-checkered',
	inAMonth: 'calendar-plus',
}

const shortcuts = computed(() => buildDateShortcuts().map(shortcut => ({
	...shortcut,
	icon: ICONS[shortcut.key],
	weekday: formatDate(shortcut.date, 'ddd'),
})))
</script>

<style lang="scss" scoped>
.date-shortcuts--list {
	display: flex;
	flex-direction: column;
	gap: .125rem;
	padding: .75rem;
	background: var(--grey-50);
	border-inline-end: 1px solid var(--grey-200);
	inline-size: 13rem;
	flex-shrink: 0;

	.date-shortcuts__heading {
		font-size: .7rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: .05em;
		color: var(--grey-600);
		padding: .125rem .625rem .5rem;
	}

	.date-shortcuts__item {
		display: flex;
		align-items: center;
		gap: .5rem;
		inline-size: 100%;
		padding: .4375rem .625rem;
		border-radius: $radius;
		font-size: .85rem;
		font-weight: 600;
		color: var(--grey-700);
		text-align: start;
		transition: background-color $transition, color $transition;

		&:hover, &.is-active {
			background: var(--grey-100);
		}

		&.is-active {
			color: var(--primary);
			font-weight: 700;

			.date-shortcuts__icon {
				color: var(--primary);
			}
		}
	}

	.date-shortcuts__icon {
		inline-size: 1rem;
		text-align: center;
		font-size: .8rem;
		color: var(--grey-400);
	}

	.date-shortcuts__weekday {
		margin-inline-start: auto;
		color: var(--text-light);
		font-weight: 500;
		font-size: .8rem;
		text-transform: capitalize;
	}
}

.date-shortcuts--chips {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
	padding: .75rem 1rem;
	border-block-end: 1px solid var(--grey-200);

	.bottom-sheet & {
		flex-wrap: nowrap;
		overflow-x: auto;
		scrollbar-width: none;

		&::-webkit-scrollbar {
			display: none;
		}
	}

	.date-shortcuts__item {
		display: inline-flex;
		align-items: center;
		gap: .25rem;
		padding: .3125rem .6875rem;
		border: 1px solid var(--grey-200);
		border-radius: $radius-rounded;
		background: var(--white);
		font-size: .85rem;
		font-weight: 600;
		color: var(--grey-700);
		white-space: nowrap;
		transition: all $transition;

		&:hover {
			background: var(--grey-100);
		}

		&.is-active {
			background: var(--primary);
			border-color: var(--primary);
			color: var(--primary-invert);

			.date-shortcuts__icon {
				opacity: 1;
			}
		}
	}

	.date-shortcuts__icon {
		font-size: .72rem;
		opacity: .65;
	}
}
</style>
