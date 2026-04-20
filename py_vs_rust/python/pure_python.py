import time
import os
import sys
import platform
import resource
import subprocess
import textwrap
from decimal import Decimal, getcontext

# ─────────────────────────────────────────────
# ALGORITMO: Chudnovsky
# ─────────────────────────────────────────────

def pi_chudnovsky(precision: int) -> Decimal:
    """Algoritmo Chudnovsky — converge ~14 dígitos por iteración."""
    getcontext().prec = precision + 10
    C = 426880 * Decimal(10005).sqrt()
    M, X, L, K, S = 1, 1, 13591409, 6, 13591409
    for i in range(1, precision // 14 + 2):
        M = M * (K**3 - 16 * K) // i**3
        X *= -262537412640768000
        L += 545140134
        K += 12
        S += Decimal(M * L) / X
    return C / S


def max_pi_digits_in_time(seconds: float) -> dict:
    """Calcula la máxima cantidad de dígitos de pi en `seconds` segundos."""
    digits = 100
    best_result = None
    latency_1000 = None
    start_total = time.perf_counter()

    while True:
        t0 = time.perf_counter()
        pi_val = pi_chudnovsky(digits)
        elapsed = time.perf_counter() - t0
        total_elapsed = time.perf_counter() - start_total

        if latency_1000 is None and digits >= 1000:
            latency_1000 = total_elapsed

        best_result = {
            "digits": digits,
            "pi_preview": str(pi_val)[:22] + "...",
            "last_iter_time": elapsed,
            "total_time": total_elapsed,
        }

        remaining = seconds - total_elapsed
        next_digits = int(digits * 1.5)
        estimated_next = elapsed * (next_digits / digits) ** 1.5

        if estimated_next > remaining:
            break
        digits = next_digits

    best_result["latency_1000"] = latency_1000
    return best_result


# ─────────────────────────────────────────────
# MÉTRICAS DE SISTEMA
# ─────────────────────────────────────────────

def get_cpu_load():
    """Lee /proc/stat para calcular uso de CPU durante el benchmark."""
    try:
        with open("/proc/stat") as f:
            lines = f.readlines()
        cpu_lines = [l for l in lines if l.startswith("cpu")]
        cores = len(cpu_lines) - 1  # primer renglón = agregado
        first = cpu_lines[0].split()
        total1 = sum(int(x) for x in first[1:])
        idle1 = int(first[4])
        time.sleep(0.2)
        with open("/proc/stat") as f:
            lines = f.readlines()
        first2 = [l for l in lines if l.startswith("cpu")][0].split()
        total2 = sum(int(x) for x in first2[1:])
        idle2 = int(first2[4])
        delta_total = total2 - total1
        delta_idle = idle2 - idle1
        usage = 100.0 * (1 - delta_idle / delta_total) if delta_total else 0.0
        return round(usage, 1), cores
    except Exception:
        return None, None


def get_peak_rss_mb():
    """Retorna el pico de RSS en MB (Linux: getrusage en KB)."""
    usage = resource.getrusage(resource.RUSAGE_SELF)
    return round(usage.ru_maxrss / 1024, 2)


def get_energy_watts():
    """Intenta leer consumo desde RAPL (Intel) o powercap."""
    paths = [
        "/sys/class/powercap/intel-rapl/intel-rapl:0/energy_uj",
        "/sys/class/powercap/intel-rapl:0/energy_uj",
    ]
    for p in paths:
        if os.path.exists(p):
            try:
                with open(p) as f:
                    e1 = int(f.read())
                time.sleep(1.0)
                with open(p) as f:
                    e2 = int(f.read())
                watts = (e2 - e1) / 1e6
                return round(watts, 2)
            except PermissionError:
                return "sin permisos (ejecuta con sudo o ajusta /proc/sys/kernel/perf_event_paranoid)"
    # Fallback: powertop (si está instalado)
    try:
        result = subprocess.run(
            ["powertop", "--time=1", "--csv=/tmp/pt.csv"],
            capture_output=True, timeout=5
        )
        return "ver /tmp/pt.csv (powertop disponible)"
    except Exception:
        return "N/A (sin RAPL ni powertop)"


def get_site_packages_size_mb():
    """Tamaño del directorio site-packages del Python activo."""
    for p in sys.path:
        if "site-packages" in p and os.path.isdir(p):
            total = 0
            for dirpath, _, filenames in os.walk(p):
                for f in filenames:
                    try:
                        total += os.path.getsize(os.path.join(dirpath, f))
                    except OSError:
                        pass
            return round(total / (1024 * 1024), 1)
    return "N/A"


def count_sloc(filepath: str) -> int:
    """Cuenta líneas lógicas (no comentarios ni vacías)."""
    sloc = 0
    with open(filepath) as f:
        for line in f:
            stripped = line.strip()
            if stripped and not stripped.startswith("#"):
                sloc += 1
    return sloc


def get_hw_info():
    cpu_model = "N/A"
    try:
        with open("/proc/cpuinfo") as f:
            for line in f:
                if "model name" in line:
                    cpu_model = line.split(":")[1].strip()
                    break
    except Exception:
        pass
    arch = platform.machine()
    return cpu_model, arch


def get_kernel_version():
    try:
        return platform.release()
    except Exception:
        return "N/A"


def get_startup_time():
    """Tiempo desde inicio del proceso hasta ahora (aproximado)."""
    try:
        pid = os.getpid()
        with open(f"/proc/{pid}/stat") as f:
            fields = f.read().split()
        start_ticks = int(fields[21])
        clk_tck = os.sysconf("SC_CLK_TCK")
        boot_time_s = None
        with open("/proc/stat") as f:
            for line in f:
                if line.startswith("btime"):
                    boot_time_s = int(line.split()[1])
                    break
        if boot_time_s is not None:
            proc_start = boot_time_s + start_ticks / clk_tck
            return round(time.time() - proc_start, 3)
    except Exception:
        pass
    return "N/A"


# ─────────────────────────────────────────────
# MAIN: BENCHMARK COMPLETO
# ─────────────────────────────────────────────

def run_benchmark():
    script_path = os.path.abspath(__file__)
    startup_time = get_startup_time()

    print("\n" + "═" * 60)
    print("  🔬 PI BENCHMARK — Python / Chudnovsky")
    print("═" * 60)

    # ── 4. ENTORNO ──────────────────────────────────────────────
    cpu_model, arch = get_hw_info()
    kernel = get_kernel_version()
    distro = "N/A"
    try:
        with open("/etc/os-release") as f:
            for line in f:
                if line.startswith("PRETTY_NAME"):
                    distro = line.split("=")[1].strip().strip('"')
                    break
    except Exception:
        pass

    py_version = sys.version.split(" ")[0]
    opt_flag = "-O" if sys.flags.optimize else "sin flags (-O no activo)"

    print("\n📦  ENTORNO")
    print(f"  Interpreter     : CPython {py_version}")
    print(f"  Optimization    : python {opt_flag}")
    print(f"  CPU             : {cpu_model}")
    print(f"  Arquitectura    : {arch}")
    print(f"  Kernel          : {kernel}")
    print(f"  OS              : {distro}")

    # ── BENCHMARKS 10s y 60s ────────────────────────────────────
    for limit in [10, 60]:
        print(f"\n{'─'*60}")
        print(f"⏱   CORRIDA: {limit} segundos")
        print(f"{'─'*60}")

        rss_before = get_peak_rss_mb()
        r = max_pi_digits_in_time(limit)
        rss_after = get_peak_rss_mb()

        digits     = r["digits"]
        total_time = r["total_time"]
        lat_1000   = r["latency_1000"]
        throughput = round(digits / total_time, 1)
        peak_rss   = max(rss_before, rss_after)

        # ── 1. RENDIMIENTO DE CÓMPUTO ────────────────────────
        print("\n  1 · RENDIMIENTO DE CÓMPUTO")
        print(f"    Total Digits   : {digits:,}")
        print(f"    Throughput     : {throughput:,.1f} D/s")
        if lat_1000:
            print(f"    Latency 1,000  : {lat_1000:.4f}s  (primer grupo de 1,000 dígitos)")
        else:
            print(f"    Latency 1,000  : N/A (no alcanzó 1,000 dígitos)")

        # ── 2. CONSUMO DE RECURSOS ───────────────────────────
        cpu_pct, cores = get_cpu_load()
        cpu_str = f"{cpu_pct}% ({'Single-Core' if cores == 1 else f'Multi-Core ({cores} cores disponibles, uso single-threaded)'})" if cpu_pct else "N/A"
        energy = get_energy_watts()

        print("\n  2 · CONSUMO DE RECURSOS")
        print(f"    Peak RSS       : {peak_rss} MB")
        print(f"    CPU Load Avg   : {cpu_str}")
        print(f"    Energy         : {energy}")

    # ── 3. EFICIENCIA DE DESPLIEGUE ──────────────────────────
    sloc = count_sloc(script_path)
    sp_size = get_site_packages_size_mb()
    startup = f"{startup_time}s" if isinstance(startup_time, float) else startup_time

    print(f"\n{'─'*60}")
    print("  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE")
    print(f"{'─'*60}")
    print(f"    SLOC            : {sloc} líneas lógicas")
    print(f"    Artifact Size   : {sp_size} MB  (site-packages)")
    print(f"    Startup Time    : {startup}  (desde lanzamiento hasta cálculo)")

    print("\n" + "═" * 60)
    print("  ✅  Benchmark completado.")
    print("═" * 60 + "\n")


if __name__ == "__main__":
    run_benchmark()