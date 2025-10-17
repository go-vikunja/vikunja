# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY src ./src

# Build TypeScript
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Install production dependencies only
COPY package*.json ./
RUN npm ci --only=production

# Copy built files from builder
COPY --from=builder /app/dist ./dist

# Create non-root user
RUN addgroup -g 1001 -S vikunja && \
    adduser -u 1001 -S vikunja -G vikunja

# Switch to non-root user
USER vikunja

# Expose MCP server port
EXPOSE 3457

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3457/health', (r) => { process.exit(r.statusCode === 200 ? 0 : 1); })"

# Start the server
CMD ["node", "dist/index.js"]
