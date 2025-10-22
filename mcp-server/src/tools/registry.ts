import { z } from 'zod';
import { ProjectTools, CreateProjectSchema, UpdateProjectSchema, DeleteProjectSchema, ArchiveProjectSchema } from './projects.js';
import { TaskTools, CreateTaskSchema, UpdateTaskSchema, CompleteTaskSchema, DeleteTaskSchema, MoveTaskSchema } from './tasks.js';
import { AssignmentTools, AssignTaskSchema, UnassignTaskSchema, AddLabelSchema, RemoveLabelSchema, CreateLabelSchema } from './assignments.js';
import { SearchTools, SearchTasksSchema, SearchProjectsSchema, GetMyTasksSchema, GetProjectTasksSchema } from './search.js';
import { BulkTools, BulkUpdateTasksSchema, BulkCompleteTasksSchema, BulkAssignTasksSchema, BulkAddLabelsSchema } from './bulk.js';
import { VikunjaClient } from '../vikunja/client.js';
import { RateLimiter } from '../ratelimit/limiter.js';
import { UserContext } from '../auth/types.js';

/**
 * MCP Tool definition
 */
export interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, unknown>;
    required?: string[];
  };
}

/**
 * Tool execution function type
 */
export type ToolExecutor = (
  args: Record<string, unknown>,
  userContext: UserContext
) => Promise<unknown>;

/**
 * Tool registry for MCP server
 */
export class ToolRegistry {
  private readonly projectTools: ProjectTools;
  private readonly taskTools: TaskTools;
  private readonly assignmentTools: AssignmentTools;
  private readonly searchTools: SearchTools;
  private readonly bulkTools: BulkTools;

  private readonly tools: Map<string, MCPTool>;
  private readonly executors: Map<string, ToolExecutor>;

  constructor(client: VikunjaClient, rateLimiter: RateLimiter) {
    this.projectTools = new ProjectTools(client, rateLimiter);
    this.taskTools = new TaskTools(client, rateLimiter);
    this.assignmentTools = new AssignmentTools(client, rateLimiter);
    this.searchTools = new SearchTools(client, rateLimiter);
    this.bulkTools = new BulkTools(client, rateLimiter);

    this.tools = new Map();
    this.executors = new Map();

    this.registerAllTools();
  }

  /**
   * Register all tools with their schemas and executors
   */
  public registerAllTools(): void {
    // Project Tools
    this.registerTool(
      'create_project',
      'Create a new project in Vikunja',
      CreateProjectSchema,
      async (args, ctx) => this.projectTools.createProject(args as z.infer<typeof CreateProjectSchema>, ctx)
    );

    this.registerTool(
      'update_project',
      'Update an existing project',
      UpdateProjectSchema,
      async (args, ctx) => this.projectTools.updateProject(args as z.infer<typeof UpdateProjectSchema>, ctx)
    );

    this.registerTool(
      'delete_project',
      'Delete a project',
      DeleteProjectSchema,
      async (args, ctx) => this.projectTools.deleteProject(args as z.infer<typeof DeleteProjectSchema>, ctx)
    );

    this.registerTool(
      'archive_project',
      'Archive or unarchive a project',
      ArchiveProjectSchema,
      async (args, ctx) => this.projectTools.archiveProject(args as z.infer<typeof ArchiveProjectSchema>, ctx)
    );

    // Task Tools
    this.registerTool(
      'create_task',
      'Create a new task in a project',
      CreateTaskSchema,
      async (args, ctx) => this.taskTools.createTask(args as z.infer<typeof CreateTaskSchema>, ctx)
    );

    this.registerTool(
      'update_task',
      'Update an existing task',
      UpdateTaskSchema,
      async (args, ctx) => this.taskTools.updateTask(args as z.infer<typeof UpdateTaskSchema>, ctx)
    );

    this.registerTool(
      'complete_task',
      'Mark a task as complete',
      CompleteTaskSchema,
      async (args, ctx) => this.taskTools.completeTask(args as z.infer<typeof CompleteTaskSchema>, ctx)
    );

    this.registerTool(
      'delete_task',
      'Delete a task',
      DeleteTaskSchema,
      async (args, ctx) => this.taskTools.deleteTask(args as z.infer<typeof DeleteTaskSchema>, ctx)
    );

    this.registerTool(
      'move_task',
      'Move a task to a different project',
      MoveTaskSchema,
      async (args, ctx) => this.taskTools.moveTask(args as z.infer<typeof MoveTaskSchema>, ctx)
    );

    // Assignment Tools
    this.registerTool(
      'assign_task',
      'Assign a user to a task',
      AssignTaskSchema,
      async (args, ctx) => this.assignmentTools.assignTask(args as z.infer<typeof AssignTaskSchema>, ctx)
    );

    this.registerTool(
      'unassign_task',
      'Remove a user from a task',
      UnassignTaskSchema,
      async (args, ctx) => this.assignmentTools.unassignTask(args as z.infer<typeof UnassignTaskSchema>, ctx)
    );

    this.registerTool(
      'add_label',
      'Add a label to a task',
      AddLabelSchema,
      async (args, ctx) => this.assignmentTools.addLabel(args as z.infer<typeof AddLabelSchema>, ctx)
    );

    this.registerTool(
      'remove_label',
      'Remove a label from a task',
      RemoveLabelSchema,
      async (args, ctx) => this.assignmentTools.removeLabel(args as z.infer<typeof RemoveLabelSchema>, ctx)
    );

    this.registerTool(
      'create_label',
      'Create a new label',
      CreateLabelSchema,
      async (args, ctx) => this.assignmentTools.createLabel(args as z.infer<typeof CreateLabelSchema>, ctx)
    );

    // Search Tools
    this.registerTool(
      'search_tasks',
      'Search for tasks by query string with advanced filtering',
      SearchTasksSchema,
      async (args, ctx) => this.searchTools.searchTasks(args as z.infer<typeof SearchTasksSchema>, ctx)
    );

    this.registerTool(
      'search_projects',
      'Search for projects by query string',
      SearchProjectsSchema,
      async (args, ctx) => this.searchTools.searchProjects(args as z.infer<typeof SearchProjectsSchema>, ctx)
    );

    this.registerTool(
      'get_my_tasks',
      'Get all tasks assigned to the current user',
      GetMyTasksSchema,
      async (args, ctx) => this.searchTools.getMyTasks(args as z.infer<typeof GetMyTasksSchema>, ctx)
    );

    this.registerTool(
      'get_project_tasks',
      'Get all tasks in a specific project',
      GetProjectTasksSchema,
      async (args, ctx) => this.searchTools.getProjectTasks(args as z.infer<typeof GetProjectTasksSchema>, ctx)
    );

    // Bulk Tools
    this.registerTool(
      'bulk_update_tasks',
      'Update multiple tasks at once (max 100 tasks)',
      BulkUpdateTasksSchema,
      async (args, ctx) => this.bulkTools.bulkUpdateTasks(args as z.infer<typeof BulkUpdateTasksSchema>, ctx)
    );

    this.registerTool(
      'bulk_complete_tasks',
      'Mark multiple tasks as complete (max 100 tasks)',
      BulkCompleteTasksSchema,
      async (args, ctx) => this.bulkTools.bulkCompleteTasks(args as z.infer<typeof BulkCompleteTasksSchema>, ctx)
    );

    this.registerTool(
      'bulk_assign_tasks',
      'Assign a user to multiple tasks (max 100 tasks)',
      BulkAssignTasksSchema,
      async (args, ctx) => this.bulkTools.bulkAssignTasks(args as z.infer<typeof BulkAssignTasksSchema>, ctx)
    );

    this.registerTool(
      'bulk_add_labels',
      'Add a label to multiple tasks (max 100 tasks)',
      BulkAddLabelsSchema,
      async (args, ctx) => this.bulkTools.bulkAddLabels(args as z.infer<typeof BulkAddLabelsSchema>, ctx)
    );
  }

  /**
   * Register a single tool with its schema and executor
   */
  private registerTool(
    name: string,
    description: string,
    schema: z.ZodType,
    executor: ToolExecutor
  ): void {
    // Convert Zod schema to JSON Schema for MCP
    const inputSchema = this.zodToJsonSchema(schema);

    this.tools.set(name, {
      name,
      description,
      inputSchema,
    });

    // Wrap executor with validation
    this.executors.set(name, async (args, ctx) => {
      // Validate args with Zod schema
      const validatedArgs = schema.parse(args);
      return executor(validatedArgs as Record<string, unknown>, ctx);
    });
  }

  /**
   * Convert Zod schema to JSON Schema (simplified version)
   */
  private zodToJsonSchema(schema: z.ZodType): {
    type: 'object';
    properties: Record<string, unknown>;
    required?: string[];
  } {
    // Get the schema shape if it's a ZodObject
    if (schema instanceof z.ZodObject) {
      const shape = schema.shape;
      const properties: Record<string, unknown> = {};
      const required: string[] = [];

      for (const [key, value] of Object.entries(shape)) {
        const zodType = value as z.ZodType;
        properties[key] = this.zodTypeToJsonSchema(zodType);

        // Check if field is required (not optional or nullable)
        if (!zodType.isOptional() && !zodType.isNullable()) {
          required.push(key);
        }
      }

      return {
        type: 'object',
        properties,
        ...(required.length > 0 && { required }),
      };
    }

    // Fallback for non-object schemas
    return {
      type: 'object',
      properties: {},
    };
  }

  /**
   * Convert a Zod type to JSON Schema type
   */
  private zodTypeToJsonSchema(zodType: z.ZodType): Record<string, unknown> {
    // Unwrap optional and nullable
    let type = zodType;
    const isOptional = type.isOptional();
    
    if (isOptional) {
      type = (type as z.ZodOptional<z.ZodType>)._def.innerType;
    }

    // Handle different Zod types
    if (type instanceof z.ZodString) {
      const stringType: Record<string, unknown> = { type: 'string' };
      
      // Check for regex validation (for hex colors, etc.)
      const checks = (type)._def.checks;
      if (checks) {
        for (const check of checks) {
          if (check.kind === 'regex') {
            stringType['pattern'] = check.regex.source;
          } else if (check.kind === 'min') {
            stringType['minLength'] = check.value;
          } else if (check.kind === 'max') {
            stringType['maxLength'] = check.value;
          }
        }
      }
      
      return stringType;
    }

    if (type instanceof z.ZodNumber) {
      const numberType: Record<string, unknown> = { type: 'number' };
      
      const checks = (type)._def.checks;
      if (checks) {
        for (const check of checks) {
          if (check.kind === 'min') {
            numberType['minimum'] = check.value;
          } else if (check.kind === 'max') {
            numberType['maximum'] = check.value;
          } else if (check.kind === 'int') {
            numberType['type'] = 'integer';
          }
        }
      }
      
      return numberType;
    }

    if (type instanceof z.ZodBoolean) {
      return { type: 'boolean' };
    }

    if (type instanceof z.ZodArray) {
      const itemType = (type as z.ZodArray<z.ZodType>)._def.type;
      return {
        type: 'array',
        items: this.zodTypeToJsonSchema(itemType),
      };
    }

    if (type instanceof z.ZodObject) {
      return this.zodToJsonSchema(type);
    }

    if (type instanceof z.ZodLiteral) {
      return {
        type: typeof (type as z.ZodLiteral<unknown>)._def.value,
        const: (type as z.ZodLiteral<unknown>)._def.value,
      };
    }

    // Default fallback
    return { type: 'string' };
  }

  /**
   * Get all registered tools
   */
  getTools(): MCPTool[] {
    return Array.from(this.tools.values());
  }

  /**
   * Execute a tool by name
   */
  async executeTool(
    name: string,
    args: Record<string, unknown>,
    userContext: UserContext
  ): Promise<unknown> {
    const executor = this.executors.get(name);
    if (!executor) {
      throw new Error(`Tool not found: ${name}`);
    }

    return executor(args, userContext);
  }

  /**
   * Check if a tool exists
   */
  hasTool(name: string): boolean {
    return this.tools.has(name);
  }
}
