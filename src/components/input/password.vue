<template>
	<div class="password-field">
		<input
			class="input"
			id="password"
			name="password"
			:placeholder="$t('user.auth.passwordPlaceholder')"
			required
			:type="passwordFieldType"
			autocomplete="current-password"
			@keyup.enter="e => $emit('submit', e)"
			:tabindex="props.tabindex"
			@focusout="validate"
			@input="handleInput"
		/>
		<a
			@click="togglePasswordFieldType"
			class="password-field-type-toggle"
			:aria-label="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')"
			v-tooltip="passwordFieldType === 'password' ? $t('user.auth.showPassword') : $t('user.auth.hidePassword')">
			<icon :icon="passwordFieldType === 'password' ? 'eye' : 'eye-slash'"/>
		</a>
	</div>
	<p class="help is-danger" v-if="!isValid">
		{{ $t('user.auth.passwordRequired') }}
	</p>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'

const props = defineProps({
	tabindex: String,
	modelValue: String,
	// This prop is a workaround to trigger validation from the outside when the user never had focus in the input.
	validateInitially: Boolean,
})

const emit = defineEmits(['submit', 'update:modelValue'])

const passwordFieldType = ref<String>('password')
const password = ref<String>('')
const isValid = ref<Boolean>(!props.validateInitially)

watch(
	() => props.validateInitially,
	(doValidate: Boolean) => {
		if (doValidate) {
			validate()
		}
	},
)

function validate() {
	useDebounceFn(() => {
		isValid.value = password.value !== ''
	}, 100)()
}

function togglePasswordFieldType() {
	passwordFieldType.value = passwordFieldType.value === 'password'
		? 'text'
		: 'password'
}

function handleInput(e) {
	password.value = e.target.value
	emit('update:modelValue', e.target.value)
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