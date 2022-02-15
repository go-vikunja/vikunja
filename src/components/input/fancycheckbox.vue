<template>
	<div :class="{'is-disabled': disabled}" class="fancycheckbox">
		<input
			:checked="checked"
			:disabled="disabled || null"
			:id="checkBoxId"
			@change="(event) => updateData(event.target.checked)"
			type="checkbox"/>
		<label :for="checkBoxId" class="check">
			<svg height="18px" viewBox="0 0 18 18" width="18px">
				<path
					d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
				<polyline points="1 9 7 14 15 4"></polyline>
			</svg>
			<span>
				<slot></slot>
			</span>
		</label>
	</div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

import {createRandomID} from '@/helpers/randomId'

export default defineComponent({
	name: 'fancycheckbox',
	data() {
		return {
			checked: false,
			checkBoxId: `fancycheckbox_${createRandomID()}`,
		}
	},
	props: {
		modelValue: {
			required: false,
		},
		disabled: {
			type: Boolean,
			required: false,
			default: false,
		},
	},
	emits: ['update:modelValue', 'change'],
	watch: {
		modelValue: {
			handler(modelValue) {
				this.checked = modelValue

			},
			immediate: true,
		},
	},
	methods: {
		updateData(checked) {
			this.checked = checked
			this.$emit('update:modelValue', checked)
			this.$emit('change', checked)
		},
	},
})
</script>


<style lang="scss" scoped>
.fancycheckbox {
  display: inline-block;
  padding-right: 5px;
  padding-top: 3px;

	// FIXME: should be a prop
	&.is-block {
		margin: .5rem .2rem;
	}
}

input[type=checkbox] {
	display: none;
}

.check {
	cursor: pointer;
	position: relative;
	margin: auto;
	width: 18px;
	height: 18px;
	-webkit-tap-highlight-color: transparent;
	transform: translate3d(0, 0, 0);
}

span {
	font-size: 0.8rem;
	vertical-align: top;
	padding-left: .5rem;
}

svg {
	position: relative;
	z-index: 1;
	fill: none;
	stroke-linecap: round;
	stroke-linejoin: round;
	stroke: #c8ccd4;
	stroke-width: 1.5;
	transform: translate3d(0, 0, 0);
	transition: all 0.2s ease;
}

.check:hover svg {
	stroke: var(--primary);
}

.is-disabled .check:hover svg {
	stroke: #c8ccd4;
}

path {
	stroke-dasharray: 60;
	stroke-dashoffset: 0;
}

polyline {
	stroke-dasharray: 22;
	stroke-dashoffset: 66;
}

input[type=checkbox]:checked + .check {
	svg {
		stroke: var(--primary);
	}

	path {
		stroke-dashoffset: 60;
		transition: all 0.3s linear;
	}

	polyline {
		stroke-dashoffset: 42;
		transition: all 0.2s linear;
		transition-delay: 0.15s;
	}
}
</style>