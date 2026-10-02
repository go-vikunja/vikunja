<template>
	<details
		v-if="total > 0"
		class="mbs-4"
	>
		<summary>{{ title }} ({{ total }})</summary>
		<table class="table is-striped is-fullwidth">
			<thead>
				<tr>
					<th>{{ $t('user.auth.username') }}</th>
					<th>{{ $t('manage.users.name') }}</th>
					<th>{{ $t('user.auth.email') }}</th>
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="(entry, index) in entries"
					:key="index"
				>
					<td>{{ entry.username }}</td>
					<td>{{ entry.name }}</td>
					<td>{{ entry.email }}</td>
				</tr>
			</tbody>
		</table>
	</details>
</template>

<script setup lang="ts">
import type {ImportEntry} from '@/client/queries/manage'

defineProps<{
	title: string
	entries: ImportEntry[]
	total: number
}>()
</script>
