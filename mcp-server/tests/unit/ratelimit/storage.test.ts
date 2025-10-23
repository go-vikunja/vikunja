import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { RedisStorage } from '../../../src/ratelimit/storage.js';
import { InternalError } from '../../../src/utils/errors.js';
import { getRedisConnection } from '../../../src/utils/redis-connection.js';

// Mock redis-connection module
vi.mock('../../../src/utils/redis-connection.js', () => ({
  getRedisConnection: vi.fn(),
}));

// Mock config
vi.mock('../../../src/config/index.js', () => ({
  config: {
    redis: {
      host: 'localhost',
      port: 6379,
      password: undefined,
    },
  },
}));

// Mock logger
vi.mock('../../../src/utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

describe('RedisStorage', () => {
  let storage: RedisStorage;
  let mockRedisInstance: any;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Create mock Redis instance
    mockRedisInstance = {
      get: vi.fn(),
      set: vi.fn(),
      setex: vi.fn(),
      del: vi.fn(),
      zadd: vi.fn(),
      zremrangebyscore: vi.fn(),
      zcard: vi.fn(),
      ping: vi.fn(),
      quit: vi.fn(),
      expire: vi.fn(),
    };
    
    // Mock getRedisConnection to return our mock instance
    vi.mocked(getRedisConnection).mockResolvedValue(mockRedisInstance as any);
    
    storage = new RedisStorage();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('connect', () => {
    it('should connect to Redis successfully', async () => {
      await storage.connect();
      
      expect(getRedisConnection).toHaveBeenCalled();
    });

    it('should retry connection on failure', async () => {
      // Mock getRedisConnection to fail
      vi.mocked(getRedisConnection).mockRejectedValue(new Error('Connection failed'));
      
      // RedisStorage wraps the error in an InternalError
      const newStorage = new RedisStorage();
      await expect(newStorage.connect()).rejects.toThrow(InternalError);
      await expect(newStorage.connect()).rejects.toThrow('Failed to connect to Redis for rate limiting');
    });

    it('should throw error after max retries', async () => {
      // Mock getRedisConnection to always fail
      vi.mocked(getRedisConnection).mockRejectedValue(new Error('Connection failed'));
      
      const newStorage = new RedisStorage();
      
      await expect(newStorage.connect()).rejects.toThrow(InternalError);
    });
  });

  describe('disconnect', () => {
    it('should disconnect from Redis', async () => {
      // First connect
      await storage.connect();
      
      // Then disconnect - should just clear reference, not call quit on shared connection
      await storage.disconnect();
      
      // Verify quit was NOT called (connection is shared)
      expect(mockRedisInstance.quit).not.toHaveBeenCalled();
    });
  });

  describe('Basic operations', () => {
    beforeEach(async () => {
      // Connect before each test
      await storage.connect();
    });

    describe('get', () => {
      it('should get value by key', async () => {
        mockRedisInstance.get.mockResolvedValue('test-value');
        
        const result = await storage.get('test-key');
        
        expect(mockRedisInstance.get).toHaveBeenCalledWith('test-key');
        expect(result).toBe('test-value');
      });

      it('should return null for non-existent key', async () => {
        mockRedisInstance.get.mockResolvedValue(null);
        
        const result = await storage.get('non-existent');
        
        expect(result).toBeNull();
      });
    });

    describe('set', () => {
      it('should set value without TTL', async () => {
        mockRedisInstance.set.mockResolvedValue('OK');
        
        await storage.set('test-key', 'test-value');
        
        expect(mockRedisInstance.set).toHaveBeenCalledWith('test-key', 'test-value');
      });

      it('should set value with TTL', async () => {
        mockRedisInstance.setex.mockResolvedValue('OK');
        
        await storage.set('test-key', 'test-value', 60);
        
        expect(mockRedisInstance.setex).toHaveBeenCalledWith('test-key', 60, 'test-value');
      });
    });

    describe('del', () => {
      it('should delete key', async () => {
        mockRedisInstance.del.mockResolvedValue(1);
        
        await storage.del('test-key');
        
        expect(mockRedisInstance.del).toHaveBeenCalledWith('test-key');
      });
    });

    describe('zadd', () => {
      it('should add member to sorted set', async () => {
        mockRedisInstance.zadd.mockResolvedValue(1);
        
        await storage.zadd('test-set', 100, 'member1');
        
        expect(mockRedisInstance.zadd).toHaveBeenCalledWith('test-set', 100, 'member1');
      });
    });

    describe('zremrangebyscore', () => {
      it('should remove members by score range', async () => {
        mockRedisInstance.zremrangebyscore.mockResolvedValue(5);
        
        await storage.zremrangebyscore('test-set', 0, 50);
        
        expect(mockRedisInstance.zremrangebyscore).toHaveBeenCalledWith('test-set', 0, 50);
      });
    });

    describe('zcard', () => {
      it('should get cardinality of sorted set', async () => {
        mockRedisInstance.zcard.mockResolvedValue(10);
        
        const result = await storage.zcard('test-set');
        
        expect(mockRedisInstance.zcard).toHaveBeenCalledWith('test-set');
        expect(result).toBe(10);
      });
    });
  });

  describe('isHealthy', () => {
    it('should return true when Redis is healthy', async () => {
      // Connect first
      await storage.connect();
      
      mockRedisInstance.ping.mockResolvedValue('PONG');
      
      const healthy = await storage.isHealthy();
      
      expect(healthy).toBe(true);
      expect(mockRedisInstance.ping).toHaveBeenCalled();
    });

    it('should return false when Redis is not connected', async () => {
      const healthy = await storage.isHealthy();
      
      expect(healthy).toBe(false);
    });

    it('should return false when ping fails', async () => {
      // Connect first
      await storage.connect();
      
      mockRedisInstance.ping.mockRejectedValue(new Error('Connection lost'));
      
      const healthy = await storage.isHealthy();
      
      expect(healthy).toBe(false);
    });
  });

  describe('Connection error handling', () => {
    it('should throw error when operation called without connection', async () => {
      await expect(storage.get('test-key')).rejects.toThrow();
    });
  });
});
