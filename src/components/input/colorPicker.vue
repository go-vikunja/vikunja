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

		<x-button
			v-if="!isEmpty"
			:disabled="isEmpty"
			@click="reset"
			class="is-small ml-2"
			:shadow="false"
			type="secondary"
		>
			{{ $t('input.resetColor') }}
		</x-button>
	</div>
</template>

<script>
import {createRandomID} from '@/helpers/randomId'

const DEFAULT_COLORS = [
	'#1973ff',
	'#7F23FF',
	'#ff4136',
	'#ff851b',
	'#ffeb10',
	'#00db60',
]

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
		modelValue: {
			required: true,
		},
		menuPosition: {
			type: String,
			default: 'top',
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		modelValue: {
			handler(modelValue) {
				this.color = modelValue
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
				this.$emit('update:modelValue', this.color)
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

<style lang="scss" scoped>
.color-picker-container {
  display: flex;
  justify-content: center;
  align-items: center;

	// reset / see https://stackoverflow.com/a/11471224/15522256
	input[type="color"] {
		-webkit-appearance: none;
		border: none;
	}
	input[type="color"]::-webkit-color-swatch-wrapper {
		padding: 0;
	}
	input[type="color"]::-webkit-color-swatch {
		border: none;
	}

	$PICKER_SIZE: 24px;
	$BORDER_WIDTH: 1px;
	.picker {
		display: grid;
		width: $PICKER_SIZE;
		height: $PICKER_SIZE;
		overflow: hidden;
		border-radius: 100%;
		border: $BORDER_WIDTH solid $grey-300;
		box-shadow: $shadow;

		& > * {
			grid-row: 1;
			grid-column: 1;
		}
	}

	input.picker__input {
		padding: 0;
		width: $PICKER_SIZE - 2 * $BORDER_WIDTH;
		height: $PICKER_SIZE - 2 * $BORDER_WIDTH;
	}

	.picker__input.is-empty {
		opacity: 0;
	}
	
	.picker__pattern {
		pointer-events: none;
	}
}
</style>
