declare global {
	interface Window {
		API_URL: string;
		SENTRY_ENABLED: boolean;
		SENTRY_DSN: string;
	}
}