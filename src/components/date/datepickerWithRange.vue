<template>
	<div class="datepicker-with-range-container">
		<popup>
			<template #trigger="{toggle}">
				<slot name="trigger" :toggle="toggle" :buttonText="buttonText"></slot>
			</template>
			<template #content="{isOpen}">
				<div class="datepicker-with-range" :class="{'is-open': isOpen}">
					<div class="selections">
						<button @click="setDateRange(null)" :class="{'is-active': customRangeActive}">
							{{ $t('misc.custom') }}
						</button>
						<button
							v-for="(value, text) in dateRanges"
							:key="text"
							@click="setDateRange(value)"
							:class="{'is-active': from === value[0] && to === value[1]}">
							{{ $t(`input.datepickerRange.ranges.${text}`) }}
						</button>
					</div>
					<div class="flatpickr-container input-group">
						<label class="label">
							{{ $t('input.datepickerRange.from') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input class="input" type="text" v-model="from" @change="inputChanged"/>
								</div>
								<div class="control">
									<x-button icon="calendar" variant="secondary" data-toggle/>
								</div>
							</div>
						</label>
						<label class="label">
							{{ $t('input.datepickerRange.to') }}
							<div class="field has-addons">
								<div class="control is-fullwidth">
									<input class="input" type="text" v-model="to" @change="inputChanged"/>
								</div>
								<div class="control">
									<x-button icon="calendar" variant="secondary" data-toggle/>
								</div>
							</div>
						</label>
						<flat-pickr
							:config="flatPickerConfig"
							v-model="flatpickrRange"
						/>

						<p>
							{{ $t('input.datepickerRange.math.canuse') }}
							<a @click="showHowItWorks = true">{{ $t('input.datepickerRange.math.learnhow') }}</a>.
						</p>

						<modal
							@close="() => showHowItWorks = false"
							:enabled="showHowItWorks"
							transition-name="fade"
							:overflow="true"
							variant="hint-modal"
						>
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
						</modal>
					</div>
				</div>
			</template>
		</popup>
	</div>
</template>

<script lang="ts" setup>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {store} from '@/store'
import {format} from 'date-fns'
import Popup from '@/components/misc/popup.vue'

import {dateRanges} from '@/components/date/dateRanges'

const {t} = useI18n()

const emit = defineEmits(['dateChanged'])

// FIXME: This seems to always contain the default value - that breaks the picker
const weekStart = computed<number>(() => store.state.auth.settings.weekStart ?? 0)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	wrap: true,
	mode: 'range',
	locale: {
		firstDayOf7Days: weekStart.value,
	},
}))

const showHowItWorks = ref(false)
const exampleDate = format(new Date(), 'yyyy-MM-dd')

const flatpickrRange = ref('')

const from = ref('')
const to = ref('')

function emitChanged() {
	emit('dateChanged', {
		dateFrom: from.value === '' ? null : from.value,
		dateTo: to.value === '' ? null : to.value,
	})
}

function inputChanged() {
	flatpickrRange.value = ''
	emitChanged()
}

watch(
	() => flatpickrRange.value,
	(newVal: string | null) => {
		if (newVal === null) {
			return
		}

		const [fromDate, toDate] = newVal.split(' to ')

		if (typeof fromDate === 'undefined' || typeof toDate === 'undefined') {
			return
		}

		from.value = fromDate
		to.value = toDate

		emitChanged()
	},
)

function setDateRange(range: string[] | null) {
	if (range === null) {
		from.value = ''
		to.value = ''
		inputChanged()

		return
	}

	from.value = range[0]
	to.value = range[1]

	inputChanged()
}

const customRangeActive = computed<Boolean>(() => {
	return !Object.values(dateRanges).some(el => from.value === el[0] && to.value === el[1])
})

const buttonText = computed<string>(() => {
	if(from.value !== '' && to.value !== '') {
		return t('input.datepickerRange.fromto', {
			from: from.value,
			to: to.value,
		})
	}
	
	return t('task.show.select')
})
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;

	:deep(.popup) {
		z-index: 10;
		margin-top: 1rem;
		border-radius: $radius;
		border: 1px solid var(--grey-200);
		background-color: var(--white);
		box-shadow: $shadow;

		&.is-open {
			width: 500px;
			height: 320px;
		}
	}
}

.datepicker-with-range {
	display: flex;
	width: 100%;
	height: 100%;
	position: absolute;

	:deep(.flatpickr-calendar) {
		margin: 0 auto 8px;
		box-shadow: none;
	}
}

.flatpickr-container {
	width: 70%;
	border-left: 1px solid var(--grey-200);
	padding: 1rem;
	font-size: .9rem;

	// Flatpickr has no option to use it without an input field so we're hiding it instead
	:deep(input.form-control.input) {
		height: 0;
		padding: 0;
		border: 0;
	}

	.field .control :deep(.button) {
		border: 1px solid var(--input-border-color);
		height: 2.25rem;

		&:hover {
			border: 1px solid var(--input-hover-border-color);
		}
	}

	.label, .input, :deep(.button) {
		font-size: .9rem;
	}
}

.selections {
	width: 30%;
	display: flex;
	flex-direction: column;
	padding-top: .5rem;
	overflow-y: scroll;

	button {
		display: block;
		width: 100%;
		text-align: left;
		padding: .5rem 1rem;
		transition: $transition;
		font-size: .9rem;
		color: var(--text);
		background: transparent;
		border: 0;
		cursor: pointer;

		&.is-active {
			color: var(--primary);
		}

		&:hover, &.is-active {
			background-color: var(--grey-100);
		}
	}
}

.how-it-works-modal {
	font-size: 1rem;

	p {
		display: inline-block !important;
	}
}
</style>
