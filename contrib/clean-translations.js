#!/usr/bin/env node

/**
 * Script to remove empty JSON keys from translation files
 * 
 * This script traverses through the specified directories and removes all
 * empty string values from JSON files recursively.
 */

const fs = require('fs');
const path = require('path');

// Get the root directory (where the script is run from)
const rootDir = process.cwd();

// Define directories to process (relative to root)
const directories = [
  path.join(rootDir, 'pkg/i18n/lang'),
  path.join(rootDir, 'frontend/src/i18n/lang')
];

/**
 * Recursively removes empty string values from an object
 * @param {Object} obj - The object to clean
 * @returns {Object} - The cleaned object with empty strings removed
 */
function removeEmptyStrings(obj) {
  if (typeof obj !== 'object' || obj === null) {
    return obj;
  }

  // Handle arrays
  if (Array.isArray(obj)) {
    return obj.map(item => removeEmptyStrings(item))
      .filter(item => item !== '');
  }

  // Handle objects
  const result = {};
  
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      const value = obj[key];
      
      if (value === '') {
        // Skip empty strings
        continue;
      } else if (typeof value === 'object' && value !== null) {
        // Recursively clean nested objects
        const cleanedValue = removeEmptyStrings(value);
        
        // Only add non-empty objects
        if (typeof cleanedValue === 'object' && 
            !Array.isArray(cleanedValue) && 
            Object.keys(cleanedValue).length === 0) {
          continue;
        }
        
        result[key] = cleanedValue;
      } else {
        // Keep non-empty values
        result[key] = value;
      }
    }
  }
  
  return result;
}

/**
 * Process a single JSON file to remove empty strings
 * @param {string} filePath - Path to the JSON file
 */
async function processFile(filePath) {
  try {
    console.log(`Processing ${filePath}`);

    // Read and parse the JSON file
    const data = await fs.promises.readFile(filePath, 'utf8');
    const json = JSON.parse(data);

    // Clean the JSON data
    const cleanedJson = removeEmptyStrings(json);

    // Write the cleaned JSON back to the file
    await fs.promises.writeFile(
      filePath,
      JSON.stringify(cleanedJson, null, '\t'),
      'utf8'
    );

    console.log(`Successfully cleaned ${filePath}`);
  } catch (error) {
    console.error(`Error processing ${filePath}:`, error);
  }
}

/**
 * Process all JSON files in the specified directories
 */
async function main() {
  for (const dir of directories) {
    try {
      await fs.promises.access(dir);
    } catch {
      console.warn(`Directory ${dir} does not exist. Skipping.`);
      continue;
    }

    const files = await fs.promises.readdir(dir);

    for (const file of files) {
      const filePath = path.join(dir, file);

      if (file.endsWith('.json') && file !== 'en.json') {
        await processFile(filePath);
      }
    }
  }

  console.log('All translation files have been processed successfully!');
}

// Run the script
main();
