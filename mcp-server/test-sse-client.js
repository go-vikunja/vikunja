#!/usr/bin/env node
/**
 * Simple test client for MCP SSE transport
 * Tests the connection and lists available tools
 */

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { SSEClientTransport } from '@modelcontextprotocol/sdk/client/sse.js';

const MCP_URL = process.env.MCP_URL || 'http://192.168.50.64:3010/sse';
const TOKEN = process.env.MCP_TOKEN || 'tk_4afb60db7138c20a7c8e97c17e6619b7c70d8574';

async function testMCPConnection() {
  console.log('🚀 Testing MCP SSE Transport');
  console.log(`📡 Connecting to: ${MCP_URL}`);
  
  try {
    // Create SSE transport
    const transport = new SSEClientTransport(
      new URL(MCP_URL),
      {
        headers: {
          'Authorization': `Bearer ${TOKEN}`,
        },
      }
    );

    // Create MCP client
    const client = new Client(
      {
        name: 'test-client',
        version: '1.0.0',
      },
      {
        capabilities: {},
      }
    );

    // Connect
    console.log('⏳ Establishing SSE connection...');
    await client.connect(transport);
    console.log('✅ Connected successfully!');

    // List available tools
    console.log('\n📋 Listing available tools...');
    const { tools } = await client.listTools();
    console.log(`\n✅ Found ${tools.length} tools:`);
    
    tools.forEach((tool, index) => {
      console.log(`\n${index + 1}. ${tool.name}`);
      console.log(`   ${tool.description}`);
      if (tool.inputSchema?.properties) {
        console.log(`   Parameters: ${Object.keys(tool.inputSchema.properties).join(', ')}`);
      }
    });

    // Close connection
    console.log('\n🔌 Closing connection...');
    await client.close();
    console.log('✅ Test complete!');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
    if (error.stack) {
      console.error('\nStack trace:', error.stack);
    }
    process.exit(1);
  }
}

testMCPConnection();
