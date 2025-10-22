import { ConfigSchema } from './schema.js';
import type { Config } from './schema.js';

/**
 * Load configuration from environment variables
 */
function loadConfig(): Config {
	const rawConfig = {
		vikunjaApiUrl: process.env['VIKUNJA_API_URL'] ?? 'http://localhost:3456',
		port: process.env['MCP_PORT'] ? parseInt(process.env['MCP_PORT'], 10) : 3457,
		httpTransport: {
			enabled: process.env['MCP_HTTP_ENABLED'] === 'true',
			port: process.env['MCP_HTTP_PORT'] ? parseInt(process.env['MCP_HTTP_PORT'], 10) : 3458,
			host: process.env['MCP_HTTP_HOST'] ?? '0.0.0.0',
		},
		redis: {
			host: process.env['REDIS_HOST'] ?? 'localhost',
			port: process.env['REDIS_PORT'] ? parseInt(process.env['REDIS_PORT'], 10) : 6379,
			password: process.env['REDIS_PASSWORD'],
			url: process.env['REDIS_URL'],
		},
		auth: {
			tokenCacheTtl: process.env['AUTH_TOKEN_CACHE_TTL']
				? parseInt(process.env['AUTH_TOKEN_CACHE_TTL'], 10)
				: 300,
			tokenCacheEnabled: process.env['AUTH_TOKEN_CACHE_ENABLED'] !== 'false',
		},
		rateLimits: {
			windowSeconds: process.env['RATE_LIMIT_WINDOW_SECONDS']
				? parseInt(process.env['RATE_LIMIT_WINDOW_SECONDS'], 10)
				: 900,
			points: process.env['RATE_LIMIT_POINTS']
				? parseInt(process.env['RATE_LIMIT_POINTS'], 10)
				: 100,
			default: process.env['RATE_LIMIT_DEFAULT']
				? parseInt(process.env['RATE_LIMIT_DEFAULT'], 10)
				: 100,
			burst: process.env['RATE_LIMIT_BURST']
				? parseInt(process.env['RATE_LIMIT_BURST'], 10)
				: 120,
			adminBypass: process.env['RATE_LIMIT_ADMIN_BYPASS'] === 'true',
		},
		session: {
			idleTimeoutMinutes: process.env['SESSION_IDLE_TIMEOUT_MINUTES']
				? parseInt(process.env['SESSION_IDLE_TIMEOUT_MINUTES'], 10)
				: 30,
			cleanupIntervalSeconds: process.env['SESSION_CLEANUP_INTERVAL_SECONDS']
				? parseInt(process.env['SESSION_CLEANUP_INTERVAL_SECONDS'], 10)
				: 300,
			orphanedTimeoutSeconds: process.env['SESSION_ORPHANED_TIMEOUT_SECONDS']
				? parseInt(process.env['SESSION_ORPHANED_TIMEOUT_SECONDS'], 10)
				: 60,
		},
		llm: process.env['LLM_PROVIDER']
			? {
					provider: process.env['LLM_PROVIDER'] as 'openai' | 'anthropic' | 'ollama',
					apiKey: process.env['LLM_API_KEY'],
					endpoint: process.env['LLM_ENDPOINT'],
			  }
			: undefined,
		logging: {
			level: (process.env['LOG_LEVEL'] ?? 'info') as 'error' | 'warn' | 'info' | 'debug',
			format: (process.env['LOG_FORMAT'] ?? 'json') as 'json' | 'simple',
		},
	};

	return ConfigSchema.parse(rawConfig);
}

/**
 * Singleton configuration instance
 */
export const config = loadConfig();
