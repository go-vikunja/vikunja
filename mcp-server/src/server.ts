import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
  type InitializeResult,
} from '@modelcontextprotocol/sdk/types.js';
import type { Authenticator } from './auth/authenticator.js';
import type { UserContext } from './auth/types.js';
import type { RateLimiter } from './ratelimit/limiter.js';
import type { VikunjaClient } from './vikunja/client.js';
import { ToolRegistry } from './tools/registry.js';
import { logger } from './utils/logger.js';

/**
 * Initialize request parameters
 */
interface InitializeParams {
  protocolVersion: string;
  capabilities: Record<string, unknown>;
  clientInfo: {
    name: string;
    version: string;
  };
}

/**
 * MCP Server for Vikunja
 * Implements the Model Context Protocol for AI agent integration
 */
export class VikunjaMCPServer {
  private readonly server: Server;
  private readonly authenticator: Authenticator;
  private readonly userContexts: Map<string, UserContext>;
  private readonly toolRegistry: ToolRegistry;

  constructor(
    authenticator: Authenticator,
    rateLimiter: RateLimiter,
    vikunjaClient: VikunjaClient
  ) {
    this.authenticator = authenticator;
    this.userContexts = new Map();

    // Initialize tool registry with client and rate limiter
    this.toolRegistry = new ToolRegistry(vikunjaClient, rateLimiter);
    this.toolRegistry.registerAllTools();

    // Initialize MCP server
    this.server = new Server(
      {
        name: 'vikunja-mcp',
        version: '1.0.0',
      },
      {
        capabilities: {
          resources: {},
          tools: {},
          prompts: {},
        },
      }
    );

    this.setupHandlers();
  }

  /**
   * Setup MCP protocol handlers
   */
  private setupHandlers(): void {
    // Register tools/list handler
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      logger.debug('Handling tools/list request');
      const tools = this.toolRegistry.getTools();
      return { tools };
    });

    // Register tools/call handler
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      logger.debug('Handling tools/call request', { tool: request.params.name });
      
      // Get user context
      // Try http-session first (for HTTP/SSE transport), then default (for stdio)
      let connectionId = 'http-session';
      let userContext = this.getUserContext(connectionId);
      
      if (!userContext) {
        connectionId = 'default';
        userContext = this.getUserContext(connectionId);
      }
      
      if (!userContext) {
        logger.error('No user context found for connection', { connectionId });
        throw new Error('Unauthorized: No user context found');
      }

      try {
        const result = await this.toolRegistry.executeTool(
          request.params.name,
          request.params.arguments ?? {},
          userContext
        );
        
        return {
          content: [
            {
              type: 'text',
              text: JSON.stringify(result, null, 2),
            },
          ],
        };
      } catch (error) {
        logger.error('Tool execution failed', { 
          tool: request.params.name, 
          error: error instanceof Error ? error.message : String(error),
        });
        throw error;
      }
    });

    logger.info('MCP server handlers initialized');
  }

  /**
   * Handle MCP initialize request
   */
  async handleInitialize(request: InitializeParams): Promise<InitializeResult> {
    logger.info('Handling initialize request', {
      protocolVersion: request.protocolVersion,
      clientName: request.clientInfo.name,
      clientVersion: request.clientInfo.version,
    });

    return {
      protocolVersion: '2024-11-05',
      serverInfo: {
        name: 'vikunja-mcp',
        version: '1.0.0',
      },
      capabilities: {
        resources: {},
        tools: {},
        prompts: {},
      },
    };
  }

  /**
   * Handle MCP initialized notification
   */
  async handleInitialized(): Promise<void> {
    logger.info('Client initialized notification received');
  }

  /**
   * Authenticate a connection with Vikunja API token
   */
  async authenticateConnection(token: string): Promise<UserContext> {
    logger.info('Authenticating connection');
    const userContext = await this.authenticator.validateToken(token);
    logger.info('Connection authenticated', { userId: userContext.userId });
    return userContext;
  }

  /**
   * Store user context for a connection
   */
  setUserContext(connectionId: string, userContext: UserContext): void {
    this.userContexts.set(connectionId, userContext);
    logger.debug('User context stored', { connectionId, userId: userContext.userId });
  }

  /**
   * Get user context for a connection
   */
  getUserContext(connectionId: string): UserContext | undefined {
    return this.userContexts.get(connectionId);
  }

  /**
   * Remove user context for a connection
   */
  removeUserContext(connectionId: string): void {
    this.userContexts.delete(connectionId);
    logger.debug('User context removed', { connectionId });
  }

  /**
   * Start the MCP server with stdio transport
   */
  async start(): Promise<void> {
    logger.info('Starting MCP server');
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    logger.info('MCP server started and connected via stdio');
  }

  /**
   * Stop the MCP server
   */
  async stop(): Promise<void> {
    logger.info('Stopping MCP server');
    await this.server.close();
    logger.info('MCP server stopped');
  }

  /**
   * Get the underlying MCP Server instance
   */
  getServer(): Server {
    return this.server;
  }
}
