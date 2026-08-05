/**
 * Hash-fragment prefix used to carry a post-login destination in the URL.
 *
 * Unlike the localStorage redirect, this lives in the address bar so the URL
 * stays copyable between browsers (needed for native OAuth clients that open
 * /oauth/authorize, see #2654). It uses the hash – not a query param – so the
 * embedded OAuth parameters never reach server or proxy access logs.
 *
 * Must stay distinct from LINK_SHARE_HASH_PREFIX, which router.beforeEach
 * special-cases.
 */
export const REDIRECT_HASH_PREFIX = '#redirect='
