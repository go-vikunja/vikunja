<template>
	<div class="filter-commands">
		<template v-if="items.length">
			<button
				v-for="(item, index) in items"
				:key="`${item.fieldType}-${item.id}`"
				class="filter-command"
				:class="{ 'is-selected': index === selectedIndex }"
				@click="selectItem(index)"
			>
				<div class="filter-command__content">
					<XLabel
						v-if="item.fieldType === 'labels'"
						:label="item.item"
						class="filter-command__label"
					/>
					<User
						v-else-if="item.fieldType === 'assignees'"
						:user="item.item"
						:avatar-size="20"
						class="filter-command__user"
					/>
					<div
						v-else
						class="filter-command__project"
					>
						{{ item.title }}
					</div>
				</div>
			</button>
		</template>
		<div
			v-else
			class="filter-command no-results"
		>
			No results
		</div>
	</div>
</template>

<script lang="ts">
/* eslint-disable vue/component-api-style */
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'

export default {
	components: {
		XLabel,
		User,
	},
	
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
.filter-commands {
	padding: 0.25rem;
	position: relative;
	border-radius: $radius;
	background: var(--white);
	color: var(--grey-900);
	overflow: hidden;
	font-size: 0.875rem;
	box-shadow: var(--shadow-md);
	border: 1px solid var(--grey-200);
	max-height: 12rem;
	overflow-y: auto;
}

.filter-command {
	display: flex;
	align-items: center;
	margin: 0;
	width: 100%;
	text-align: left;
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

.filter-command__content {
	display: flex;
	align-items: center;
	width: 100%;
}

.filter-command__label {
	font-size: 0.75rem;
}

.filter-command__user {
	font-size: 0.875rem;
}

.filter-command__project {
	color: var(--grey-800);
	font-weight: 500;
}
</style>