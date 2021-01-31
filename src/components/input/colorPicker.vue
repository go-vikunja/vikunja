<template>
	<div class="color-picker-container">
		<verte
			:showHistory="true"
			:colorHistory="[
				'#1973ff',
				'#7F23FF',
				'#ff4136',
				'#ff851b',
				'#ffeb10',
				'#00db60',
			]"
			:enableAlpha="false"
			:menuPosition="menuPosition"
			:rgbSliders="true"
			model="hex"
			picker="square"
			v-model="color"
			:class="{'is-empty': empty}"
		/>
		<x-button @click="reset" class="is-small ml-2" :shadow="false" type="secondary">
			Reset Color
		</x-button>
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
	computed: {
		empty() {
			return this.color === '#000000' || this.color === ''
		},
	},
	methods: {
		update(force = false) {

			if(this.empty && !force) {
				return
			}

			if (this.lastChangeTimeout !== null) {
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
			this.update(true)
		},
	},
}
</script>

<style lang="scss">
.verte.is-empty {
	.verte__icon {
		opacity: 0;
	}

	.verte__guide {
		background: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAGklEQVQYlWM4c+bMf3TMgA0MBYWDzDkUKQQAlHCpV9ycHeMAAAAASUVORK5CYII=);
	}
}
</style>
