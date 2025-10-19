import Redis from 'ioredis';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';
import { InternalError } from '../utils/errors.js';

/**
 * Redis storage for rate limiting
 */
export class RedisStorage {
  private client: Redis | null = null;
  private readonly maxRetries = 5;
  private readonly retryDelayMs = 1000;

  /**
   * Connect to Redis
   */
  async connect(): Promise<void> {
    let retries = 0;

    while (retries < this.maxRetries) {
      try {
        const redisOptions: {
          host: string;
          port: number;
          password?: string;
          retryStrategy: (times: number) => number | null;
        } = {
          host: config.redis.host,
          port: config.redis.port,
          retryStrategy: (times: number) => {
            if (times > this.maxRetries) {
              return null; // Stop retrying
            }
            return Math.min(times * this.retryDelayMs, 3000);
          },
        };

        if (config.redis.password) {
          redisOptions.password = config.redis.password;
        }

        this.client = new Redis(redisOptions);

        // Wait for connection
        await new Promise<void>((resolve, reject) => {
          if (!this.client) {
            reject(new Error('Redis client is null'));
            return;
          }

          this.client.once('ready', () => {
            logger.info('Connected to Redis');
            resolve();
          });

          this.client.once('error', (err: Error) => {
            reject(err);
          });
        });

        return;
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
  }

  /**
   * Disconnect from Redis
   */
  async disconnect(): Promise<void> {
    if (this.client) {
      await this.client.quit();
      this.client = null;
      logger.info('Disconnected from Redis');
    }
  }

  /**
   * Get value by key
   */
  async get(key: string): Promise<string | null> {
    this.ensureConnected();
    return this.client!.get(key);
  }

  /**
   * Set value with optional TTL
   */
  async set(key: string, value: string, ttl?: number): Promise<void> {
    this.ensureConnected();
    if (ttl) {
      await this.client!.setex(key, ttl, value);
    } else {
      await this.client!.set(key, value);
    }
  }

  /**
   * Delete key
   */
  async del(key: string): Promise<void> {
    this.ensureConnected();
    await this.client!.del(key);
  }

  /**
   * Add member to sorted set
   */
  async zadd(key: string, score: number, member: string): Promise<void> {
    this.ensureConnected();
    await this.client!.zadd(key, score, member);
  }

  /**
   * Remove members from sorted set by score range
   */
  async zremrangebyscore(key: string, min: number, max: number): Promise<void> {
    this.ensureConnected();
    await this.client!.zremrangebyscore(key, min, max);
  }

  /**
   * Get cardinality of sorted set
   */
  async zcard(key: string): Promise<number> {
    this.ensureConnected();
    return this.client!.zcard(key);
  }

  /**
   * Check Redis health
   */
  async isHealthy(): Promise<boolean> {
    try {
      if (!this.client) {
        return false;
      }
      await this.client.ping();
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Ensure Redis is connected
   */
  private ensureConnected(): void {
    if (!this.client) {
      throw new InternalError('Redis client is not connected');
    }
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}
