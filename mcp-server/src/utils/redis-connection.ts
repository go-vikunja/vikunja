import Redis from 'ioredis';
import { config } from '../config/index.js';
import { logger } from './logger.js';
import { InternalError } from './errors.js';

/**
 * Singleton Redis connection manager
 * 
 * Provides a shared Redis connection to prevent resource exhaustion
 * from multiple independent Redis clients.
 * 
 * Features:
 * - Singleton pattern - only one connection instance
 * - Automatic connection with retry logic
 * - Graceful disconnection
 * - Health checking
 * 
 * @example
 * ```typescript
 * const redis = await RedisConnectionManager.getConnection();
 * await redis.set('key', 'value');
 * ```
 */
export class RedisConnectionManager {
	private static instance: Redis | null = null;
	private static connecting: Promise<Redis> | null = null;
	private static readonly maxRetries = 5;
	private static readonly retryDelayMs = 1000;

	/**
	 * Get the singleton Redis connection
	 * Creates connection on first call, returns existing connection on subsequent calls
	 */
	static async getConnection(): Promise<Redis> {
		// Return existing instance if available
		if (this.instance) {
			return this.instance;
		}

		// Wait for in-progress connection
		if (this.connecting) {
			return this.connecting;
		}

		// Create new connection
		this.connecting = this.connect();
		
		try {
			this.instance = await this.connecting;
			return this.instance;
		} finally {
			this.connecting = null;
		}
	}

	/**
	 * Connect to Redis with retry logic
	 */
	private static async connect(): Promise<Redis> {
		let retries = 0;

		while (retries < this.maxRetries) {
			try {
				const client = await this.createClient();
				logger.info('Redis connection established (singleton)');
				return client;
			} catch (error) {
				retries++;
				logger.warn(`Redis connection attempt ${retries}/${this.maxRetries} failed`, {
					error,
				});

				if (retries >= this.maxRetries) {
					throw new InternalError('Failed to connect to Redis after multiple attempts');
				}

				await this.sleep(this.retryDelayMs * retries);
			}
		}

		throw new InternalError('Failed to connect to Redis');
	}

	/**
	 * Create Redis client with proper configuration
	 */
	private static async createClient(): Promise<Redis> {
		// Try URL-based connection first (recommended)
		if (config.redis.url) {
			const client = new Redis(config.redis.url, {
				retryStrategy: (times: number) => {
					if (times > this.maxRetries) {
						return null; // Stop retrying
					}
					return Math.min(times * this.retryDelayMs, 3000);
				},
				maxRetriesPerRequest: 3,
				enableReadyCheck: true,
			});

			await this.waitForReady(client);
			return client;
		}

		// Fallback to host/port configuration
		const options: {
			host: string;
			port: number;
			password?: string;
			retryStrategy: (times: number) => number | null;
			maxRetriesPerRequest: number;
			enableReadyCheck: boolean;
		} = {
			host: config.redis.host,
			port: config.redis.port,
			retryStrategy: (times: number) => {
				if (times > this.maxRetries) {
					return null;
				}
				return Math.min(times * this.retryDelayMs, 3000);
			},
			maxRetriesPerRequest: 3,
			enableReadyCheck: true,
		};

		if (config.redis.password) {
			options.password = config.redis.password;
		}

		const client = new Redis(options);
		await this.waitForReady(client);
		return client;
	}

	/**
	 * Wait for Redis connection to be ready
	 */
	private static async waitForReady(client: Redis): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			client.once('ready', () => {
				resolve();
			});

			client.once('error', (err: Error) => {
				reject(err);
			});

			// Timeout after 10 seconds
			setTimeout(() => {
				reject(new Error('Redis connection timeout'));
			}, 10000);
		});
	}

	/**
	 * Disconnect from Redis
	 * Should be called during application shutdown
	 */
	static async disconnect(): Promise<void> {
		if (this.instance) {
			await this.instance.quit();
			this.instance = null;
			logger.info('Redis connection closed (singleton)');
		}
	}

	/**
	 * Check if Redis connection is healthy
	 */
	static async isHealthy(): Promise<boolean> {
		try {
			if (!this.instance) {
				return false;
			}
			await this.instance.ping();
			return true;
		} catch {
			return false;
		}
	}

	/**
	 * Get current connection status
	 */
	static isConnected(): boolean {
		return this.instance !== null && this.instance.status === 'ready';
	}

	/**
	 * Sleep helper for retry logic
	 */
	private static sleep(ms: number): Promise<void> {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}
}

/**
 * Get the singleton Redis connection
 * Convenience export for common usage
 */
export async function getRedisConnection(): Promise<Redis> {
	return RedisConnectionManager.getConnection();
}

/**
 * Disconnect the singleton Redis connection
 * Convenience export for shutdown
 */
export async function disconnectRedis(): Promise<void> {
	return RedisConnectionManager.disconnect();
}
