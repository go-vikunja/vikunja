<template>
	<div class="password-field">
		<input
			id="password"
			class="input"
			name="password"
			:placeholder="$t('user.auth.passwordPlaceholder')"
			required
			:type="passwordFieldType"
			autocomplete="current-password"
			:tabindex="tabindex"
			@keyup.enter="e => $emit('submit', e)"
			@focusout="() => {validate(); validateAfterFirst = true}"
			@keyup="() => {validateAfterFirst ? validate() : null}"
			@input="handleInput"
		>
		<BaseButton
			v-tooltip="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
			class="password-field-type-toggle"
			:aria-label="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
			@click="togglePasswordFieldType"
		>
			<Icon :icon="passwordFieldType === 'password' ? 'eye' : 'eye-slash'" />
		</BaseButton>
	</div>
	<p
		v-if="isValid !== true"
		class="help is-danger"
	>
		{{ isValid }}
	</p>
</template>

<script lang="ts" setup>
import {ref, watchEffect} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import BaseButton from '@/components/base/BaseButton.vue'
import {validatePassword} from '@/helpers/validatePasswort'

const props = withDefaults(defineProps<{
	modelValue: string,
	tabindex?: string,
	// This prop is a workaround to trigger validation from the outside when the user never had focus in the input.
	validateInitially?: boolean,
	validateMinLength?: boolean,
}>(), {
	tabindex: undefined,
	validateMinLength: true,
})

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'submit': [event: Event],
}>()
const {t} = useI18n()
const passwordFieldType = ref('password')
const password = ref('')
// eslint-disable-next-line vue/no-setup-props-reactivity-loss
const isValid = ref<true | string>(props.validateInitially === true ? true : '')
const validateAfterFirst = ref(false)

watchEffect(() => props.validateInitially && validate())

const validate = useDebounceFn(() => {
	const valid = validatePassword(password.value, props.validateMinLength)
	isValid.value = valid === true ? true : t(valid)
}, 100)

function togglePasswordFieldType() {
	passwordFieldType.value = passwordFieldType.value === 'password'
		? 'text'
		: 'password'
}

function handleInput(e: Event) {
	password.value = (e.target as HTMLInputElement)?.value
	emit('update:modelValue', password.value)
}
</script>

<style scoped>
.password-field {
	position: relative;
}

.password-field-type-toggle {
	position: absolute;
	color: var(--grey-400);
	inset-block-start: 50%;
	inset-inline-end: 1rem;
	transform: translateY(-50%);
}
</style>
