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
			:tabindex="props.tabindex"
			@keyup.enter="e => $emit('submit', e)"
			@focusout="validate"
			@input="handleInput"
		>
		<BaseButton
			v-tooltip="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
			class="password-field-type-toggle"
			:aria-label="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
			@click="togglePasswordFieldType"
		>
			<icon :icon="passwordFieldType === 'password' ? 'eye' : 'eye-slash'" />
		</BaseButton>
	</div>
	<p
		v-if="!isValid"
		class="help is-danger"
	>
		{{ $t('user.auth.passwordRequired') }}
	</p>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps({
	tabindex: String,
	modelValue: String,
	// This prop is a workaround to trigger validation from the outside when the user never had focus in the input.
	validateInitially: Boolean,
})

const emit = defineEmits(['submit', 'update:modelValue'])

const passwordFieldType = ref('password')
const password = ref('')
const isValid = ref(!props.validateInitially)

watch(
	() => props.validateInitially,
	() => props.validateInitially && validate(),
	{immediate: true},
)

const validate = useDebounceFn(() => {
	isValid.value = password.value !== ''
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
	top: 50%;
	right: 1rem;
	transform: translateY(-50%);
}
</style>