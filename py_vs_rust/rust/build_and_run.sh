#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────
#  build_and_run.sh — Instala dependencias, compila y ejecuta
#  el benchmark de pi en Rust (Fedora / dnf)
# ──────────────────────────────────────────────────────────────
set -euo pipefail

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  PI BENCHMARK — Setup & Build (Fedora)"
echo "═══════════════════════════════════════════════════════════"

# ── 1. Dependencias del sistema (gmp, mpfr, libmpc) ──────────
echo ""
echo "▶  Instalando dependencias del sistema..."
sudo dnf install -y \
    gmp-devel \
    mpfr-devel \
    libmpc-devel \
    gcc \
    2>/dev/null || true

# ── 2. Rust (si no está instalado) ───────────────────────────
if ! command -v cargo &>/dev/null; then
    echo ""
    echo "▶  Instalando Rust via rustup..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source "$HOME/.cargo/env"
else
    echo "▶  Rust ya instalado: $(rustc --version)"
fi

source "$HOME/.cargo/env" 2>/dev/null || true

# ── 3. Exportar versión del compilador para embed ────────────
export RUSTC_VERSION
RUSTC_VERSION=$(rustc --version | awk '{print $2}')

# ── 4. Build release ─────────────────────────────────────────
echo ""
echo "▶  Compilando (cargo build --release)..."
BUILD_START=$(date +%s%3N)

cargo build --release 2>&1

BUILD_END=$(date +%s%3N)
BUILD_MS=$(( BUILD_END - BUILD_START ))

echo ""
echo "  Tiempo de compilación : ${BUILD_MS} ms"
echo "  Tamaño del binario    : $(du -sh target/release/pi_benchmark | cut -f1)"

# ── 5. Ejecutar ───────────────────────────────────────────────
echo ""
echo "▶  Ejecutando benchmark..."
echo ""

time ./target/release/pi_benchmark