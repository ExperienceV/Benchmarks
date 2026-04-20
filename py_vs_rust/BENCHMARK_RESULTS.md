# Benchmark Results — PI Calculation Performance

**Execution Date:** Mon Apr 20 23:04:00 UTC 2026

**System Info:**
- OS: Linux
- Kernel: 6.8.0-1044-azure
- Architecture: x86_64
- Python: 3.12.1

---

## 1. Python Puro (Algoritmo Chudnovsky)

**Descripción:** Implementación pura en Python usando `Decimal` para precisión arbitraria.

### Resultados
```

════════════════════════════════════════════════════════════
  🔬 PI BENCHMARK — Python / Chudnovsky
════════════════════════════════════════════════════════════

📦  ENTORNO
  Interpreter     : CPython 3.12.1
  Optimization    : python sin flags (-O no activo)
  CPU             : AMD EPYC 7763 64-Core Processor
  Arquitectura    : x86_64
  Kernel          : 6.8.0-1044-azure
  OS              : Ubuntu 24.04.4 LTS

────────────────────────────────────────────────────────────
⏱   CORRIDA: 10 segundos
────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   : 12,921
    Throughput     : 2,616.1 D/s
    Latency 1,000  : 0.0071s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 13.62 MB
    CPU Load Avg   : 2.6% (Multi-Core (2 cores disponibles, uso single-threaded))
    Energy         : N/A (sin RAPL ni powertop)

────────────────────────────────────────────────────────────
⏱   CORRIDA: 60 segundos
────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   : 29,071
    Throughput     : 633.3 D/s
    Latency 1,000  : 0.0117s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 14.05 MB
    CPU Load Avg   : N/A
    Energy         : N/A (sin RAPL ni powertop)

────────────────────────────────────────────────────────────
  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE
────────────────────────────────────────────────────────────
    SLOC            : 224 líneas lógicas
    Artifact Size   : 23.1 MB  (site-packages)
    Startup Time    : 0.418s  (desde lanzamiento hasta cálculo)

════════════════════════════════════════════════════════════
  ✅  Benchmark completado.
════════════════════════════════════════════════════════════
```

---

## 2. Python con Librerías (mpmath)

**Descripción:** Usando la librería `mpmath` que delega cálculos a GMP a través de libmp.

### Resultados
```

══════════════════════════════════════════════════════════════
  🔬 PI BENCHMARK — mpmath (libmp / Chudnovsky en C)
══════════════════════════════════════════════════════════════

📦  ENTORNO
  Interpreter      : CPython 3.12.1
  Librería         : mpmath 1.4.1
  Backend          : gmpy2 (GMP backend — máximo rendimiento)
  Optimization     : sin flags
  CPU              : AMD EPYC 7763 64-Core Processor
  Arquitectura     : x86_64
  Kernel           : 6.8.0-1044-azure
  OS               : Ubuntu 24.04.4 LTS

──────────────────────────────────────────────────────────────
⏱   CORRIDA: 10 segundos
──────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   : 4,096,000
    Throughput     : 647,691.0 D/s
    Latency 1,000  : 0.0003s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 88.03 MB
    CPU Load Avg   : 5.0% (Single-threaded sobre 2 cores disponibles)
    Energy         : N/A (sin RAPL ni powertop)

──────────────────────────────────────────────────────────────
⏱   CORRIDA: 60 segundos
──────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   : 16,384,000
    Throughput     : 582,109.8 D/s
    Latency 1,000  : 0.0001s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 286.3 MB
    CPU Load Avg   : 2.6% (Single-threaded sobre 2 cores disponibles)
    Energy         : N/A (sin RAPL ni powertop)

──────────────────────────────────────────────────────────────
  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE
──────────────────────────────────────────────────────────────
    SLOC             : 215 líneas lógicas
    Artifact Size    : 4.68 MB  (solo mpmath)
    Startup Time     : 0.501s  (desde lanzamiento)

══════════════════════════════════════════════════════════════
  ✅  Benchmark completado.
══════════════════════════════════════════════════════════════
```

---

## 3. Rust (GMP + MPFR)

**Descripción:** Implementación en Rust usando librerías GMP y MPFR para máximo rendimiento.

**Rust Version:** rustc 1.95.0 (59807616e 2026-04-14)  
**Binary Size:** 696K  
**Build Time:** 3917ms

### Resultados
```

════════════════════════════════════════════════════════════════
  🔬 PI BENCHMARK — Rust / rug (GMP + MPFR)
════════════════════════════════════════════════════════════════

📦  ENTORNO
  Compiler         : rustc 1.x (desconocido)
  Librería         : rug 1.24 (GMP + MPFR via C FFI)
  Optimization     : cargo build --release  (opt-level=3, LTO=fat, codegen-units=1, strip=true)
  CPU              : AMD EPYC 7763 64-Core Processor
  Arquitectura     : x86_64
  Kernel           : 6.8.0-1044-azure
  OS               : Ubuntu 24.04.4 LTS

────────────────────────────────────────────────────────────────
⏱   CORRIDA: 10 segundos
────────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   :    4,096,000
    Throughput     :      997,648 D/s
    Latency 1,000  : 0.0021s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 41.18 MB
    CPU Load Avg   : 0.0%  (single-threaded, 2 cores disponibles)
    Energy         : N/A (sin RAPL — usa sudo o ajusta perf_event_paranoid)

────────────────────────────────────────────────────────────────
⏱   CORRIDA: 60 segundos
────────────────────────────────────────────────────────────────

  1 · RENDIMIENTO DE CÓMPUTO
    Total Digits   :   32,768,000
    Throughput     :      641,496 D/s
    Latency 1,000  : 0.0001s  (primer grupo de 1,000 dígitos)

  2 · CONSUMO DE RECURSOS
    Peak RSS       : 308.89 MB
    CPU Load Avg   : 0.0%  (single-threaded, 2 cores disponibles)
    Energy         : N/A (sin RAPL — usa sudo o ajusta perf_event_paranoid)

────────────────────────────────────────────────────────────────
  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE
────────────────────────────────────────────────────────────────
    SLOC             : 297 líneas lógicas
    Binary Size      : 694.6 KB  (--release, LTO, strip)
    Startup Time     : 10.0 ms  (desde lanzamiento)

════════════════════════════════════════════════════════════════
  ✅  Benchmark completado.
════════════════════════════════════════════════════════════════
```

