#!/usr/bin/env bash
#
# install-gospeed.sh
# Proxmox LXC installer for GoSpeed (community-style)
#
# Usage (simple):
#   bash -c "$(wget -qLO - https://raw.githubusercontent.com/<you>/gospeed/main/install-gospeed.sh)"
#
# Override defaults via environment variables:
#   LXC_ID=123 LXC_NAME=gospeed LXC_DISK=2 LXC_MEM=256 LXC_CORES=1 ./install-gospeed.sh
#

set -euo pipefail
IFS=$'\n\t'

# ----------------------------
# Configuration (defaults)
# ----------------------------
APP_NAME="${APP_NAME:-gospeed}"
APP_PORT="${APP_PORT:-9090}"

# GitHub release asset base name pattern. Script will append arch (amd64/arm64).
GITHUB_REPO="${GITHUB_REPO:-aneelatwal/gospeed}"
RELEASE_BASE_URL="https://github.com/${GITHUB_REPO}/releases/latest/download"

# LXC settings (can be overridden by env vars)
LXC_ID="${LXC_ID:-next}"       # "next" will auto-pick with pct nextid
LXC_NAME="${LXC_NAME:-gospeed}"
LXC_DISK="${LXC_DISK:-2}"      # GB
LXC_MEM="${LXC_MEM:-256}"      # MB
LXC_CORES="${LXC_CORES:-1}"
LXC_OS="${LXC_OS:-debian}"     # preferred OS family
LXC_VERSION="${LXC_VERSION:-12}" # debian 12 by default

# Storage and network defaults
STORAGE="${STORAGE:-local-lvm}"
BRIDGE="${BRIDGE:-vmbr0}"
HOST_PERSIST_DIR="${HOST_PERSIST_DIR:-/var/lib/gospeed}" # optional host directory to bind-mount
CONTAINER_PERSIST_DIR="${CONTAINER_PERSIST_DIR:-/var/lib/gospeed}"

# Retry / timeouts
WAIT_START_SECS=4
CONTAINER_BOOT_TIMEOUT=60

# Colors
info()    { printf "\033[1;34m[INFO]\033[0m %s\n" "$*"; }
success() { printf "\033[1;32m[SUCCESS]\033[0m %s\n" "$*"; }
warn()    { printf "\033[1;33m[WARN]\033[0m %s\n" "$*"; }
error()   { printf "\033[1;31m[ERROR]\033[0m %s\n" "$*"; }

# ----------------------------
# Sanity checks
# ----------------------------
if ! command -v pct >/dev/null 2>&1; then
  error "This script must be run on a Proxmox VE host with 'pct' available."
  exit 1
fi

if ! command -v pveam >/dev/null 2>&1; then
  warn "'pveam' not found; template auto-download may not be available. Proceeding..."
fi

# Find next available LXC ID if requested
if [ "${LXC_ID}" = "next" ]; then
  LXC_ID=$(pvesh get /cluster/nextid)
  info "Picked next available LXC ID: ${LXC_ID}"
fi

# Keep track whether we created the container (for cleanup)
CREATED_CONTAINER=0

trap_on_error() {
  rc=$?
  if [ "$CREATED_CONTAINER" -eq 1 ]; then
    warn "An error occurred — cleaning up container ${LXC_ID}..."
    pct stop "${LXC_ID}" >/dev/null 2>&1 || true
    pct destroy "${LXC_ID}" >/dev/null 2>&1 || true
  fi
  error "Install failed (exit code ${rc})."
  exit ${rc}
}
trap 'trap_on_error' ERR

# ----------------------------
# Helper: choose template
# ----------------------------
choose_template() {
  # Try Debian 12, fallback to Ubuntu 22.04
  local tmpl=""
  if command -v pveam >/dev/null 2>&1; then
    # Look for a debian 12 standard template in local repo
    if pveam list local | grep -i -E "debian.*${LXC_VERSION}.*standard" >/dev/null 2>&1; then
      tmpl=$(pveam list local | awk '/standard/ && /debian/ && /'"${LXC_VERSION}"'/ {print $1; exit}')
    fi

    if [ -z "$tmpl" ]; then
      # Try ubuntu 22.04
      if pveam list local | grep -i -E "ubuntu.*22.04.*standard" >/dev/null 2>&1; then
        tmpl=$(pveam list local | awk '/standard/ && /ubuntu/ && /22.04/ {print $1; exit}')
      fi
    fi

    # If still empty, attempt to list remote and download
    if [ -z "$tmpl" ]; then
      info "No local template found — checking remote templates..."
      # prefer Debian 12 remote name
      if pveam available | grep -i -E "debian.*${LXC_VERSION}.*standard" >/dev/null 2>&1; then
        remote=$(pveam available | awk '/standard/ && /debian/ && /'"${LXC_VERSION}"'/ {print $1; exit}')
        info "Downloading template ${remote}..."
        pveam download local "${remote}"
        tmpl="local:vztmpl/${remote}"
      elif pveam available | grep -i -E "ubuntu.*22.04.*standard" >/dev/null 2>&1; then
        remote=$(pveam available | awk '/standard/ && /ubuntu/ && /22.04/ {print $1; exit}')
        info "Downloading template ${remote}..."
        pveam download local "${remote}"
        tmpl="local:vztmpl/${remote}"
      fi
    fi
  fi

  # If pveam not available or we couldn't detect, fallback to a commonly named file (best effort)
  if [ -z "$tmpl" ]; then
    # try the common naming pattern under local:vztmpl
    if ls /var/lib/vz/template/cache | grep -i -E "debian-.*${LXC_VERSION}-standard" >/dev/null 2>&1; then
      fname=$(ls /var/lib/vz/template/cache | grep -i -E "debian-.*${LXC_VERSION}-standard" | head -n1)
      tmpl="local:vztmpl/${fname}"
    elif ls /var/lib/vz/template/cache | grep -i -E "ubuntu-.*22.04-standard" >/dev/null 2>&1; then
      fname=$(ls /var/lib/vz/template/cache | grep -i -E "ubuntu-.*22.04-standard" | head -n1)
      tmpl="local:vztmpl/${fname}"
    fi
  fi

  # If still empty, abort
  if [ -z "$tmpl" ]; then
    error "Could not find or download a suitable LXC template (Debian 12 or Ubuntu 22.04). Please add one to your Proxmox templates and re-run."
    exit 1
  fi

  # If tmpl looks like a short name (not containing colon), convert to local:vztmpl/<name>
  if ! echo "$tmpl" | grep -q ":"; then
    tmpl="local:vztmpl/${tmpl}"
  fi

  printf "%s" "$tmpl"
}

# ----------------------------
# Detect architecture (for release asset name)
# ----------------------------
detect_arch() {
  # Determine host architecture and map to our release names
  local arch
  arch=$(dpkg --print-architecture 2>/dev/null || true)
  case "$arch" in
    amd64|x86_64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) 
      warn "Unknown host architecture '${arch}', defaulting to amd64."
      echo "amd64"
      ;;
  esac
}

# ----------------------------
# Main flow
# ----------------------------
info "Starting GoSpeed LXC installer"

TEMPLATE=$(choose_template)
info "Using LXC template: ${TEMPLATE}"

#
# Determine rootfs specification depending on storage type
#
if [[ "${STORAGE}" == *"lvm"* ]]; then
  ROOTFS_SPEC="${STORAGE}:${LXC_DISK}"
else
  ROOTFS_SPEC="${STORAGE}:size=${LXC_DISK}G"
fi

# create container
info "Creating LXC ${LXC_ID} (${LXC_NAME})..."
pct create "${LXC_ID}" "${TEMPLATE}" \
  --hostname "${LXC_NAME}" \
  --cores "${LXC_CORES}" \
  --memory "${LXC_MEM}" \
  --rootfs "${ROOTFS_SPEC}" \
  --net0 "name=eth0,bridge=${BRIDGE},ip=dhcp" \
  --unprivileged 1 \
  --features nesting=1

CREATED_CONTAINER=1

# Start container
info "Starting container ${LXC_ID}..."
pct start "${LXC_ID}"
sleep "${WAIT_START_SECS}"

# Wait for container to get an IP (within timeout)
info "Waiting for container to obtain IP..."
end=$((SECONDS + CONTAINER_BOOT_TIMEOUT))
CONTAINER_IP=""
while [ ${SECONDS} -lt ${end} ]; do
  CONTAINER_IP=$(pct exec "${LXC_ID}" -- hostname -I 2>/dev/null | awk '{print $1}' || true)
  if [ -n "${CONTAINER_IP}" ]; then
    break
  fi
  sleep 2
done

if [ -z "${CONTAINER_IP}" ]; then
  warn "Container did not report an IP within ${CONTAINER_BOOT_TIMEOUT}s; continuing — you may need to inspect the container manually."
else
  info "Container IP: ${CONTAINER_IP}"
fi

# Optionally set up persistent host mount if host dir exists
if [ -n "${HOST_PERSIST_DIR}" ] && [ -d "${HOST_PERSIST_DIR}" ]; then
  info "Mounting host persistence directory ${HOST_PERSIST_DIR} into container at ${CONTAINER_PERSIST_DIR}..."
  pct set "${LXC_ID}" -mp0 "${HOST_PERSIST_DIR},mp=${CONTAINER_PERSIST_DIR}"
else
  info "No host persistence directory found at ${HOST_PERSIST_DIR}; container will store data inside its own filesystem."
fi

# Exec inside container: install deps, create directories, download binary
ARCH=$(detect_arch)
RELEASE_ASSET="gospeed-linux-${ARCH}"
RELEASE_URL="${RELEASE_BASE_URL}/${RELEASE_ASSET}"

info "Deploying GoSpeed binary for architecture ${ARCH} from ${RELEASE_ASSET}"

pct exec "${LXC_ID}" -- bash -c "set -euo pipefail
  apt-get update -y
  apt-get install -y --no-install-recommends curl ca-certificates
  mkdir -p /opt/${APP_NAME}
  cd /opt/${APP_NAME}
  echo 'Downloading ${RELEASE_URL}...'
"

# download binary into container (use curl from host piping into pct exec)
info "Downloading binary into container..."
# Use curl host -> pipe to pct exec > file inside container
curl -fsSL "${RELEASE_URL}" | pct exec "${LXC_ID}" -- bash -c "cat > /opt/${APP_NAME}/${APP_NAME} && chmod +x /opt/${APP_NAME}/${APP_NAME}"

# Create systemd service file inside container
info "Creating systemd service inside container..."
pct exec "${LXC_ID}" -- bash -c "cat > /etc/systemd/system/${APP_NAME}.service <<'EOF'
[Unit]
Description=GoSpeed Network Speed Test
After=network.target

[Service]
ExecStart=/opt/${APP_NAME}/${APP_NAME}
WorkingDirectory=/opt/${APP_NAME}
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable ${APP_NAME}
systemctl start ${APP_NAME}
"

# Small wait to allow service to come up
sleep 2

# Check service status
if pct exec "${LXC_ID}" -- systemctl is-active --quiet "${APP_NAME}"; then
  success "GoSpeed service is running inside container ${LXC_ID}."
else
  warn "GoSpeed service failed to start. Check container logs with: pct exec ${LXC_ID} -- journalctl -u ${APP_NAME} --no-pager"
fi

# fetch IP (re-check)
CONTAINER_IP=$(pct exec "${LXC_ID}" -- hostname -I | awk '{print $1}' || true)

echo "--------------------------------------------------"
success "✅ GoSpeed installed successfully!"
if [ -n "${CONTAINER_IP}" ]; then
  echo "Access it at: http://${CONTAINER_IP}:${APP_PORT}"
else
  echo "Container IP not detected. Use 'pct exec ${LXC_ID} -- hostname -I' to find it."
fi
echo "To view logs: pct exec ${LXC_ID} -- journalctl -u ${APP_NAME} -f"
echo "To connect to container shell: pct enter ${LXC_ID}"
echo "--------------------------------------------------"

# done - disable the trap cleanup
trap - ERR
exit 0