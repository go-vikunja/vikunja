<template>
	<div class="color-picker-container">
		<verte
				v-model="color"
				:menuPosition="menuPosition"
				picker="square"
				model="hex"
				:enableAlpha="false"
				:rgbSliders="true"/>
		<a @click="reset" class="reset">
			Reset Color
		</a>
	</div>
</template>

<script>
	import verte from 'verte'
	import 'verte/dist/verte.css'

	export default {
		name: 'colorPicker',
		data() {
			return {
				color: '',
				lastChangeTimeout: null,
			}
		},
		components: {
			verte,
		},
		props: {
			value: {
				required: true,
			},
			menuPosition: {
				type: String,
				default: 'top',
			},
		},
		watch: {
			value(newVal) {
				this.color = newVal
			},
			color() {
				this.update()
			},
		},
		mounted() {
			this.color = this.value
		},
		methods: {
			update() {

				if(this.lastChangeTimeout !== null) {
					clearTimeout(this.lastChangeTimeout)
				}

				this.lastChangeTimeout = setTimeout(() => {
					this.$emit('input', this.color)
					this.$emit('change')
				}, 500)
			},
			reset() {
				// FIXME: I havn't found a way to make it clear to the user the color war reset.
				//  Not sure if verte is capable of this - it does not show the change when setting this.color = ''
				this.color = ''
				this.update()
			},
		},
	}
</script>
