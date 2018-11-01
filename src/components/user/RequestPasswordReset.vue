<template>
	<div>
		<h2 class="title">Reset your password</h2>
		<div class="box">
			<form id="loginform" @submit.prevent="submit" v-if="!isSuccess">
				<div class="field">
					<div class="control">
						<input type="text" class="input" name="username" placeholder="Username" v-model="username" required>
					</div>
				</div>

				<div class="field is-grouped">
					<div class="control">
						<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Send me a password reset link</button>
						<router-link :to="{ name: 'login' }" class="button">Login</router-link>
					</div>
				</div>
				<div class="notification is-danger" v-if="error">
					{{ error }}
				</div>
			</form>
			<div v-if="isSuccess" class="has-text-centered">
				<div class="notification is-success">
					Check your inbox! You should have a mail with instructions on how to reset your password.
				</div>
				<router-link :to="{ name: 'login' }" class="button is-primary">Login</router-link>
			</div>
		</div>
	</div>
</template>

<script>
    import {HTTP} from '../../http-common'

    export default {
        data() {
            return {
                username: '',
                error: '',
                isSuccess: false,
                loading: false
            }
        },
        methods: {
            submit() {
                this.loading = true
                this.error = ''
                let credentials = {
                    user_name: this.username,
                }

                HTTP.post(`user/password/token`, credentials)
                    .then(() => {
                        this.loading = false
						this.isSuccess = true
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            handleError(e) {
                this.loading = false
                this.error = e.response.data.message
            },
        }
    }
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
