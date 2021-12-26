<template>
	<div class="dropdown is-right is-active" ref="dropdown">
		<div class="dropdown-trigger is-flex" @click="open = !open">
			<slot name="trigger" :close="close">
				<icon :icon="triggerIcon" class="icon"/>
			</slot>
		</div>
		<transition name="fade">
			<div class="dropdown-menu" v-if="open">
				<div class="dropdown-content">
					<slot :close="close"></slot>
				</div>
			</div>
		</transition>
	</div>
</template>

<script>
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'

export default {
	name: 'dropdown',
	data() {
		return {
			open: false,
		}
	},
	mounted() {
		document.addEventListener('click', this.handleClickOutside)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.handleClickOutside)
	},
	props: {
		triggerIcon: {
			type: String,
			default: 'ellipsis-h',
		},
	},
	emits: ['close'],
	methods: {
		close() {
			this.open = false
		},
		toggleOpen() {
			this.open = !this.open
		},
		handleClickOutside(e) {
			if (!this.open) {
				return
			}
			closeWhenClickedOutside(e, this.$refs.dropdown, () => {
				this.open = false
				this.$emit('close', e)
			})
		},
	},
}
</script>
