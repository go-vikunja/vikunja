<template>
	<div class="password-field">
		<input
			:id="id"
			class="input"
			name="password"
			:placeholder="placeholder ?? $t('user.auth.passwordPlaceholder')"
			:value="password"
			:autofocus="autofocus || undefined"
			required
			:type="passwordFieldType"
			:autocomplete="autocomplete"
			:aria-invalid="isValid !== true ? true : undefined"
			:aria-describedby="errorId"
			@keyup.enter="handleSubmit"
			@focusout="e => {handleInput(e); validate(); validateAfterFirst = true}"
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
		:id="errorId"
		class="help is-danger"
		role="alert"
	>
		{{ isValid }}
	</p>
</template>

<script lang="ts" setup>
import {computed, ref, watchEffect} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import BaseButton from '@/components/base/BaseButton.vue'
import {validatePassword} from '@/helpers/validatePasswort'

const props = withDefaults(defineProps<{
	id?: string,
	placeholder?: string,
	autofocus?: boolean,
	// This prop is a workaround to trigger validation from the outside when the user never had focus in the input.
	validateInitially?: boolean,
	validateMinLength?: boolean,
	autocomplete?: string,
}>(), {
	id: 'password',
	placeholder: undefined,
	autofocus: false,
	validateMinLength: true,
	autocomplete: 'current-password',
})

const emit = defineEmits<{
	'submit': [event: Event],
}>()
const password = defineModel<string>({required: true})
const {t} = useI18n()
const passwordFieldType = ref('password')
const isValid = ref<true | string>(true)
const validateAfterFirst = ref(false)
const errorId = computed(() => isValid.value !== true ? `${props.id}-error` : undefined)

function validateValue(value: string) {
	const valid = validatePassword(value, props.validateMinLength)
	isValid.value = valid === true ? true : t(valid)
}

const validate = useDebounceFn(() => validateValue(password.value), 100)

watchEffect(() => {
	if (props.validateInitially) {
		validateAfterFirst.value = true
		validate()
	}
})

function togglePasswordFieldType() {
	passwordFieldType.value = passwordFieldType.value === 'password'
		? 'text'
		: 'password'
}

function handleInput(e: Event) {
	password.value = (e.target as HTMLInputElement).value
}

function handleSubmit(e: Event) {
	// Autofill may not fire an input event.
	handleInput(e)
	validateValue((e.target as HTMLInputElement).value)
	validateAfterFirst.value = true
	emit('submit', e)
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
