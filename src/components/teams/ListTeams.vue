<template>
	<div class="content loader-container" v-bind:class="{ 'is-loading': loading}">
		<router-link :to="{name:'newTeam'}" class="button is-success button-right" >
			New Team
		</router-link>
		<h1>Teams</h1>
		<ul class="teams box">
			<li v-for="t in teams" :key="t.id">
				<router-link :to="{name: 'editTeam', params: {id: t.id}}">
					{{t.name}}
				</router-link>
			</li>
		</ul>
	</div>
</template>

<script>
    import auth from '../../auth'
    import router from '../../router'
    import {HTTP} from '../../http-common'
    import message from '../../message'

    export default {
        name: "ListTeams",
        data() {
            return {
                teams: [],
                error: '',
                loading: false,
            }
        },
        beforeMount() {
            // Check if the user is already logged in, if so, redirect him to the homepage
            if (!auth.user.authenticated) {
                router.push({name: 'home'})
            }
        },
        created() {
            this.loadTeams()
        },
		methods: {
			loadTeams() {
				const cancel = message.setLoading(this)

                HTTP.get(`teams`, {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}})
                    .then(response => {
                        this.$set(this, 'teams', response.data)
                        cancel()
                    })
                    .catch(e => {
                        this.handleError(e)
                    })
			},
            handleError(e) {
                message.error(e, this)
            },
		}
    }
</script>

<style lang="scss" scoped>
	.button-right{
		float: right;
	}

	ul.teams{

		padding: 0;
		margin-left: 0;

		li{
			list-style: none;
			margin: 0;
			border-bottom: 1px solid darken(#fff, 25%);

			a{
				color: #363636;
				display: block;
				padding: 0.5rem 1rem;

				&:hover{
					background: darken(#fff, 2%);
				}
			}
		}

		li:last-child{
			border-bottom: none;
		}
	}
</style>