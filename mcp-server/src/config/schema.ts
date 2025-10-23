import { z } from 'zod';

/**
 * HTTP Transport configuration schema
 */
export const HttpTransportConfigSchema = z.object({
	enabled: z.boolean().default(false),
	port: z.number().int().positive().default(3458),
	host: z.string().default('0.0.0.0'),
	enableJsonResponse: z.boolean().default(false),
});

/**
 * Redis connection configuration schema
 */
export const RedisConfigSchema = z.object({
	host: z.string().default('localhost'),
	port: z.number().int().positive().default(6379),
	password: z.string().optional(),
	url: z.string().optional(), // Alternative to host/port/password
});

/**
 * Authentication configuration schema
 */
export const AuthConfigSchema = z.object({
	tokenCacheTtl: z.number().int().positive().default(300), // 5 minutes
	tokenCacheEnabled: z.boolean().default(true),
});

/**
 * Rate limiting configuration schema
 */
export const RateLimitConfigSchema = z.object({
	windowSeconds: z.number().int().positive().default(900), // 15 minutes
	points: z.number().int().positive().default(100), // 100 requests per window
	default: z.number().int().positive().default(100), // Legacy support
	burst: z.number().int().positive().default(120), // Legacy support
	adminBypass: z.boolean().default(false),
});

/**
 * Session management configuration schema
 */
export const SessionConfigSchema = z.object({
	idleTimeoutMinutes: z.number().int().positive().default(30),
	cleanupIntervalSeconds: z.number().int().positive().default(300), // 5 minutes
	orphanedTimeoutSeconds: z.number().int().positive().default(60), // 1 minute
});

/**
 * Logging configuration schema
 */
export const LoggingConfigSchema = z.object({
	level: z.enum(['error', 'warn', 'info', 'debug']).default('info'),
	format: z.enum(['json', 'simple']).default('json'),
});

/**
 * LLM configuration schema
 */
export const LLMConfigSchema = z.object({
	provider: z.enum(['openai', 'anthropic', 'ollama']),
	apiKey: z.string().optional(),
	endpoint: z.string().url().optional(),
});

/**
 * Complete configuration schema
 */
export const ConfigSchema = z.object({
	vikunjaApiUrl: z.string().url(),
	port: z.number().int().positive().default(3457), // stdio mode port (legacy)
	httpTransport: HttpTransportConfigSchema,
	redis: RedisConfigSchema,
	auth: AuthConfigSchema,
	rateLimits: RateLimitConfigSchema,
	session: SessionConfigSchema,
	llm: LLMConfigSchema.optional(),
	logging: LoggingConfigSchema,
});

export type Config = z.infer<typeof ConfigSchema>;
export type HttpTransportConfig = z.infer<typeof HttpTransportConfigSchema>;
export type RedisConfig = z.infer<typeof RedisConfigSchema>;
export type AuthConfig = z.infer<typeof AuthConfigSchema>;
export type RateLimitConfig = z.infer<typeof RateLimitConfigSchema>;
export type SessionConfig = z.infer<typeof SessionConfigSchema>;
export type LoggingConfig = z.infer<typeof LoggingConfigSchema>;
export type LLMConfig = z.infer<typeof LLMConfigSchema>;
