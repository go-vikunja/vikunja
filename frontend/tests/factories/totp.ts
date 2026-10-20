import {Factory} from '../support/factory'

// Fixed base32 secret so tests can generate deterministic codes with otplib.
export const FIXED_TOTP_SECRET = 'HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ'

export class TotpFactory extends Factory {
	static table = 'totp'

	static factory() {
		return {
			id: '{increment}',
			user_id: 1,
			secret: FIXED_TOTP_SECRET,
			enabled: true,
			url: `otpauth://totp/Vikunja:test?secret=${FIXED_TOTP_SECRET}&issuer=Vikunja`,
		}
	}
}
