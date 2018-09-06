<template>
	<div>
		<h2>
			Login
		</h2>
		<form id="loginform" @submit.prevent="submit">
			<input type="text" name="username" placeholder="Username" v-model="credentials.username" required>
			<input type="password" name="password" placeholder="Password" v-model="credentials.password" required>
			<button type="submit" class="ui fluid large blue submit button">Login</button>
			<div class="ui info message" v-if="loading">
				<icon name="refresh" spin></icon>&nbsp;&nbsp;
				Loading...
			</div>
			<div class="ui error message" v-if="error" style="display: block;">
				{{ error }}
			</div>
		</form>
	</div>
</template>

<script>
    import auth from '../auth'
    import router from '../router'

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

</style>
