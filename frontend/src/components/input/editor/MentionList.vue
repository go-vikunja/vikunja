<template>
	<div class="mention-items">
		<template v-if="items.length">
			<button
				v-for="(item, index) in items"
				:key="item.id"
				class="mention-item"
				:class="{ 'is-selected': index === selectedIndex }"
				@click="selectItem(index)"
			>
				<img
					:src="item.avatarUrl"
					alt=""
					class="mention-avatar"
				>
				<div class="mention-info">
					<p class="mention-name">
						{{ item.label }}
					</p>
					<p class="mention-username">
						@{{ item.username }}
					</p>
				</div>
			</button>
		</template>
		<div
			v-else
			class="mention-item no-results"
		>
			No users found
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
		onKeyDown({event}: {event: KeyboardEvent}) {
			if (event.key === 'ArrowUp') {
				this.upHandler()
				return true
			}

			if (event.key === 'ArrowDown') {
				this.downHandler()
				return true
			}

			if (event.key === 'Enter') {
				if (event.isComposing) {
					return false
				}
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

		selectItem(index: number) {
			const item = this.items[index]

			if (item) {
				this.command(item)
			}
		},
	},
}
</script>

<style lang="scss" scoped>
.mention-items {
	padding: 0.2rem;
	position: relative;
	border-radius: 0.5rem;
	background: var(--white);
	color: var(--grey-900);
	overflow: hidden;
	font-size: 0.9rem;
	box-shadow: var(--shadow-md);
	min-inline-size: 200px;
	max-block-size: 300px;
	overflow-y: auto;
}

.mention-item {
	display: flex;
	align-items: center;
	margin: 0;
	inline-size: 100%;
	text-align: start;
	background: transparent;
	border-radius: $radius;
	border: 0;
	padding: 0.4rem 0.6rem;
	transition: background-color $transition;

	&.is-selected, &:hover {
		background: var(--grey-100);
		cursor: pointer;
	}
	
	&.no-results {
		color: var(--grey-500);
		cursor: default;
	}
}

.mention-avatar {
	inline-size: 32px;
	block-size: 32px;
	border-radius: 50%;
	margin-inline-end: 0.75rem;
	flex-shrink: 0;
}

.mention-info {
	display: flex;
	flex-direction: column;
	min-inline-size: 0;
	flex: 1;
	
	p {
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.mention-name {
	font-size: 0.9rem;
	color: var(--grey-800);
	font-weight: 500;
}

.mention-username {
	font-size: 0.75rem;
	color: var(--grey-500);
}
</style>
