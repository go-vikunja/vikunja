<template>
	<div class="filter-autocompletes">
		<template v-if="items.length">
			<button
				v-for="(item, index) in items"
				:key="`${item.fieldType}-${item.id}`"
				class="filter-autocomplete"
				:class="{ 'is-selected': index === selectedIndex }"
				@click="selectItem(index)"
			>
				<div class="filter-autocomplete__content">
					<XLabel
						v-if="item.fieldType === 'labels'"
						:label="(item.item as unknown as ILabel)"
						class="filter-autocomplete__label"
					/>
					<User
						v-else-if="item.fieldType === 'assignees'"
						:user="(item.item as unknown as IUser)"
						:avatar-size="20"
						class="filter-autocomplete__user"
					/>
					<div
						v-else
						class="filter-autocomplete__project"
					>
						{{ item.title }}
					</div>
				</div>
			</button>
		</template>
		<div
			v-else
			class="filter-autocomplete no-results"
		>
			{{ $t('filters.noResults') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'
import { ref, watch } from 'vue'
import type { ILabel } from '@/modelTypes/ILabel'
import type { IUser } from '@/modelTypes/IUser'
import type { AutocompleteItem } from './FilterAutocomplete'

interface Props {
	items: AutocompleteItem[]
	command: (item: AutocompleteItem) => void
}

const props = defineProps<Props>()

const selectedIndex = ref(0)

watch(
	() => props.items,
	() => {
		selectedIndex.value = 0
	},
)

function onKeyDown({event}: { event: KeyboardEvent }) {
	if (event.key === 'ArrowUp') {
		event.preventDefault()
		event.stopPropagation()
		upHandler()
		return true
	}

	if (event.key === 'ArrowDown') {
		event.preventDefault()
		event.stopPropagation()
		downHandler()
		return true
	}

	if (event.key === 'Enter') {
		event.preventDefault()
		event.stopPropagation()
		enterHandler()
		return true
	}

	return false
}

function upHandler() {
	selectedIndex.value = ((selectedIndex.value + props.items.length) - 1) % props.items.length
}

function downHandler() {
	selectedIndex.value = (selectedIndex.value + 1) % props.items.length
}

function enterHandler() {
	selectItem(selectedIndex.value)
}

function selectItem(index: number) {
	const item = props.items[index]
	if (item) {
		props.command(item)
	}
}

defineExpose({
	onKeyDown,
})
</script>

<style lang="scss" scoped>
.filter-autocompletes {
	position: relative;
	border-radius: $radius;
	background: var(--white);
	color: var(--grey-900);
	overflow: hidden;
	font-size: 0.875rem;
	box-shadow: var(--shadow-md);
	border: 1px solid var(--grey-200);
	max-block-size: 12rem;
	overflow-y: auto;
}

.filter-autocomplete {
	display: flex;
	align-items: center;
	margin: 0;
	inline-size: 100%;
	text-align: start;
	background: transparent;
	border-radius: $radius;
	border: 0;
	padding: 0.375rem 0.5rem;
	transition: background-color var(--transition);
	cursor: pointer;

	&.is-selected,
	&:hover {
		background: var(--grey-100);
	}

	&.no-results {
		color: var(--grey-500);
		cursor: default;
		
		&:hover {
			background: transparent;
		}
	}
}

.filter-autocomplete__content {
	display: flex;
	align-items: center;
	inline-size: 100%;
}

.filter-autocomplete__label {
	font-size: 0.75rem;
}

.filter-autocomplete__user {
	font-size: 0.875rem;
}

.filter-autocomplete__project {
	color: var(--grey-800);
	font-weight: 500;
}
</style>
