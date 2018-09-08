<template>
	<div>
		<div class="full-loader-wrapper" v-if="loading">
			<div class="half-circle-spinner">
				<div class="circle circle-1"></div>
				<div class="circle circle-2"></div>
			</div>
		</div>
		<div class="content">
			<h1>{{ list.title }}</h1>
			<ul>
				<li v-for="l in list.tasks" v-bind:key="l.id">{{l.text}}</li>
			</ul>
		</div>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'

    export default {
        data() {
            return {
                listID: this.$route.params.id,
				list: {},
                error: '',
                loading: false
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        created () {
            this.loadList()
        },
        watch: {
            // call again the method if the route changes
            '$route': 'loadList'
        },
        methods: {
            loadList() {
                this.loading = true

                HTTP.get(`lists/` + this.$route.params.id, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.loading = false
                        // This adds a new elemednt "list" to our object which contains all lists
                        this.$set(this, 'list', response.data)
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
            },
            handleError(e) {
                this.loading = false
                message.error(e, this)
            }
        }
    }
</script>

<style scoped>

</style>