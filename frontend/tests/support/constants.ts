/**
 * Shared test constants
 */

/**
 * Default password used for test users.
 * The bcrypt hash for this password is used in the user factory.
 */
export const TEST_PASSWORD = '1234'

/**
 * Bcrypt hash of TEST_PASSWORD ('1234') for database seeding
 */
export const TEST_PASSWORD_HASH = '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.'

/**
 * Default API URL used in test environments.
 * Override with the API_URL environment variable when running against a different host or base path.
 */
export const TEST_API_URL = process.env.API_URL || 'http://127.0.0.1:3456/api/v1'
