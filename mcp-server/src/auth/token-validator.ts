import Redis from 'ioredis';
import crypto from 'crypto';
import axios from 'axios';
import { config } from '../config/index.js';
import { AuthenticationError } from '../utils/errors.js';
import { logAuth, logError } from '../utils/logger.js';

/**
 * User context after successful token validation
 */
export interface UserContext {
	userId: number;
	username: string;
	email?: string;
	permissions: string[];
	tokenScopes?: string[];
	validatedAt: Date;
}

/**
 * Token validator with Redis caching
 */
export class TokenValidator {
	private redis: Redis | null = null;
	private inMemoryCache = new Map<string, { context: UserContext; expiresAt: number }>();

	constructor() {
		this.initializeRedis();
	}

	/**
	 * Initialize Redis connection
	 */
	private initializeRedis(): void {
		if (!config.auth.tokenCacheEnabled) {
			return;
		}

		try {
			const redisUrl = config.redis.url;
			if (redisUrl) {
				this.redis = new Redis(redisUrl);
			} else {
				const options: {
					host: string;
					port: number;
					password?: string;
				} = {
					host: config.redis.host,
					port: config.redis.port,
				};
				if (config.redis.password) {
					options.password = config.redis.password;
				}
				this.redis = new Redis(options);
			}

			this.redis.on('error', (error) => {
				logError(error, { context: 'redis-connection', component: 'TokenValidator' });
				// Fallback to in-memory cache
				this.redis = null;
			});

			this.redis.on('connect', () => {
				logAuth('token_cached', undefined, { message: 'Redis connected for token caching' });
			});
		} catch (error) {
			logError(error as Error, { context: 'redis-initialization', component: 'TokenValidator' });
			this.redis = null;
		}
	}

	/**
	 * Hash token for secure storage
	 */
	private hashToken(token: string): string {
		return crypto.createHash('sha256').update(token).digest('hex');
	}

	/**
	 * Validate token against Vikunja API
	 */
	async validateToken(token: string): Promise<UserContext> {
		const tokenHash = this.hashToken(token);

		// Check cache first
		const cached = await this.getCachedContext(tokenHash);
		if (cached) {
			logAuth('token_cached', tokenHash, { userId: cached.userId });
			return cached;
		}

		// Validate against Vikunja API
		try {
			const response = await axios.get(`${config.vikunjaApiUrl}/api/v1/user`, {
				headers: {
					Authorization: `Bearer ${token}`,
				},
			});

			const userData = response.data;
			const userContext: UserContext = {
				userId: userData.id,
				username: userData.username,
				email: userData.email,
				permissions: ['read', 'write'], // Vikunja doesn't expose granular permissions via API
				validatedAt: new Date(),
			};

			// Cache the result
			await this.cacheContext(tokenHash, userContext);

			logAuth('token_validated', tokenHash, { userId: userContext.userId });
			return userContext;
		} catch (error) {
			if (axios.isAxiosError(error)) {
				if (error.response?.status === 401) {
					logAuth('auth_failed', tokenHash, { reason: 'invalid_token' });
					throw new AuthenticationError('Invalid or expired token', { code: 'INVALID_TOKEN' });
				}
				if (error.response?.status === 403) {
					logAuth('auth_failed', tokenHash, { reason: 'forbidden' });
					throw new AuthenticationError('Access forbidden', { code: 'FORBIDDEN' });
				}
			}

			logError(error as Error, { context: 'token-validation', tokenHash });
			throw new AuthenticationError('Token validation failed', { code: 'VALIDATION_ERROR' });
		}
	}

	/**
	 * Get cached user context
	 */
	private async getCachedContext(tokenHash: string): Promise<UserContext | null> {
		// Try Redis first
		if (this.redis) {
			try {
				const cached = await this.redis.get(`auth:token:${tokenHash}`);
				if (cached) {
					const parsed = JSON.parse(cached);
					return {
						...parsed,
						validatedAt: new Date(parsed.validatedAt),
					};
				}
			} catch (error) {
				logError(error as Error, { context: 'redis-get', tokenHash });
				// Fall through to in-memory cache
			}
		}

		// Try in-memory cache
		const inMemory = this.inMemoryCache.get(tokenHash);
		if (inMemory && inMemory.expiresAt > Date.now()) {
			return inMemory.context;
		}

		// Clean up expired in-memory entries
		if (inMemory && inMemory.expiresAt <= Date.now()) {
			this.inMemoryCache.delete(tokenHash);
		}

		return null;
	}

	/**
	 * Cache user context
	 */
	private async cacheContext(tokenHash: string, context: UserContext): Promise<void> {
		const ttl = config.auth.tokenCacheTtl;

		// Cache in Redis
		if (this.redis) {
			try {
				await this.redis.setex(
					`auth:token:${tokenHash}`,
					ttl,
					JSON.stringify(context)
				);
			} catch (error) {
				logError(error as Error, { context: 'redis-set', tokenHash });
				// Fall through to in-memory cache
			}
		}

		// Always cache in memory as fallback
		this.inMemoryCache.set(tokenHash, {
			context,
			expiresAt: Date.now() + ttl * 1000,
		});
	}

	/**
	 * Invalidate cached token
	 */
	async invalidateToken(token: string): Promise<void> {
		const tokenHash = this.hashToken(token);

		// Remove from Redis
		if (this.redis) {
			try {
				await this.redis.del(`auth:token:${tokenHash}`);
			} catch (error) {
				logError(error as Error, { context: 'redis-del', tokenHash });
			}
		}

		// Remove from in-memory cache
		this.inMemoryCache.delete(tokenHash);
	}

	/**
	 * Clean up resources
	 */
	async close(): Promise<void> {
		if (this.redis) {
			await this.redis.quit();
		}
		this.inMemoryCache.clear();
	}
}
