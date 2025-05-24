<script lang="ts" setup>
import {ref} from 'vue'
import {logEvent} from 'histoire/client'
import FancyCheckbox from './FancyCheckbox.vue'

const isDisabled = ref<boolean | undefined>()

const isChecked = ref(false)

const isCheckedInitiallyEnabled = ref(true)

const isCheckedDisabled = ref(false)

const withoutInitialState = ref<boolean | undefined>()
</script>


<template>
	<Story :layout="{ type: 'grid', width: '200px' }">
		<Variant title="Default">
			<FancyCheckbox
				v-model="isChecked"
				:disabled="isDisabled"
			>
				This is probably not important
			</FancyCheckbox>

			Visualisation
			<input
				v-model="isChecked"
				type="checkbox"
			>
			{{ isChecked }}
		</Variant>
		<Variant title="Enabled Initially">
			<FancyCheckbox
				v-model="isCheckedInitiallyEnabled"
				:disabled="isDisabled"
			>
				We want you to use this option
			</FancyCheckbox>

			Visualisation
			<input
				v-model="isCheckedInitiallyEnabled"
				type="checkbox"
			>
			{{ isCheckedInitiallyEnabled }}
		</Variant>
		<Variant title="Disabled">
			<FancyCheckbox
				disabled
				:model-value="isCheckedDisabled"
				@update:modelValue="logEvent('Setting disabled: This should never happen', $event)"
			>
				You can't change this
			</FancyCheckbox>

			Visualisation
			<input
				v-model="isCheckedDisabled"
				type="checkbox"
				disabled
			>
			{{ isCheckedDisabled }}
		</Variant>

		<Variant title="Undefined initial State">
			<FancyCheckbox
				v-model="withoutInitialState"
				:disabled="isDisabled"
			>
				Not sure what the value should be
			</FancyCheckbox>

			Visualisation
			<input
				v-model="withoutInitialState"
				type="checkbox"
				disabled
			>
			{{ withoutInitialState }}
		</Variant>
	</Story>
</template>
