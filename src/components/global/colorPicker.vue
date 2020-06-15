<template>
	<div class="color-picker-container">
		<verte
				v-model="color"
				menuPosition="top"
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
			}
		},
		components: {
			verte,
		},
		props: {
			value: {
				required: true,
			},
		},
		watch: {
			value(newVal) {
				this.color = newVal
			},
			color() {
				this.update()
			}
		},
		mounted() {
			this.color = this.value
		},
		methods: {
			update() {
				this.$emit('input', this.color)
				this.$emit('change')
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
