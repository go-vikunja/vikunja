import {describe, it, expect, beforeEach} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import DatepickerWithRange from './DatepickerWithRange.vue'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountPicker() {
    return mount(DatepickerWithRange, {
        props: {modelValue: {dateFrom: '', dateTo: ''}},
        global: {
            plugins: [i18n],
            stubs: ['RouterLink', 'Modal', 'XButton', 'BaseButton', 'Popup', 'flat-pickr'],
        },
    })
}

describe('DatepickerWithRange predefined ranges', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
    })
    it('selects Last Week range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/w-1w', 'now/w'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/w-1w', dateTo: 'now/w'})
    })

    it('selects Last Month range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/M-1M', 'now/M'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/M-1M', dateTo: 'now/M'})
    })
})
