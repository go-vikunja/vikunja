import type { StorybookConfig } from '@storybook/vue3-vite';

const config: StorybookConfig = {
  stories: [
    '../src/**/*.stories.@(ts|js|vue)'
  ],
  addons: [
    '@storybook/addon-links'
  ],
  framework: {
    name: '@storybook/vue3-vite',
    options: {}
  }
};
export default config;
