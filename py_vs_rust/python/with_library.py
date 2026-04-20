import time
import os
import sys
import platform
import resource
import subprocess
from mpmath import mp, pi as mppi

# ─────────────────────────────────────────────
# ALGORITMO: mpmath (Chudnovsky en C via libmp)
# ─────────────────────────────────────────────

def pi_mpmath(digits: int) -> str:
    """Calcula pi con `digits` dígitos decimales usando mpmath."""
    mp.dps = digits + 5
    return mp.nstr(mppi, digits, strip_zeros=False)


def max_pi_digits_in_time(seconds: float) -> dict:
    """Calcula la máxima cantidad de dígitos de pi en `seconds` segundos."""
    digits = 1000
    best_result = None
    latency_1000 = None
    start_total = time.perf_counter()

    while True:
        t0 = time.perf_counter()
        pi_val = pi_mpmath(digits)
        elapsed = time.perf_counter() - t0
        total_elapsed = time.perf_counter() - start_total

        if latency_1000 is None and digits >= 1000:
            latency_1000 = total_elapsed

        best_result = {
            "digits": digits,
            "pi_preview": pi_val[:22] + "...",
            "last_iter_time": elapsed,
            "total_time": total_elapsed,
        }

        remaining = seconds - total_elapsed
        next_digits = int(digits * 2)
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
    try:
        with open("/proc/stat") as f:
            lines = f.readlines()
        cpu_lines = [l for l in lines if l.startswith("cpu")]
        cores = len(cpu_lines) - 1
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
    usage = resource.getrusage(resource.RUSAGE_SELF)
    return round(usage.ru_maxrss / 1024, 2)


def get_energy_watts():
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
                return round((e2 - e1) / 1e6, 2)
            except PermissionError:
                return "sin permisos (ejecuta con sudo)"
    try:
        subprocess.run(
            ["powertop", "--time=1", "--csv=/tmp/pt.csv"],
            capture_output=True, timeout=5
        )
        return "ver /tmp/pt.csv (powertop disponible)"
    except Exception:
        return "N/A (sin RAPL ni powertop)"


def get_mpmath_size_mb():
    """Tamaño del paquete mpmath en site-packages."""
    import mpmath
    pkg_dir = os.path.dirname(mpmath.__file__)
    total = 0
    for dirpath, _, filenames in os.walk(pkg_dir):
        for f in filenames:
            try:
                total += os.path.getsize(os.path.join(dirpath, f))
            except OSError:
                pass
    return round(total / (1024 * 1024), 2)


def count_sloc(filepath: str) -> int:
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
    return cpu_model, platform.machine()


def get_startup_time():
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
        if boot_time_s:
            return round(time.time() - (boot_time_s + start_ticks / clk_tck), 3)
    except Exception:
        pass
    return "N/A"


def get_mpmath_backend():
    """Detecta si mpmath está usando gmpy2 (más rápido) o libmp puro."""
    try:
        import gmpy2
        return "gmpy2 (GMP backend — máximo rendimiento)"
    except ImportError:
        return "libmp (Python puro — instala gmpy2 para +velocidad)"


# ─────────────────────────────────────────────
# MAIN
# ─────────────────────────────────────────────

def run_benchmark():
    script_path = os.path.abspath(__file__)
    startup_time = get_startup_time()

    print("\n" + "═" * 62)
    print("  🔬 PI BENCHMARK — mpmath (libmp / Chudnovsky en C)")
    print("═" * 62)

    cpu_model, arch = get_hw_info()
    kernel = platform.release()
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
    import mpmath
    lib_version = mpmath.__version__
    backend = get_mpmath_backend()

    print("\n📦  ENTORNO")
    print(f"  Interpreter      : CPython {py_version}")
    print(f"  Librería         : mpmath {lib_version}")
    print(f"  Backend          : {backend}")
    print(f"  Optimization     : {'python -O (activo)' if sys.flags.optimize else 'sin flags'}")
    print(f"  CPU              : {cpu_model}")
    print(f"  Arquitectura     : {arch}")
    print(f"  Kernel           : {kernel}")
    print(f"  OS               : {distro}")

    for limit in [10, 60]:
        print(f"\n{'─'*62}")
        print(f"⏱   CORRIDA: {limit} segundos")
        print(f"{'─'*62}")

        rss_before = get_peak_rss_mb()
        r = max_pi_digits_in_time(limit)
        rss_after = get_peak_rss_mb()

        digits     = r["digits"]
        total_time = r["total_time"]
        lat_1000   = r["latency_1000"]
        throughput = round(digits / total_time, 1)
        peak_rss   = max(rss_before, rss_after)

        print("\n  1 · RENDIMIENTO DE CÓMPUTO")
        print(f"    Total Digits   : {digits:,}")
        print(f"    Throughput     : {throughput:,.1f} D/s")
        if lat_1000:
            print(f"    Latency 1,000  : {lat_1000:.4f}s  (primer grupo de 1,000 dígitos)")
        else:
            print(f"    Latency 1,000  : N/A")

        cpu_pct, cores = get_cpu_load()
        cpu_str = (
            f"{cpu_pct}% (Single-threaded sobre {cores} cores disponibles)"
            if cpu_pct else "N/A"
        )
        energy = get_energy_watts()

        print("\n  2 · CONSUMO DE RECURSOS")
        print(f"    Peak RSS       : {peak_rss} MB")
        print(f"    CPU Load Avg   : {cpu_str}")
        print(f"    Energy         : {energy}")

    sloc = count_sloc(script_path)
    mpmath_size = get_mpmath_size_mb()
    startup = f"{startup_time}s" if isinstance(startup_time, float) else startup_time

    print(f"\n{'─'*62}")
    print("  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE")
    print(f"{'─'*62}")
    print(f"    SLOC             : {sloc} líneas lógicas")
    print(f"    Artifact Size    : {mpmath_size} MB  (solo mpmath)")
    print(f"    Startup Time     : {startup}  (desde lanzamiento)")

    print("\n" + "═" * 62)
    print("  ✅  Benchmark completado.")
    print("═" * 62 + "\n")


if __name__ == "__main__":
    run_benchmark()