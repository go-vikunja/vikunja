import { mount } from 'cypress/vue'

type MountParams = Parameters<typeof mount>;
type OptionsParam = MountParams[1];

declare global {
  namespace Cypress {
    interface Chainable {
      mount: typeof mount;
    }
  }
}
