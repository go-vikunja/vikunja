<script setup lang="ts">
import FormField from '@/components/input/FormField.vue'
import Password from '@/components/input/Password.vue'

withDefaults(defineProps<{
	passwordId: string,
	passcodeId?: string,
	passwordLabel?: string,
	showPasscode?: boolean,
	focus?: boolean,
	autofocus?: boolean,
}>(), {
	passcodeId: undefined,
	passwordLabel: undefined,
	showPasscode: false,
	focus: false,
	autofocus: false,
})

defineEmits<{submit: []}>()
const password = defineModel<string>('password', {required: true})
const passcode = defineModel<string>('passcode', {default: ''})
</script>

<template>
	<FormField
		:id="passwordId"
		v-focus="focus"
		:label="passwordLabel ?? $t('user.settings.currentPassword')"
	>
		<Password
			:id="passwordId"
			v-model="password"
			:autofocus="autofocus"
			:placeholder="$t('user.settings.currentPasswordPlaceholder')"
			:validate-min-length="false"
			@submit="$emit('submit')"
		/>
	</FormField>
	<FormField
		v-if="showPasscode"
		:id="passcodeId"
		v-model="passcode"
		:label="$t('user.settings.totp.passcode')"
		:placeholder="$t('user.settings.totp.passcodePlaceholder')"
		autocomplete="one-time-code"
		type="text"
		inputmode="numeric"
		@keyup.enter="$emit('submit')"
	/>
</template>
