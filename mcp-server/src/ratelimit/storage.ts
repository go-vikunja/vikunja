import type Redis from 'ioredis';
import { logger } from '../utils/logger.js';
import { getRedisConnection } from '../utils/redis-connection.js';
import { InternalError } from '../utils/errors.js';

/**
 * Redis storage for rate limiting
 * Uses shared Redis connection from RedisConnectionManager
 */
export class RedisStorage {
  private client: Redis | null = null;

  /**
   * Connect to Redis using shared connection manager
   */
  async connect(): Promise<void> {
    try {
      this.client = await getRedisConnection();
      logger.info('Using shared Redis connection for rate limiting');
    } catch (error) {
      logger.error('Failed to get Redis connection for rate limiting', { error });
      throw new InternalError('Failed to connect to Redis for rate limiting');
    }
  }

  /**
   * Disconnect from Redis
   * Note: Redis connection is managed by RedisConnectionManager singleton
   */
  async disconnect(): Promise<void> {
    // Redis connection is shared and managed by RedisConnectionManager
    // Just clear our reference
    this.client = null;
    logger.info('RedisStorage disconnected (shared connection retained)');
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
   * Set expiry on a key (in seconds)
   */
  async expire(key: string, ttl: number): Promise<void> {
    this.ensureConnected();
    await this.client!.expire(key, ttl);
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
}
