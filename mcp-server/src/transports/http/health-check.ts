import type { Request, Response } from 'express';
import { config } from '../../config/index.js';
import type { SessionManager } from './session-manager.js';
import axios from 'axios';
import Redis from 'ioredis';
import { logError } from '../../utils/logger.js';

/**
 * Health check configuration
 */
export interface HealthCheckConfig {
	sessionManager: SessionManager;
}

/**
 * Health check response status
 */
type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

/**
 * Individual component health check
 */
interface ComponentHealth {
	status: HealthStatus;
	latency?: number;
	error?: string;
}

/**
 * Overall health check response
 */
interface HealthResponse {
	status: HealthStatus;
	timestamp: string;
	version: string;
	uptime: number;
	checks: {
		redis?: ComponentHealth;
		vikunja_api?: ComponentHealth;
	};
	sessions: {
		active: number;
		total_created?: number;
		total_terminated?: number;
	};
}

/**
 * Health check endpoint handler
 * 
 * Checks:
 * 1. Redis connection (if enabled)
 * 2. Vikunja API connectivity
 * 3. Session manager stats
 * 
 * Returns:
 * - 200 OK if all checks pass
 * - 503 Service Unavailable if any critical check fails
 */
export class HealthCheckHandler {
	private readonly sessionManager: SessionManager;
	private readonly startTime: number;
	private redis: Redis | null = null;

	constructor(config: HealthCheckConfig) {
		this.sessionManager = config.sessionManager;
		this.startTime = Date.now();

		// Initialize Redis connection for health checks
		if (config.sessionManager) {
			this.initializeRedis();
		}
	}

	/**
	 * Initialize Redis connection for health checks
	 */
	private initializeRedis(): void {
		try {
			const redisUrl = config.redis.url;
			if (redisUrl) {
				this.redis = new Redis(redisUrl);
			} else if (config.redis.host && config.redis.port) {
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

			if (this.redis) {
				this.redis.on('error', (error) => {
					logError(error, { context: 'health-check-redis' });
					this.redis = null;
				});
			}
		} catch (error) {
			logError(error as Error, { context: 'health-check-redis-init' });
			this.redis = null;
		}
	}

	/**
	 * Check Redis connectivity
	 */
	private async checkRedis(): Promise<ComponentHealth> {
		if (!this.redis) {
			return {
				status: 'degraded',
				error: 'Redis not configured or connection failed',
			};
		}

		const start = Date.now();
		try {
			await this.redis.ping();
			const latency = Date.now() - start;
			return {
				status: 'healthy',
				latency,
			};
		} catch (error) {
			return {
				status: 'unhealthy',
				error: error instanceof Error ? error.message : 'Redis check failed',
			};
		}
	}

	/**
	 * Check Vikunja API connectivity
	 */
	private async checkVikunjaAPI(): Promise<ComponentHealth> {
		const start = Date.now();
		try {
			// Check API info endpoint (doesn't require authentication)
			const response = await axios.get(`${config.vikunjaApiUrl}/api/v1/info`, {
				timeout: 5000, // 5 second timeout
			});

			const latency = Date.now() - start;

			if (response.status === 200) {
				return {
					status: 'healthy',
					latency,
				};
			} else {
				return {
					status: 'degraded',
					error: `Unexpected status: ${response.status}`,
					latency,
				};
			}
		} catch (error) {
			return {
				status: 'unhealthy',
				error: error instanceof Error ? error.message : 'Vikunja API check failed',
			};
		}
	}

	/**
	 * Get session statistics
	 */
	private getSessionStats(): HealthResponse['sessions'] {
		const metrics = this.sessionManager.getMetrics();
		return {
			active: metrics.activeSessions,
			total_created: metrics.totalCreated,
			total_terminated: metrics.totalTerminated,
		};
	}

	/**
	 * Determine overall health status
	 */
	private determineOverallStatus(checks: HealthResponse['checks']): HealthStatus {
		const statuses = Object.values(checks)
			.map((check) => check?.status)
			.filter((status): status is HealthStatus => status !== undefined);

		if (statuses.includes('unhealthy')) {
			return 'unhealthy';
		}
		if (statuses.includes('degraded')) {
			return 'degraded';
		}
		return 'healthy';
	}

	/**
	 * Handle health check request
	 */
	async handleRequest(_req: Request, res: Response): Promise<void> {
		try {
			// Run all health checks
			const [redisCheck, vikunjCheck] = await Promise.all([
				this.checkRedis(),
				this.checkVikunjaAPI(),
			]);

			const checks = {
				redis: redisCheck,
				vikunja_api: vikunjCheck,
			};

			const response: HealthResponse = {
				status: this.determineOverallStatus(checks),
				timestamp: new Date().toISOString(),
				version: '1.1.0', // TODO: Read from package.json
				uptime: Math.floor((Date.now() - this.startTime) / 1000),
				checks,
				sessions: this.getSessionStats(),
			};

			// Return appropriate HTTP status
			const httpStatus = response.status === 'healthy' ? 200 : 503;

			res.status(httpStatus).json(response);
		} catch (error) {
			logError(error as Error, { context: 'health-check' });
			res.status(500).json({
				status: 'unhealthy',
				timestamp: new Date().toISOString(),
				version: '1.1.0',
				uptime: Math.floor((Date.now() - this.startTime) / 1000),
				checks: {},
				sessions: { active: 0 },
				error: error instanceof Error ? error.message : 'Health check failed',
			});
		}
	}

	/**
	 * Clean up resources
	 */
	async close(): Promise<void> {
		if (this.redis) {
			await this.redis.quit();
		}
	}
}
