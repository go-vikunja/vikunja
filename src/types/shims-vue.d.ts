declare module '*.vue' {
  import Vue from 'vue'
  export default Vue
}

// https://next.vuex.vuejs.org/guide/migrating-to-4-0-from-3-x.html#typescript-support
import { ComponentCustomProperties } from 'vue'
import { Store } from 'vuex'

declare module '@vue/runtime-core' {
  // Declare your own store states.
  interface State {
    count: number
  }

  interface ComponentCustomProperties {
    $store: Store<State>
  }
}