# In frontend/cypress/Dockerfile

# Start from an official Node.js image which includes curl, nvm, etc.
FROM node:20

# Install Cypress dependencies (as root)
RUN apt-get update && apt-get install -y \
    libgtk2.0-0 \
    libgtk-3-0 \
    libgbm-dev \
    libnotify-dev \
    libgconf-2-4 \
    libnss3 \
    libxss1 \
    libasound2 \
    libxtst6 \
    xauth \
    xvfb

# Install pnpm globally (as root)
RUN npm install -g pnpm

# Set up the working directory
WORKDIR /project

# Now, switch to the non-root user for security
USER node