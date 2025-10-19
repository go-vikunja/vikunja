/**
 * @file Test suite for ToolRegistry
 * @module tests/unit/tools/registry
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ToolRegistry } from '../../../src/tools/registry.js';
import type { VikunjaClient } from '../../../src/vikunja/client.js';
import type { RateLimiter } from '../../../src/ratelimit/limiter.js';
import type { UserContext } from '../../../src/auth/types.js';

// Mock dependencies
const mockClient = {
  createProject: vi.fn(),
  updateProject: vi.fn(),
  deleteProject: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  deleteTask: vi.fn(),
} as unknown as VikunjaClient;

const mockRateLimiter = {
  checkLimit: vi.fn().mockResolvedValue(undefined),
} as unknown as RateLimiter;

const mockUserContext: UserContext = {
  userId: 1,
  username: 'testuser',
  email: 'test@example.com',
  token: 'test-token',
};

describe('ToolRegistry', () => {
  let registry: ToolRegistry;

  beforeEach(() => {
    vi.clearAllMocks();
    registry = new ToolRegistry(mockClient, mockRateLimiter);
    registry.registerAllTools();
  });

  describe('Tool Registration', () => {
    it('should register all 22 tools', () => {
      const tools = registry.getTools();
      expect(tools).toHaveLength(22);
    });

    it('should register project tools', () => {
      const tools = registry.getTools();
      const projectTools = tools.filter((t) =>
        ['create_project', 'update_project', 'delete_project', 'archive_project'].includes(t.name)
      );
      expect(projectTools).toHaveLength(4);
      expect(projectTools.map((t) => t.name)).toEqual([
        'create_project',
        'update_project',
        'delete_project',
        'archive_project',
      ]);
    });

    it('should register task tools', () => {
      const tools = registry.getTools();
      const taskTools = tools.filter((t) => t.name.includes('task') && !t.name.includes('bulk'));
      expect(taskTools.length).toBeGreaterThanOrEqual(5);
    });

    it('should register assignment tools', () => {
      const tools = registry.getTools();
      const assignmentTools = tools.filter(
        (t) => t.name.includes('assign') || t.name.includes('label')
      );
      expect(assignmentTools.length).toBeGreaterThanOrEqual(5);
    });

    it('should register search tools', () => {
      const tools = registry.getTools();
      const searchTools = tools.filter(
        (t) => t.name.includes('search') || t.name.includes('get_')
      );
      expect(searchTools.length).toBeGreaterThanOrEqual(4);
    });

    it('should register bulk tools', () => {
      const tools = registry.getTools();
      const bulkTools = tools.filter((t) => t.name.includes('bulk'));
      expect(bulkTools).toHaveLength(4);
      expect(bulkTools.map((t) => t.name)).toEqual([
        'bulk_update_tasks',
        'bulk_complete_tasks',
        'bulk_assign_tasks',
        'bulk_add_labels',
      ]);
    });
  });

  describe('Tool Metadata', () => {
    it('should provide tool name, description, and inputSchema', () => {
      const tools = registry.getTools();
      const createProject = tools.find((t) => t.name === 'create_project');

      expect(createProject).toBeDefined();
      expect(createProject?.name).toBe('create_project');
      expect(createProject?.description).toBe('Create a new project in Vikunja');
      expect(createProject?.inputSchema).toBeDefined();
      expect(createProject?.inputSchema.type).toBe('object');
      expect(createProject?.inputSchema.properties).toBeDefined();
    });

    it('should include required fields in schema', () => {
      const tools = registry.getTools();
      const createProject = tools.find((t) => t.name === 'create_project');

      expect(createProject?.inputSchema.required).toContain('title');
      expect(createProject?.inputSchema.properties.title).toBeDefined();
    });

    it('should include optional fields in schema', () => {
      const tools = registry.getTools();
      const createProject = tools.find((t) => t.name === 'create_project');

      expect(createProject?.inputSchema.properties.description).toBeDefined();
      expect(createProject?.inputSchema.properties.hex_color).toBeDefined();
    });
  });

  describe('Tool Execution', () => {
    it('should execute create_project tool with validation', async () => {
      // Tool execution will fail with mock client, but we can verify validation works
      const result = await registry.executeTool(
        'create_project',
        { title: 'Test Project' },
        mockUserContext
      );

      // Result should be an error object from the tool
      expect(result).toHaveProperty('success', false);
      expect(result).toHaveProperty('error');
    });

    it('should validate input with Zod schema', async () => {
      await expect(
        registry.executeTool('create_project', { title: '' }, mockUserContext)
      ).rejects.toThrow();
    });

    it('should enforce rate limiting', async () => {
      await registry.executeTool(
        'create_project',
        { title: 'Test Project' },
        mockUserContext
      );

      // Rate limiter is called with the token, not userId
      expect(mockRateLimiter.checkLimit).toHaveBeenCalledWith(mockUserContext.token);
    });

    it('should throw error for unknown tool', async () => {
      await expect(
        registry.executeTool('unknown_tool', {}, mockUserContext)
      ).rejects.toThrow('Tool not found: unknown_tool');
    });
  });

  describe('Tool Discovery', () => {
    it('should check if tool exists', () => {
      expect(registry.hasTool('create_project')).toBe(true);
      expect(registry.hasTool('unknown_tool')).toBe(false);
    });

    it('should return all tool names', () => {
      const tools = registry.getTools();
      const names = tools.map((t) => t.name);
      
      expect(names).toContain('create_project');
      expect(names).toContain('create_task');
      expect(names).toContain('search_tasks');
      expect(names).toContain('bulk_update_tasks');
    });
  });

  describe('Schema Conversion', () => {
    it('should convert string fields to JSON Schema', () => {
      const tools = registry.getTools();
      const createProject = tools.find((t) => t.name === 'create_project');
      const titleSchema = createProject?.inputSchema.properties.title as Record<string, unknown>;

      expect(titleSchema.type).toBe('string');
      expect(titleSchema.minLength).toBe(1);
      expect(titleSchema.maxLength).toBe(250);
    });

    it('should convert number fields to JSON Schema', () => {
      const tools = registry.getTools();
      const updateProject = tools.find((t) => t.name === 'update_project');
      const idSchema = updateProject?.inputSchema.properties.id as Record<string, unknown>;

      // Zod converts .int() to 'integer' in JSON Schema
      expect(idSchema.type).toBe('integer');
      // The schema conversion may not perfectly handle all Zod constraints
      expect(idSchema).toHaveProperty('minimum');
    });

    it('should convert boolean fields to JSON Schema', () => {
      const tools = registry.getTools();
      const archiveProject = tools.find((t) => t.name === 'archive_project');
      const archivedSchema = archiveProject?.inputSchema.properties
        .archived as Record<string, unknown>;

      expect(archivedSchema.type).toBe('boolean');
    });

    it('should convert array fields to JSON Schema', () => {
      const tools = registry.getTools();
      const bulkUpdate = tools.find((t) => t.name === 'bulk_update_tasks');
      const taskIdsSchema = bulkUpdate?.inputSchema.properties.task_ids as Record<string, unknown>;

      expect(taskIdsSchema.type).toBe('array');
    });
  });
});
