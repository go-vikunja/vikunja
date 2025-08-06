<template>
	<div class="items">
		<template v-if="items.length">
			<button
				v-for="(item, index) in items"
				:key="index"
				class="item"
				:class="{ 'is-selected': index === selectedIndex }"
				@click="selectItem(index)"
			>
				<Icon :icon="item.icon" />
				<div class="description">
					<p>{{ item.title }}</p>
					<p>{{ item.description }}</p>
				</div>
			</button>
		</template>
		<div
			v-else
			class="item"
		>
			No result
		</div>
	</div>
</template>

<script lang="ts">
/* eslint-disable vue/component-api-style */
export default {
	props: {
		items: {
			type: Array,
			required: true,
		},

		command: {
			type: Function,
			required: true,
		},
	},

	data() {
		return {
			selectedIndex: 0,
		}
	},

	watch: {
		items() {
			this.selectedIndex = 0
		},
	},

	methods: {
		onKeyDown({event}) {
			if (event.key === 'ArrowUp') {
				this.upHandler()
				return true
			}

			if (event.key === 'ArrowDown') {
				this.downHandler()
				return true
			}

			if (event.key === 'Enter') {
				this.enterHandler()
				return true
			}

			return false
		},

		upHandler() {
			this.selectedIndex = ((this.selectedIndex + this.items.length) - 1) % this.items.length
		},

		downHandler() {
			this.selectedIndex = (this.selectedIndex + 1) % this.items.length
		},

		enterHandler() {
			this.selectItem(this.selectedIndex)
		},

		selectItem(index) {
			const item = this.items[index]

			if (item) {
				this.command(item)
			}
		},
	},
}
</script>

<style lang="scss" scoped>
.items {
	padding: 0.2rem;
	position: relative;
	border-radius: 0.5rem;
	background: var(--white);
	color: var(--grey-900);
	overflow: hidden;
	font-size: 0.9rem;
	box-shadow: var(--shadow-md);
}

.item {
	display: flex;
	align-items: center;
	margin: 0;
	inline-size: 100%;
	text-align: start;
	background: transparent;
	border-radius: $radius;
	border: 0;
	padding: 0.2rem 0.4rem;
	transition: background-color $transition;

	&.is-selected, &:hover {
		background: var(--grey-100);
		cursor: pointer;
	}
	
	> svg {
		box-sizing: border-box;
		inline-size: 2rem;
		block-size: 2rem;
		border: 1px solid var(--grey-300);
		padding: .5rem;
		margin-inline-end: .5rem;
		border-radius: $radius;
		color: var(--grey-700);
	}
}

.description {
	display: flex;
	flex-direction: column;
	font-size: .9rem;
	color: var(--grey-800);
	
	p:last-child {
		font-size: .75rem;
		color: var(--grey-500);
	}
}
</style>
