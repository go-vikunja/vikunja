import 'dotenv/config';
import express, { type Express } from 'express';
import { config } from './config/index.js';
import { logger } from './utils/logger.js';
import { RedisStorage } from './ratelimit/storage.js';
import { RedisConnectionManager } from './utils/redis-connection.js';
import { Authenticator } from './auth/authenticator.js';
import { RateLimiter } from './ratelimit/limiter.js';
import { VikunjaClient } from './vikunja/client.js';
import { VikunjaMCPServer } from './server.js';
import { HTTPStreamableTransport } from './transports/http/http-streamable.js';
import { SSETransport } from './transports/http/sse-transport.js';
import { TokenValidator } from './auth/token-validator.js';
import { SessionManager } from './transports/http/session-manager.js';
import { timeoutMiddleware, haltOnTimeout } from './utils/timeout-middleware.js';

/**
 * Application state
 */
interface AppState {
  redis: RedisStorage;
  vikunjaClient: VikunjaClient;
  authenticator: Authenticator;
  rateLimiter: RateLimiter;
  mcpServer: VikunjaMCPServer;
  httpServer: Express;
  tokenValidator: TokenValidator;
  sessionManager: SessionManager;
  httpStreamableTransport?: HTTPStreamableTransport;
  sseTransport?: SSETransport;
  startTime: number;
}

let appState: AppState | null = null;

/**
 * Initialize all components
 */
async function initializeApp(): Promise<AppState> {
  logger.info('Initializing Vikunja MCP Server');

  // Initialize shared Redis connection
  logger.info('Initializing shared Redis connection', {
    host: config.redis.host,
    port: config.redis.port,
  });
  await RedisConnectionManager.getConnection();

  // Create Redis storage (uses shared connection)
  const redis = new RedisStorage();
  await redis.connect();

  // Create Vikunja API client
  logger.info('Creating Vikunja API client', {
    apiUrl: config.vikunjaApiUrl,
  });
  const vikunjaClient = new VikunjaClient();

  // Create authenticator
  const authenticator = new Authenticator();

  // Create token validator (for HTTP transport)
  const tokenValidator = new TokenValidator();

  // Create session manager (for HTTP transport)
  const sessionManager = new SessionManager();

  // Create rate limiter
  const rateLimiter = new RateLimiter(redis);

  // Create MCP server
  const mcpServer = new VikunjaMCPServer(authenticator, rateLimiter, vikunjaClient);

  // Create HTTP server for health checks
  const httpServer = express();
  httpServer.use(express.json());

  // Health check endpoint
  httpServer.get('/health', async (_req, res) => {
    const redisHealthy = await redis.isHealthy();
    const uptime = process.uptime();
    const status = redisHealthy ? 'ok' : 'degraded';

    res.status(redisHealthy ? 200 : 503).json({
      status,
      version: '1.0.0',
      uptime,
      redis: redisHealthy ? 'connected' : 'disconnected',
      timestamp: new Date().toISOString(),
    });
  });

  // Metrics endpoint (optional)
  httpServer.get('/metrics', (_req, res) => {
    res.json({
      requests: 0, // TODO: implement metrics tracking
      connections: 0, // TODO: implement connection tracking
      errors: 0, // TODO: implement error tracking
      uptime: process.uptime(),
    });
  });

  // Info endpoint
  httpServer.get('/info', (_req, res) => {
    res.json({
      name: 'vikunja-mcp',
      version: '1.0.0',
      protocol: 'MCP 2024-11-05',
      vikunjaApiUrl: config.vikunjaApiUrl,
    });
  });

  // HTTP Streamable transport endpoint (POST /mcp) - Primary recommended transport
  let httpStreamableTransport: HTTPStreamableTransport | undefined;
  let sseTransport: SSETransport | undefined;
  
  if (config.httpTransport.enabled) {
    // Initialize HTTP Streamable transport (recommended)
    httpStreamableTransport = new HTTPStreamableTransport({
      mcpServer,
      sessionManager,
      tokenValidator,
      rateLimiter,
      enableJsonResponse: config.httpTransport.enableJsonResponse,
    });

    // Add timeout middleware to MCP endpoint (30 seconds)
    const timeout30s = timeoutMiddleware({ timeoutMs: 30000 });

    // Capture transport in const to avoid non-null assertion
    const streamableTransport = httpStreamableTransport;
    httpServer.post('/mcp', timeout30s, haltOnTimeout, (req, res) => {
      void streamableTransport.handleRequest(req, res);
    });

    logger.info('HTTP Streamable transport endpoint registered at POST /mcp (30s timeout)', {
      enableJsonResponse: config.httpTransport.enableJsonResponse,
    });

    // Initialize SSE transport (deprecated, for backward compatibility)
    sseTransport = new SSETransport({
      mcpServer,
      sessionManager,
      tokenValidator,
      rateLimiter,
      postEndpoint: '/sse',
    });

    // Capture transport in const to avoid non-null assertion
    const sse = sseTransport;
    
    // GET /sse - Client opens SSE connection to receive messages (longer timeout for streaming)
    httpServer.get('/sse', timeoutMiddleware({ timeoutMs: 300000 }), haltOnTimeout, (req, res) => {
      void sse.handleStream(req, res);
    });
    
    // POST /sse - Client sends messages to server (30 second timeout)
    httpServer.post('/sse', timeout30s, haltOnTimeout, (req, res) => {
      void sse.handleMessage(req, res);
    });

    logger.info('SSE transport endpoints registered at GET/POST /sse (DEPRECATED)');
  }

  // Start HTTP server if HTTP transport is enabled
  if (config.httpTransport.enabled) {
    const port = config.httpTransport.port;
    await new Promise<void>((resolve) => {
      httpServer.listen(port, config.httpTransport.host, () => {
        logger.info(`HTTP server listening on ${config.httpTransport.host}:${port}`);
        logger.info(`SSE endpoint available at GET/POST http://${config.httpTransport.host}:${port}/sse`);
        logger.info('HTTP transport mode enabled - stdio transport disabled');
        resolve();
      });
    });
  } else {
    // Start HTTP server on legacy port for health checks only (no SSE endpoints)
    const port = config.port;
    await new Promise<void>((resolve) => {
      httpServer.listen(port, () => {
        logger.info(`HTTP server listening on port ${port} (health checks only)`);
        resolve();
      });
    });
  }

  // Start MCP server (stdio transport) only if HTTP transport is disabled
  if (!config.httpTransport.enabled) {
    logger.info('Starting MCP server in stdio mode');
    await mcpServer.start();
  } else {
    logger.info('Stdio transport disabled (HTTP transport mode)');
  }

  return {
    redis,
    vikunjaClient,
    authenticator,
    rateLimiter,
    mcpServer,
    httpServer,
    tokenValidator,
    sessionManager,
    ...(httpStreamableTransport ? { httpStreamableTransport } : {}),
    ...(sseTransport ? { sseTransport } : {}),
    startTime: Date.now(),
  };
}

/**
 * Graceful shutdown
 */
async function shutdown(): Promise<void> {
  logger.info('Shutting down Vikunja MCP Server');

  if (!appState) {
    return;
  }

  try {
    // Stop HTTP Streamable transport
    if (appState.httpStreamableTransport) {
      await appState.httpStreamableTransport.close();
    }

    // Stop SSE transport
    if (appState.sseTransport) {
      await appState.sseTransport.closeAll();
    }

    // Stop session cleanup
    await appState.sessionManager.shutdown();

    // Stop MCP server
    await appState.mcpServer.stop();

    // Disconnect from Redis storage
    await appState.redis.disconnect();

    // Disconnect shared Redis connection
    await RedisConnectionManager.disconnect();

    logger.info('Shutdown complete');
  } catch (error) {
    logger.error('Error during shutdown', { error });
  }
}

/**
 * Error handler
 */
function handleError(error: Error): void {
  logger.error('Unhandled error', { error });
  process.exit(1);
}

// Handle shutdown signals
process.on('SIGTERM', async () => {
  await shutdown();
  process.exit(0);
});

process.on('SIGINT', async () => {
  await shutdown();
  process.exit(0);
});

// Handle uncaught errors
process.on('uncaughtException', handleError);
process.on('unhandledRejection', handleError);

// Start the application
initializeApp()
  .then((state) => {
    appState = state;
    const mode = config.httpTransport.enabled ? 'HTTP' : 'stdio';
    logger.info('Vikunja MCP Server started successfully', {
      version: '1.0.0',
      mode,
      port: config.httpTransport.enabled ? config.httpTransport.port : config.port,
    });
  })
  .catch((error: Error) => {
    logger.error('Failed to start server', { error });
    process.exit(1);
  });
