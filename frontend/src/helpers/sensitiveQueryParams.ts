// One-time tokens from email links: a leaked password reset token is an account takeover.
export const SENSITIVE_QUERY_PARAMS = ['userPasswordReset', 'userEmailConfirm', 'accountDeletionConfirm']
