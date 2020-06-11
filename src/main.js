import Vue from 'vue'
import App from './App.vue'
import router from './router'

import {VERSION} from './version.json'
console.info(`Vikunja frontend version ${VERSION}`)

// Make sure the api url does not contain a / at the end
if(window.API_URL.substr(window.API_URL.length - 1, window.API_URL.length) === '/') {
	window.API_URL = window.API_URL.substr(0, window.API_URL.length - 1)
}

// Register the modal
import Modal from './components/modal/Modal'
Vue.component('modal', Modal)

// Add CSS
import './styles/vikunja.scss'

Vue.config.productionTip = false

// Notifications
import Notifications from 'vue-notification'
Vue.use(Notifications)

// Icons
import { library } from '@fortawesome/fontawesome-svg-core'
import { faSignOutAlt } from '@fortawesome/free-solid-svg-icons'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faListOl } from '@fortawesome/free-solid-svg-icons'
import { faTasks } from '@fortawesome/free-solid-svg-icons'
import { faCog } from '@fortawesome/free-solid-svg-icons'
import { faAngleRight } from '@fortawesome/free-solid-svg-icons'
import { faLayerGroup } from '@fortawesome/free-solid-svg-icons'
import { faTrashAlt } from '@fortawesome/free-solid-svg-icons'
import { faUsers } from '@fortawesome/free-solid-svg-icons'
import { faUser } from '@fortawesome/free-solid-svg-icons'
import { faLock } from '@fortawesome/free-solid-svg-icons'
import { faPen } from '@fortawesome/free-solid-svg-icons'
import { faTimes } from '@fortawesome/free-solid-svg-icons'
import { faTachometerAlt } from '@fortawesome/free-solid-svg-icons'
import { faCalendar } from '@fortawesome/free-solid-svg-icons'
import { faBars } from '@fortawesome/free-solid-svg-icons'
import { faPowerOff } from '@fortawesome/free-solid-svg-icons'
import { faCalendarWeek } from '@fortawesome/free-solid-svg-icons'
import { faExclamation } from '@fortawesome/free-solid-svg-icons'
import { faTags } from '@fortawesome/free-solid-svg-icons'
import { faChevronDown } from '@fortawesome/free-solid-svg-icons'
import { faCheck } from '@fortawesome/free-solid-svg-icons'
import { faPaste } from '@fortawesome/free-solid-svg-icons'
import { faPencilAlt } from '@fortawesome/free-solid-svg-icons'
import { faTimesCircle } from '@fortawesome/free-regular-svg-icons'
import { faCalendarAlt } from '@fortawesome/free-regular-svg-icons'
import { faCloudDownloadAlt } from '@fortawesome/free-solid-svg-icons'
import { faCloudUploadAlt } from '@fortawesome/free-solid-svg-icons'
import { faPercent } from '@fortawesome/free-solid-svg-icons'
import { faStar } from '@fortawesome/free-regular-svg-icons'
import { faAlignLeft } from '@fortawesome/free-solid-svg-icons'
import { faPaperclip } from '@fortawesome/free-solid-svg-icons'
import { faClock } from '@fortawesome/free-regular-svg-icons'
import { faHistory } from '@fortawesome/free-solid-svg-icons'
import { faSearch } from '@fortawesome/free-solid-svg-icons'
import { faCheckDouble } from '@fortawesome/free-solid-svg-icons'
import { faTh } from '@fortawesome/free-solid-svg-icons'
import { faSort } from '@fortawesome/free-solid-svg-icons'
import { faSortUp } from '@fortawesome/free-solid-svg-icons'
import { faList } from '@fortawesome/free-solid-svg-icons'
import { faEllipsisV } from '@fortawesome/free-solid-svg-icons'
import { faFilter } from '@fortawesome/free-solid-svg-icons'
import { faComments } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSignOutAlt)
library.add(faPlus)
library.add(faListOl)
library.add(faTasks)
library.add(faCog)
library.add(faAngleRight)
library.add(faLayerGroup)
library.add(faTrashAlt)
library.add(faUsers)
library.add(faUser)
library.add(faLock)
library.add(faPen)
library.add(faTimes)
library.add(faTachometerAlt)
library.add(faCalendar)
library.add(faTimesCircle)
library.add(faBars)
library.add(faPowerOff)
library.add(faCalendarWeek)
library.add(faCalendarAlt)
library.add(faExclamation)
library.add(faTags)
library.add(faChevronDown)
library.add(faCheck)
library.add(faPaste)
library.add(faPencilAlt)
library.add(faCloudDownloadAlt)
library.add(faCloudUploadAlt)
library.add(faPercent)
library.add(faStar)
library.add(faAlignLeft)
library.add(faPaperclip)
library.add(faClock)
library.add(faHistory)
library.add(faSearch)
library.add(faCheckDouble)
library.add(faComments)
library.add(faTh)
library.add(faSort)
library.add(faSortUp)
library.add(faList)
library.add(faEllipsisV)
library.add(faFilter)

Vue.component('icon', FontAwesomeIcon)

// Tooltip
import VTooltip from 'v-tooltip'
Vue.use(VTooltip)

// PWA
import './registerServiceWorker'

// Set focus
Vue.directive('focus', {
	// When the bound element is inserted into the DOM...
	inserted: el => {
		// Focus the element only if the viewport is big enough
		// auto focusing elements on mobile can be annoying since in these cases the
		// keyboard always pops up and takes half of the available space on the screen.
		// The threshhold is the same as the breakpoints in css.
		if (window.innerWidth > 769) {
			el.focus()
		}
	}
})

// Mixins
import message from './message'
import {format, formatDistance} from 'date-fns'
Vue.mixin({
	methods: {
		formatDateSince: date => {
			const currentDate = new Date()
			let formatted = '';
			if (date > currentDate) {
				formatted += 'in '
			}
			formatted += formatDistance(date, currentDate)
			if(date < currentDate) {
				formatted += ' ago'
			}

			return formatted;
		},
		formatDate: date => format(date, 'PPPPpppp'),
		error: (e, context, actions = []) => message.error(e, context, actions),
		success: (s, context, actions = []) => message.success(s, context, actions),
	}
})

// Vuex
import {store} from './store'

new Vue({
	router,
	store,
	render: h => h(App)
}).$mount('#app')
