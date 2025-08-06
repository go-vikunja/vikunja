<template>
	<div class="color-picker-container">
		<datalist :id="colorListID">
			<option
				v-for="defaultColor in defaultColors"
				:key="defaultColor"
				:value="defaultColor"
			/>
		</datalist>
		
		<div class="picker">
			<input
				v-model="color"
				class="picker__input"
				type="color"
				:list="colorListID"
				:class="{'is-empty': isEmpty}"
			>
			<svg
				v-show="isEmpty"
				class="picker__pattern"
				viewBox="0 0 22 22"
				fill="fff"
			>
				<pattern
					id="checker"
					width="11"
					height="11"
					patternUnits="userSpaceOnUse"
					fill="FFF"
				>
					<rect
						fill="#cccccc"
						x="0"
						width="5.5"
						height="5.5"
						y="0"
					/>
					<rect
						fill="#cccccc"
						x="5.5"
						width="5.5"
						height="5.5"
						y="5.5"
					/>
				</pattern>
				<rect
					width="22"
					height="22"
					fill="url(#checker)"
				/>
			</svg>
		</div>

		<XButton
			v-if="!isEmpty"
			:disabled="isEmpty"
			class="is-small mis-2"
			:shadow="false"
			variant="secondary"
			@click="reset"
		>
			{{ $t('input.resetColor') }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {createRandomID} from '@/helpers/randomId'
import XButton from '@/components/input/Button.vue'

const model = defineModel<string>({ required: true })

const DEFAULT_COLORS = [
	'#1973ff',
	'#7F23FF',
	'#ff4136',
	'#ff851b',
	'#ffeb10',
	'#00db60',
]

const color = ref('')
const lastChangeTimeout = ref<ReturnType<typeof setTimeout> | null>(null)
const defaultColors = ref(DEFAULT_COLORS)
const colorListID = ref(createRandomID())

watch(
	model,
	(newValue) => {
		if (newValue === '' || newValue.startsWith('var(')) {
			color.value = ''
			return
		}

		if (!newValue.startsWith('#') && (newValue.length === 6 || newValue.length === 3)) {
			newValue = `#${newValue}`
		}

		color.value = newValue
	},
	{immediate: true},
)

watch(color, () => update())

const isEmpty = computed(() => color.value === '')

function update(force = false) {
	if(isEmpty.value && !force) {
		return
	}

	if (lastChangeTimeout.value !== null) {
		clearTimeout(lastChangeTimeout.value)
	}

	lastChangeTimeout.value = setTimeout(() => {
		model.value = color.value
	}, 500)
}

function reset() {
	// FIXME: I havn't found a way to make it clear to the user the color war reset.
	//  Not sure if verte is capable of this - it does not show the change when setting this.color = ''
	color.value = ''
	update(true)
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

	$picker-size: 24px;
	$border-width: 1px;
	.picker {
		display: grid;
		inline-size: $picker-size;
		block-size: $picker-size;
		overflow: hidden;
		border-radius: 100%;
		border: $border-width solid var(--grey-300);
		box-shadow: $shadow;

		& > * {
			grid-row: 1;
			grid-column: 1;
		}
	}

	input.picker__input {
		padding: 0;
		inline-size: $picker-size - 2 * $border-width;
		block-size: $picker-size - 2 * $border-width;
	}

	.picker__input.is-empty {
		opacity: 0;
	}
	
	.picker__pattern {
		pointer-events: none;
	}
}
</style>
