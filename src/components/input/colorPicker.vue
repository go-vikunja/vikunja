<template>
	<div class="color-picker-container">
		<datalist :id="colorListID">
			<option v-for="color in defaultColors" :key="color" :value="color" />
		</datalist>
		
		<div class="picker">
			<input
				class="picker__input"
				type="color"
				v-model="color"
				:list="colorListID"
				:class="{'is-empty': isEmpty}"
			/>
			<svg class="picker__pattern" v-show="isEmpty" viewBox="0 0 22 22" fill="fff">
				<pattern id="checker" width="11" height="11" patternUnits="userSpaceOnUse" fill="FFF">
					<rect fill="#cccccc" x="0" width="5.5" height="5.5" y="0"></rect>
					<rect fill="#cccccc" x="5.5" width="5.5" height="5.5" y="5.5"></rect>
				</pattern>
				<rect width="22" height="22" fill="url(#checker)"></rect>
			</svg>
		</div>

		<x-button :disabled="isEmpty" @click="reset" class="is-small ml-2" :shadow="false" type="secondary">
			{{ $t('input.resetColor') }}
		</x-button>
	</div>
</template>

<script>
const DEFAULT_COLORS = [
	'#1973ff',
	'#7F23FF',
	'#ff4136',
	'#ff851b',
	'#ffeb10',
	'#00db60',
]

function createRandomID() {
	const ID_LENGTH = 9
	return Math.random().toString(36).substr(2, ID_LENGTH)
}

export default {
	name: 'colorPicker',
	data() {
		return {
			color: '',
			lastChangeTimeout: null,
			defaultColors: DEFAULT_COLORS,
			colorListID: createRandomID(),
		}
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
		value: {
			handler(value) {
				this.color = value
			},
			immediate: true,
		},
		color() {
			this.update()
		},
	},
	computed: {
		isEmpty() {
			return this.color === '#000000' || this.color === ''
		},
	},
	methods: {
		update(force = false) {

			if(this.isEmpty && !force) {
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