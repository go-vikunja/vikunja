<template>
	<card class="has-no-shadow how-it-works-modal"
		  :title="$t('input.datepickerRange.math.title')">
		<p>
			{{ $t('input.datepickerRange.math.intro') }}
		</p>
		<p>
			<i18n-t keypath="input.datepickerRange.math.expression">
				<code>now</code>
				<code>||</code>
			</i18n-t>
		</p>
		<p>
			<i18n-t keypath="input.datepickerRange.math.similar">
				<a href="https://grafana.com/docs/grafana/latest/dashboards/time-range-controls/"
				   rel="noreferrer noopener nofollow" target="_blank">
					Grafana
				</a>
				<a href="https://www.elastic.co/guide/en/elasticsearch/reference/7.3/common-options.html#date-math"
				   rel="noreferrer noopener nofollow" target="_blank">
					Elasticsearch
				</a>
			</i18n-t>
		</p>
		<p>{{ $t('misc.forExample') }}</p>
		<ul>
			<li><code>+1d</code>{{ $t('input.datepickerRange.math.add1Day') }}</li>
			<li><code>-1d</code>{{ $t('input.datepickerRange.math.minus1Day') }}</li>
			<li><code>/d</code>{{ $t('input.datepickerRange.math.roundDay') }}</li>
		</ul>
		<p>{{ $t('input.datepickerRange.math.supportedUnits') }}</p>
		<table class="table">
			<tbody>
			<tr>
				<td><code>s</code></td>
				<td>{{ $t('input.datepickerRange.math.units.seconds') }}</td>
			</tr>
			<tr>
				<td><code>m</code></td>
				<td>{{ $t('input.datepickerRange.math.units.minutes') }}</td>
			</tr>
			<tr>
				<td><code>h</code></td>
				<td>{{ $t('input.datepickerRange.math.units.hours') }}</td>
			</tr>
			<tr>
				<td><code>H</code></td>
				<td>{{ $t('input.datepickerRange.math.units.hours') }}</td>
			</tr>
			<tr>
				<td><code>d</code></td>
				<td>{{ $t('input.datepickerRange.math.units.days') }}</td>
			</tr>
			<tr>
				<td><code>w</code></td>
				<td>{{ $t('input.datepickerRange.math.units.weeks') }}</td>
			</tr>
			<tr>
				<td><code>M</code></td>
				<td>{{ $t('input.datepickerRange.math.units.months') }}</td>
			</tr>
			<tr>
				<td><code>y</code></td>
				<td>{{ $t('input.datepickerRange.math.units.years') }}</td>
			</tr>
			</tbody>
		</table>

		<p>{{ $t('input.datepickerRange.math.someExamples') }}</p>
		<table class="table">
			<tbody>
			<tr>
				<td><code>now</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.now') }}</td>
			</tr>
			<tr>
				<td><code>now+24h</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.in24h') }}</td>
			</tr>
			<tr>
				<td><code>now/d</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.today') }}</td>
			</tr>
			<tr>
				<td><code>now/w</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.beginningOfThisWeek') }}</td>
			</tr>
			<tr>
				<td><code>now/w+1w</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.endOfThisWeek') }}</td>
			</tr>
			<tr>
				<td><code>now+30d</code></td>
				<td>{{ $t('input.datepickerRange.math.examples.in30Days') }}</td>
			</tr>
			<tr>
				<td><code>{{ exampleDate }}||+1M/d</code></td>
				<td>
					<i18n-t keypath="input.datepickerRange.math.examples.datePlusMonth">
						<code>{{ exampleDate }}</code>
					</i18n-t>
				</td>
			</tr>
			</tbody>
		</table>
	</card>
</template>

<script lang="ts" setup>
import {format} from 'date-fns'

const exampleDate = format(new Date(), 'yyyy-MM-dd')
</script>
