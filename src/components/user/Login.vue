<template>
	<div class="container has-text-centered">
		<div class="column is-4 is-offset-4">
			<img src="images/logo-full.svg"/>
			<h2 class="title">Login</h2>
			<div class="box">
				<form id="loginform" @submit.prevent="submit">
					<div class="field">
						<div class="control">
							<input type="text" class="input" name="username" placeholder="Username" v-model="credentials.username" required>
						</div>
					</div>
					<div class="field">
						<div class="control">
							<input type="password" class="input" name="password" placeholder="Password" v-model="credentials.password" required>
						</div>
					</div>

					<div class="field is-grouped">
						<div class="control">
							<button type="submit" class="button is-primary" v-bind:class="{ 'is-loading': loading}">Login</button>
							<router-link :to="{ name: 'register' }" class="button">Register</router-link>
						</div>
					</div>
					<div class="notification is-danger" v-if="error">
						{{ error }}
					</div>
				</form>

			</div>
		</div>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'

    export default {
        data() {
            return {
                credentials: {
                    username: '',
                    password: ''
                },
                error: '',
                loading: false
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        methods: {
            submit() {
                this.loading = true
                this.error = ''
                let credentials = {
                    username: this.credentials.username,
                    password: this.credentials.password
                }

                auth.login(this, credentials, 'home')
            }
        }
    }
</script>

<style scoped>
	.button {
		margin: 0 0.4em 0 0;
	}
</style>
