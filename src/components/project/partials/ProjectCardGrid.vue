<template>
    <ul class="list-grid">
			<li
				v-for="(item, index) in filteredLists"
				:key="`list_${item.id}_${index}`"
				class="list-grid-item"
			>
				<ListCard :list="item" />
			</li>
    </ul>
</template>

<script lang="ts" setup>
import {computed, type PropType} from 'vue'
import type {IList} from '@/modelTypes/IList'

import ListCard from './ListCard.vue'

const props = defineProps({
	lists: {
		type: Array as PropType<IList[]>,
		default: () => [],
	},
	showArchived: {
		default: false,
		type: Boolean,
	},
	itemLimit: {
		type: Boolean,
		default: false,
	},
})

const filteredLists = computed(() => {
	return props.showArchived
		? props.lists
		: props.lists.filter(l => !l.isArchived)
})
</script>

<style lang="scss" scoped>
$list-height: 150px;
$list-spacing: 1rem;

.list-grid {
	margin: 0; // reset li
	list-style-type: none;
	display: grid;
	grid-template-columns: repeat(var(--list-columns), 1fr);
	grid-auto-rows: $list-height;
	gap: $list-spacing;

	@media screen and (min-width: $mobile) {
		--list-rows: 4;
		--list-columns: 1;
	}

	@media screen and (min-width: $mobile) and (max-width: $tablet) {
		--list-columns: 2;
	}

	@media screen and (min-width: $tablet) and (max-width: $widescreen) {
		--list-columns: 3;
		--list-rows: 3;
	}

	@media screen and (min-width: $widescreen) {
		--list-columns: 5;
		--list-rows: 2;
	}
}

.list-grid-item {
	display: grid;
	margin-top: 0; // remove padding coming form .content li + li
}
</style>