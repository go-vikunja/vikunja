<script lang="ts" setup>
import {ref} from 'vue'
import {logEvent} from 'histoire/client'
import FancyCheckbox from './fancycheckbox.vue'

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
			<input type="checkbox" v-model="isChecked">
			{{ isChecked }}
		</Variant>
		<Variant title="Enabled Initially">
			<FancyCheckbox
				:disabled="isDisabled"
				v-model="isCheckedInitiallyEnabled"
			>
				We want you to use this option
			</FancyCheckbox>

			Visualisation
			<input type="checkbox" v-model="isCheckedInitiallyEnabled">
			{{ isCheckedInitiallyEnabled }}
		</Variant>
		<Variant title="Disabled">
			<FancyCheckbox
				disabled
				:modelValue="isCheckedDisabled"
				@update:model-value="logEvent('Setting disabled: This should never happen', $event)"
			>
				You can't change this
			</FancyCheckbox>

			Visualisation
			<input type="checkbox" v-model="isCheckedDisabled" disabled>
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
			<input type="checkbox" v-model="withoutInitialState" disabled>
			{{ withoutInitialState }}
		</Variant>
	</Story>
</template>