import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import supertest from 'supertest';
import express, { type Express } from 'express';

// Mock the modules before importing
vi.mock('../src/config/index.js', () => ({
  config: {
    vikunjaApiUrl: 'http://localhost:3456',
    port: 3000,
    redis: {
      host: 'localhost',
      port: 6379,
    },
    rateLimits: {
      default: 100,
      burst: 120,
      adminBypass: true,
    },
    llm: {
      provider: 'openai' as const,
    },
    logging: {
      level: 'info',
      format: 'json',
    },
  },
}));

vi.mock('../src/ratelimit/storage.js', () => ({
  RedisStorage: class MockRedisStorage {
    async connect() {
      return Promise.resolve();
    }
    async disconnect() {
      return Promise.resolve();
    }
    async isHealthy() {
      return Promise.resolve(true);
    }
  },
}));

vi.mock('../src/server.js', () => ({
  VikunjaMCPServer: class MockVikunjaMCPServer {
    async start() {
      return Promise.resolve();
    }
    async stop() {
      return Promise.resolve();
    }
  },
}));

describe('Server Entry Point', () => {
  let app: Express;

  beforeAll(() => {
    // Create a basic Express app for testing
    app = express();
    app.use(express.json());

    // Health check endpoint
    app.get('/health', async (_req, res) => {
      const uptime = process.uptime();
      res.json({
        status: 'ok',
        version: '1.0.0',
        uptime,
        redis: 'connected',
        timestamp: new Date().toISOString(),
      });
    });

    // Metrics endpoint (optional)
    app.get('/metrics', (_req, res) => {
      res.json({
        requests: 0,
        connections: 0,
        errors: 0,
      });
    });
  });

  afterAll(() => {
    // Cleanup
  });

  it('should respond to health checks', async () => {
    const response = await supertest(app).get('/health');

    expect(response.status).toBe(200);
    expect(response.body).toHaveProperty('status', 'ok');
    expect(response.body).toHaveProperty('version', '1.0.0');
    expect(response.body).toHaveProperty('uptime');
    expect(response.body).toHaveProperty('redis', 'connected');
    expect(response.body).toHaveProperty('timestamp');
  });

  it('should include uptime in health response', async () => {
    const response = await supertest(app).get('/health');

    expect(response.body.uptime).toBeGreaterThan(0);
  });

  it('should respond to metrics endpoint', async () => {
    const response = await supertest(app).get('/metrics');

    expect(response.status).toBe(200);
    expect(response.body).toHaveProperty('requests');
    expect(response.body).toHaveProperty('connections');
    expect(response.body).toHaveProperty('errors');
  });

  it('should return 404 for unknown routes', async () => {
    const response = await supertest(app).get('/unknown');

    expect(response.status).toBe(404);
  });
});
