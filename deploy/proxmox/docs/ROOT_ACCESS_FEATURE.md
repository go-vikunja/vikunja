# Root Access Configuration Feature

**Added**: October 20, 2025  
**Status**: Complete  
**Version**: 1.0.0

## Overview

This feature adds professional root access configuration to Vikunja's Proxmox LXC deployment system. Administrators can configure password-based authentication, SSH key-based authentication, or both, with comprehensive security hardening and validation.

## Feature Highlights

### 🔐 Security Features

- **Cryptographically Secure Password Generation**: Uses OpenSSL to generate 32-character random passwords
- **SSH Key Validation**: Validates public key format before injection (supports RSA, Ed25519, ECDSA, FIDO/U2F)
- **Proper Permission Management**: Automatically sets correct ownership and permissions (700 for .ssh, 600 for authorized_keys)
- **SSH Daemon Hardening**: Disables password auth when keys are used, disables empty passwords, disables X11 forwarding
- **Console Access Preservation**: Root password always set for emergency console access via `pct enter`

### 🎯 Flexible Authentication Modes

1. **Password Only** - Simple password authentication (convenient for testing)
2. **SSH Key Only** - Key-based authentication (recommended for production) ⭐
3. **Both Password and SSH Key** - Flexible access with moderate security
4. **Auto-generated Password** - Random password generated, access via `pct enter`

### 🛠️ Implementation

The feature includes:

- **Interactive Mode**: User-friendly prompts with 4 authentication method choices
- **Non-Interactive Mode**: CLI options for automated deployments
- **Deployment Summary**: Displays SSH connection command and credentials
- **Comprehensive Documentation**: Security best practices and troubleshooting

## Usage Examples

### Interactive Installation (Recommended)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/004-proxmox-deployment/deploy/proxmox/vikunja-install.sh)
```

During installation, you'll be prompted:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Root Access Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Configure root access to the LXC container:
  1) Password only (less secure, convenient for testing)
  2) SSH key only (recommended for production)
  3) Both password and SSH key (flexible, moderate security)
  4) Auto-generated password, no SSH (can access via 'pct enter')

Select root access method [2]:
```

### Non-Interactive Mode

#### SSH Key Authentication (Production Recommended)

```bash
./vikunja-install-main.sh --non-interactive \
  --root-ssh-key ~/.ssh/id_ed25519.pub \
  --disable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1
```

#### Password Authentication (Testing)

```bash
./vikunja-install-main.sh --non-interactive \
  --root-password "SecurePassword123!" \
  --enable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1
```

#### Both Password and SSH Key

```bash
./vikunja-install-main.sh --non-interactive \
  --root-password "SecurePassword123!" \
  --root-ssh-key ~/.ssh/id_rsa.pub \
  --enable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1
```

## CLI Options Reference

| Option | Description | Default |
|--------|-------------|---------|
| `--root-password PASS` | Set root password | Auto-generated secure random |
| `--root-ssh-key FILE` | Path to SSH public key file | None |
| `--enable-root-password` | Enable SSH password authentication | Auto (enabled if no key) |
| `--disable-root-password` | Disable SSH password authentication | Auto (disabled if key provided) |

## Deployment Summary Output

After successful deployment, the summary displays root access information:

```
╔════════════════════════════════════════════════════════════╗
║               Container Root Access                        ║
╚════════════════════════════════════════════════════════════╝

SSH Access:        ssh root@192.168.1.100
Authentication:    SSH key only (password auth disabled)

Console Access:    pct enter 100
```

If a password was generated or set:

```
⚠️  IMPORTANT - Save this root password securely:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Root Password:     Xy9K3mNp8QrT2vWz5LhJ6BcD4FgA7sE1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

This password is for console/emergency access.
SSH password authentication is DISABLED (key-only access).
```

## Security Best Practices

### For Production Deployments

✅ **Use SSH key authentication only**
```bash
--root-ssh-key ~/.ssh/id_ed25519.pub --disable-root-password
```

✅ **Use Ed25519 keys** (strongest, fastest)
```bash
ssh-keygen -t ed25519 -C "vikunja-container-access"
```

✅ **Protect SSH private keys** with strong passphrases
```bash
ssh-keygen -t ed25519 -C "vikunja" -N "your-strong-passphrase"
```

✅ **Rotate keys regularly** and audit authorized_keys

✅ **Use SSH certificate authorities** for infrastructure at scale

### For Development/Testing

⚠️ **Password authentication acceptable** for local testing environments

⚠️ **Auto-generated passwords are secure** but must be saved in password manager

⚠️ **Console access via `pct enter`** always available regardless of SSH configuration

## Technical Implementation

### Files Modified

1. **`lib/lxc-setup.sh`** - Enhanced `setup_ssh_access()` function with:
   - SSH public key injection and validation
   - Cryptographically secure password generation
   - SSH daemon configuration and hardening
   - Comprehensive error handling and logging

2. **`vikunja-install-main.sh`** - Added:
   - Global configuration variables for root access
   - CLI argument parsing for root access options
   - Interactive prompts with 4 authentication modes
   - Integration into deployment orchestration (step 3.5)
   - Enhanced deployment summary with root access info

3. **`templates/deployment-config.yaml`** - Added:
   - `root_access` configuration section
   - Tracking for ssh_enabled, password_auth, key_auth

4. **`README.md`** - Added:
   - Comprehensive "Root Access Configuration" section
   - Security best practices documentation
   - CLI options reference
   - Troubleshooting guide for root access issues

5. **`specs/004-proxmox-deployment/tasks.md`** - Documented:
   - Phase 3.95 feature implementation tasks
   - Security features checklist
   - Testing validation criteria

### Code Quality

- ✅ **Professional error handling** with detailed troubleshooting messages
- ✅ **Input validation** for all parameters
- ✅ **Secure defaults** (key-only auth recommended)
- ✅ **Comprehensive logging** (debug, info, warning, error, success levels)
- ✅ **Idempotent operations** (safe to re-run)
- ✅ **ShellCheck compliant** (no linting errors)
- ✅ **Follows project conventions** (matches existing code style)

## Supported SSH Key Types

The implementation validates and supports all modern SSH key types:

- ✅ **RSA keys** (minimum 2048 bits): `ssh-rsa ...`
- ✅ **Ed25519 keys** (recommended): `ssh-ed25519 ...`
- ✅ **ECDSA keys**: `ecdsa-sha2-nistp256 ...`, `ecdsa-sha2-nistp384 ...`, `ecdsa-sha2-nistp521 ...`
- ✅ **FIDO/U2F keys**: `sk-ssh-ed25519@openssh.com ...`, `sk-ecdsa-sha2-nistp256@openssh.com ...`

## Access Methods After Installation

### SSH Access (if configured)
```bash
ssh root@<container-ip>
```

### Console Access (always available)
```bash
pct enter <container-id>
```

### From Proxmox Web UI
Navigate to container → Console

## Troubleshooting

### Cannot SSH to Container

**Solutions**:
1. Verify SSH service is running: `pct exec <id> systemctl status sshd`
2. Check firewall rules don't block port 22
3. Verify SSH key was injected correctly: `pct exec <id> cat /root/.ssh/authorized_keys`
4. Check SSH daemon configuration: `pct exec <id> cat /etc/ssh/sshd_config | grep -E '(PermitRootLogin|PasswordAuthentication)'`
5. Use console access as fallback: `pct enter <id>`

### Lost Root Password

**Solutions**:
1. Use console access: `pct enter <container-id>` (always works)
2. Reset password from host: `pct exec <id> bash -c "echo 'root:newpassword' | chpasswd"`
3. Inject new SSH key from host (see vikunja-manage.sh reconfigure - future feature)

### SSH Key Not Working

**Solutions**:
1. Verify key format: `ssh-keygen -l -f ~/.ssh/id_ed25519.pub`
2. Check authorized_keys permissions: `pct exec <id> ls -la /root/.ssh/`
3. Test with verbose output: `ssh -v root@<container-ip>`
4. Verify SSH daemon allows pubkey auth: `pct exec <id> grep PubkeyAuthentication /etc/ssh/sshd_config`

## Testing Validation

### Manual Testing Checklist

- [X] Password-only authentication works
- [X] SSH key-only authentication works
- [X] Both password and key authentication works
- [X] Auto-generated password displayed in summary
- [X] Console access via `pct enter` always works
- [X] SSH daemon properly hardened when key is used
- [X] Invalid SSH key files properly rejected
- [X] Empty SSH key files properly rejected
- [X] Missing SSH key files properly handled
- [X] Password mismatch during confirmation properly handled
- [X] Deployment summary shows correct access information

### Security Testing Checklist

- [X] Auto-generated passwords are cryptographically secure (OpenSSL)
- [X] SSH key validation rejects invalid formats
- [X] Authorized_keys has correct permissions (600)
- [X] .ssh directory has correct permissions (700)
- [X] Password authentication disabled when using key-only mode
- [X] Empty passwords prevented by SSH daemon config
- [X] X11 forwarding disabled by default
- [X] Root login permitted (required for container access)

## Future Enhancements

Potential improvements for future versions:

- [ ] Support for multiple SSH keys
- [ ] Integration with vikunja-manage.sh reconfigure for post-deployment key updates
- [ ] SSH key rotation automation
- [ ] Integration with SSH certificate authorities
- [ ] Support for custom SSH daemon ports
- [ ] 2FA/MFA support via PAM modules
- [ ] Audit logging for SSH access attempts

## Version History

- **1.0.0** (October 20, 2025) - Initial implementation
  - Interactive mode with 4 authentication options
  - Non-interactive mode with CLI flags
  - SSH key injection and validation
  - Cryptographically secure password generation
  - SSH daemon hardening
  - Comprehensive documentation

## References

- [OpenSSH Documentation](https://www.openssh.com/manual.html)
- [Proxmox LXC Documentation](https://pve.proxmox.com/wiki/Linux_Container)
- [SSH Best Practices (Mozilla)](https://infosec.mozilla.org/guidelines/openssh)
- [NIST Guidelines for SSH](https://nvlpubs.nist.gov/nistpubs/ir/2015/NIST.IR.7966.pdf)

## License

This feature is part of the Vikunja project and is licensed under AGPLv3.
