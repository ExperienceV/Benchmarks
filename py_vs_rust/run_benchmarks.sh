#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────
# run_benchmarks.sh — Ejecuta los 3 benchmarks (Python puro,
# Python con librerías, Rust) y guarda los resultados en .md
# ──────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_FILE="$SCRIPT_DIR/BENCHMARK_RESULTS.md"
TIMESTAMP=$(date)
SYSTEM_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Valores por defecto para Rust (en caso de no estar instalado)
RUSTC_VERSION="[No disponible]"
BUILD_MS="N/A"
BINARY_SIZE="N/A"
RUST_OUTPUT="[Rust no instalado]"

echo "═══════════════════════════════════════════════════════════"
echo "  PI BENCHMARK — Suite Completa (Python + Rust)"
echo "═══════════════════════════════════════════════════════════"
echo ""

# ── 1. Python Puro ──────────────────────────────────────────
echo "▶  Ejecutando: Python Puro (Chudnovsky)..."
echo ""

PYTHON_PURE_OUTPUT=$(cd "$SCRIPT_DIR/python" && python pure_python.py 2>&1)
echo "$PYTHON_PURE_OUTPUT"
echo ""

# ── 2. Python con Librerías ─────────────────────────────────
echo "▶  Ejecutando: Python con Librerías (mpmath)..."
echo ""

PYTHON_LIB_OUTPUT=$(cd "$SCRIPT_DIR/python" && python with_library.py 2>&1)
echo "$PYTHON_LIB_OUTPUT"
echo ""

# ── 3. Rust ──────────────────────────────────────────────────
echo "▶  Compilando y ejecutando: Rust (GMP + MPFR)..."
echo ""

cd "$SCRIPT_DIR/rust"

# Verificar si Rust está instalado
if ! command -v rustc &>/dev/null; then
    echo "❌  ERROR: Rust no está instalado"
    echo ""
    echo "Para instalar Rust, ejecuta:"
    echo "  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    echo ""
    echo "Luego recarga el shell:"
    echo "  source \$HOME/.cargo/env"
    echo ""
    RUSTC_VERSION="[NO INSTALADO]"
    BUILD_MS="N/A"
    BINARY_SIZE="N/A"
    RUST_OUTPUT="[Rust no está instalado — ver instrucciones arriba]"
else
    # Mostrar versión de Rust
    RUSTC_VERSION=$(rustc --version)
    echo "Rust Version: $RUSTC_VERSION"
    echo ""

    # Build release
    echo "Compilando (cargo build --release)..."
    BUILD_START=$(date +%s%3N)
    cargo build --release 2>&1 | tail -5
    BUILD_END=$(date +%s%3N)
    BUILD_MS=$(( BUILD_END - BUILD_START ))

    BINARY_SIZE=$(du -sh target/release/pi_benchmark | cut -f1)
    echo "Tiempo de compilación: ${BUILD_MS} ms"
    echo "Tamaño del binario: $BINARY_SIZE"
    echo ""

    # Ejecutar benchmark
    echo "Ejecutando benchmark..."
    RUST_OUTPUT=$(./target/release/pi_benchmark 2>&1)
    echo "$RUST_OUTPUT"
    echo ""
fi

echo ""

# ── Generar archivo markdown con todos los resultados ─────────
cat > "$RESULTS_FILE" << 'MARKDOWN_EOF'
# Benchmark Results — PI Calculation Performance

MARKDOWN_EOF

cat >> "$RESULTS_FILE" << EOF
**Execution Date:** $TIMESTAMP

**System Info:**
- OS: $(uname -s)
- Kernel: $(uname -r)
- Architecture: $(uname -m)
- Python: $(python --version 2>&1 | cut -d' ' -f2)

---

## 1. Python Puro (Algoritmo Chudnovsky)

**Descripción:** Implementación pura en Python usando \`Decimal\` para precisión arbitraria.

### Resultados
\`\`\`
$PYTHON_PURE_OUTPUT
\`\`\`

---

## 2. Python con Librerías (mpmath)

**Descripción:** Usando la librería \`mpmath\` que delega cálculos a GMP a través de libmp.

### Resultados
\`\`\`
$PYTHON_LIB_OUTPUT
\`\`\`

---

## 3. Rust (GMP + MPFR)

**Descripción:** Implementación en Rust usando librerías GMP y MPFR para máximo rendimiento.

**Rust Version:** $RUSTC_VERSION  
**Binary Size:** $BINARY_SIZE  
**Build Time:** ${BUILD_MS}ms

### Resultados
\`\`\`
$RUST_OUTPUT
\`\`\`

---

## Resumen Comparativo

| Aspecto | Python Puro | Python (mpmath) | Rust |
|---------|-------------|-----------------|------|
| **Tipo de implementación** | Chudnovsky en Python puro | Chudnovsky en C (mpmath/libmp) | Chudnovsky en Rust (GMP/MPFR) |
| **Líneas de código** | ~200 | ~100 (deps: ~10K) | N/A (ver src/main.rs) |
| **Compilación** | Ninguna | Ninguna | Sí (${BUILD_MS}ms) |
| **Rendimiento esperado** | ↓ Lento | ↑ Muy rápido | ↑↑ Máximo |
| **Facilidad** | ★★★★★ | ★★★★☆ | ★★★☆☆ |

---

**Generado por:** run_benchmarks.sh  
**Timestamp:** $SYSTEM_DATE
EOF

# ── Finalizar ────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  ✅  Benchmarks completados"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "📄  Resultados guardados en: $RESULTS_FILE"
echo ""
