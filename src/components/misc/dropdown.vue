<template>
	<div class="dropdown is-right is-active" ref="dropdown">
		<div class="dropdown-trigger" @click="open = !open">
			<slot name="trigger">
				<icon :icon="triggerIcon" class="icon"/>
			</slot>
		</div>
		<transition name="fade">
			<div class="dropdown-menu" v-if="open">
				<div class="dropdown-content">
					<slot></slot>
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
		document.addEventListener('click', this.hide)
	},
	beforeDestroy() {
		document.removeEventListener('click', this.hide)
	},
	props: {
		triggerIcon: {
			type: String,
			default: 'ellipsis-h',
		},
	},
	methods: {
		hide(e) {
			if (this.open) {
				closeWhenClickedOutside(e, this.$refs.dropdown, () => {
					this.open = false
					this.$emit('close', e)
				})
			}
		},
	},
}
</script>
