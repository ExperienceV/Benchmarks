// ─────────────────────────────────────────────────────────────
//  PI BENCHMARK — Rust / rug (GMP + MPFR)
//  Algoritmo: Chudnovsky — ~14 dígitos decimales por iteración
// ─────────────────────────────────────────────────────────────

use rug::{Float, Integer, ops::Pow};
use std::fs;
use std::time::{Duration, Instant};

// ─────────────────────────────────────────────
// ALGORITMO: Chudnovsky con rug::Float (MPFR)
// ─────────────────────────────────────────────

fn digits_to_bits(digits: u32) -> u32 {
    // 1 dígito decimal ≈ 3.32193 bits
    ((digits as f64) * 3.321928 + 64.0) as u32
}

fn pi_chudnovsky(digits: u32) -> Float {
    let prec = digits_to_bits(digits + 10);

    // Constante C = 426880 * sqrt(10005)
    let c = {
        let mut v = Float::with_val(prec, 10005u32);
        v.sqrt_mut();
        v *= 426880u32;
        v
    };

    let mut big_m   = Integer::from(1u32);
    let mut big_x   = Integer::from(1i32);
    let mut big_l   = Integer::from(13591409u32);
    let mut big_k   = Integer::from(6i32);
    let mut s       = Float::with_val(prec, 13591409u32);

    let iters = digits / 14 + 2;

    for i in 1u32..=iters {
        // M = M * (K^3 - 16*K) / i^3
        let k3 = big_k.clone().pow(3u32);
        let k16 = big_k.clone() * 16;
        let num = k3 - k16;
        big_m *= num;
        big_m /= Integer::from(i).pow(3u32);

        // X *= -262537412640768000
        big_x *= Integer::from(-262537412640768000i64);

        // L += 545140134
        big_l += 545140134u32;

        // K += 12
        big_k += 12;

        // S += M * L / X
        let term = {
            let ml = Float::with_val(prec, &big_m) * Float::with_val(prec, &big_l);
            ml / Float::with_val(prec, &big_x)
        };
        s += term;
    }

    c / s
}

fn max_pi_in_time(seconds: f64) -> (u32, f64, Option<f64>) {
    let mut digits: u32 = 1_000;
    let mut best_digits = digits;
    let mut latency_1000: Option<f64> = None;
    let start_total = Instant::now();

    loop {
        let t0 = Instant::now();
        let _ = pi_chudnovsky(digits);
        let elapsed = t0.elapsed().as_secs_f64();
        let total   = start_total.elapsed().as_secs_f64();

        if latency_1000.is_none() && digits >= 1_000 {
            latency_1000 = Some(total);
        }

        best_digits = digits;

        let remaining     = seconds - total;
        let next_digits   = digits * 2;
        let estimated_next = elapsed * (next_digits as f64 / digits as f64).powf(1.5);

        if estimated_next > remaining {
            break;
        }
        digits = next_digits;
    }

    let total_time = start_total.elapsed().as_secs_f64();
    (best_digits, total_time, latency_1000)
}

// ─────────────────────────────────────────────
// MÉTRICAS DE SISTEMA
// ─────────────────────────────────────────────

fn get_peak_rss_mb() -> f64 {
    // /proc/self/status → VmRSS (kB)
    if let Ok(status) = fs::read_to_string("/proc/self/status") {
        for line in status.lines() {
            if line.starts_with("VmRSS:") {
                let kb: u64 = line
                    .split_whitespace()
                    .nth(1)
                    .and_then(|s| s.parse().ok())
                    .unwrap_or(0);
                return kb as f64 / 1024.0;
            }
        }
    }
    0.0
}

fn get_peak_rss_vmhwm_mb() -> f64 {
    // VmHWM = peak RSS histórico
    if let Ok(status) = fs::read_to_string("/proc/self/status") {
        for line in status.lines() {
            if line.starts_with("VmHWM:") {
                let kb: u64 = line
                    .split_whitespace()
                    .nth(1)
                    .and_then(|s| s.parse().ok())
                    .unwrap_or(0);
                return kb as f64 / 1024.0;
            }
        }
    }
    0.0
}

fn get_cpu_usage_percent() -> (f64, usize) {
    // Lee /proc/stat dos veces con 200ms de pausa
    let read_stat = || -> Option<(u64, u64, usize)> {
        let content = fs::read_to_string("/proc/stat").ok()?;
        let mut cores = 0usize;
        let mut total = 0u64;
        let mut idle  = 0u64;
        for line in content.lines() {
            if line.starts_with("cpu ") {
                let f: Vec<u64> = line.split_whitespace()
                    .skip(1)
                    .filter_map(|s| s.parse().ok())
                    .collect();
                total = f.iter().sum();
                idle  = f.get(3).copied().unwrap_or(0);
            } else if line.starts_with("cpu") {
                cores += 1;
            }
        }
        Some((total, idle, cores))
    };

    if let (Some((t1, i1, cores)), _) = (read_stat(), std::thread::sleep(Duration::from_millis(200))) {
        if let Some((t2, i2, _)) = read_stat() {
            let dt = t2.saturating_sub(t1);
            let di = i2.saturating_sub(i1);
            if dt > 0 {
                let pct = 100.0 * (1.0 - di as f64 / dt as f64);
                return (pct, cores);
            }
        }
    }
    (0.0, 0)
}

fn get_energy_watts() -> String {
    let paths = [
        "/sys/class/powercap/intel-rapl/intel-rapl:0/energy_uj",
        "/sys/class/powercap/intel-rapl:0/energy_uj",
    ];
    for p in &paths {
        if let Ok(s) = fs::read_to_string(p) {
            if let Ok(e1) = s.trim().parse::<u64>() {
                std::thread::sleep(Duration::from_secs(1));
                if let Ok(s2) = fs::read_to_string(p) {
                    if let Ok(e2) = s2.trim().parse::<u64>() {
                        let watts = (e2.saturating_sub(e1)) as f64 / 1e6;
                        return format!("{:.2} W (RAPL)", watts);
                    }
                }
            }
        }
    }
    "N/A (sin RAPL — usa sudo o ajusta perf_event_paranoid)".to_string()
}

fn get_binary_size_kb() -> f64 {
    // Tamaño del propio binario en ejecución
    if let Ok(path) = std::env::current_exe() {
        if let Ok(meta) = fs::metadata(&path) {
            return meta.len() as f64 / 1024.0;
        }
    }
    0.0
}

fn get_cpu_model() -> String {
    if let Ok(info) = fs::read_to_string("/proc/cpuinfo") {
        for line in info.lines() {
            if line.starts_with("model name") {
                return line.splitn(2, ':').nth(1).unwrap_or("N/A").trim().to_string();
            }
        }
    }
    "N/A".to_string()
}

fn get_arch() -> String {
    std::env::consts::ARCH.to_string()
}

fn get_kernel() -> String {
    if let Ok(v) = fs::read_to_string("/proc/version") {
        v.split_whitespace().nth(2).unwrap_or("N/A").to_string()
    } else {
        "N/A".to_string()
    }
}

fn get_distro() -> String {
    if let Ok(content) = fs::read_to_string("/etc/os-release") {
        for line in content.lines() {
            if line.starts_with("PRETTY_NAME") {
                return line.splitn(2, '=').nth(1)
                    .unwrap_or("N/A")
                    .trim_matches('"')
                    .to_string();
            }
        }
    }
    "N/A".to_string()
}

fn get_startup_time_ms() -> f64 {
    // /proc/self/stat campo 22 = starttime en ticks desde boot
    if let (Ok(stat), Ok(uptime_str)) = (
        fs::read_to_string("/proc/self/stat"),
        fs::read_to_string("/proc/uptime"),
    ) {
        let fields: Vec<&str> = stat.split_whitespace().collect();
        if let Some(ticks_str) = fields.get(21) {
            if let Ok(ticks) = ticks_str.parse::<u64>() {
                let clk_tck = unsafe { libc::sysconf(libc::_SC_CLK_TCK) } as f64;
                let uptime: f64 = uptime_str.split_whitespace()
                    .next().and_then(|s| s.parse().ok()).unwrap_or(0.0);
                let proc_start = ticks as f64 / clk_tck;
                let age = uptime - proc_start;
                return (age * 1000.0).max(0.0);
            }
        }
    }
    0.0
}

fn count_sloc() -> usize {
    // Cuenta líneas lógicas del propio fuente (main.rs)
    let src = include_str!("main.rs");
    src.lines()
        .filter(|l| {
            let t = l.trim();
            !t.is_empty() && !t.starts_with("//")
        })
        .count()
}

// ─────────────────────────────────────────────
// MAIN
// ─────────────────────────────────────────────

fn main() {
    let startup_ms = get_startup_time_ms();

    println!("\n{}", "═".repeat(64));
    println!("  🔬 PI BENCHMARK — Rust / rug (GMP + MPFR)");
    println!("{}", "═".repeat(64));

    // ── 4. ENTORNO ──────────────────────────────────────────────
    println!("\n📦  ENTORNO");
    println!("  Compiler         : rustc {}", rustc_version());
    println!("  Librería         : rug 1.24 (GMP + MPFR via C FFI)");
    println!(
        "  Optimization     : cargo build --release  \
        (opt-level=3, LTO=fat, codegen-units=1, strip=true)"
    );
    println!("  CPU              : {}", get_cpu_model());
    println!("  Arquitectura     : {}", get_arch());
    println!("  Kernel           : {}", get_kernel());
    println!("  OS               : {}", get_distro());

    // ── CORRIDAS ────────────────────────────────────────────────
    for &limit in &[10.0f64, 60.0f64] {
        println!("\n{}", "─".repeat(64));
        println!("⏱   CORRIDA: {:.0} segundos", limit);
        println!("{}", "─".repeat(64));

        let rss_before = get_peak_rss_mb();
        let (digits, total_time, latency_1000) = max_pi_in_time(limit);
        let rss_after   = get_peak_rss_vmhwm_mb();
        let peak_rss    = rss_after.max(rss_before);
        let throughput  = digits as f64 / total_time;

        println!("\n  1 · RENDIMIENTO DE CÓMPUTO");
        println!("    Total Digits   : {:>12}", format_number(digits as u64));
        println!("    Throughput     : {:>12} D/s", format_fp(throughput));
        match latency_1000 {
            Some(l) => println!("    Latency 1,000  : {:.4}s  (primer grupo de 1,000 dígitos)", l),
            None    => println!("    Latency 1,000  : N/A"),
        }

        let (cpu_pct, cores) = get_cpu_usage_percent();
        let energy = get_energy_watts();

        println!("\n  2 · CONSUMO DE RECURSOS");
        println!("    Peak RSS       : {:.2} MB", peak_rss);
        println!(
            "    CPU Load Avg   : {:.1}%  (single-threaded, {} cores disponibles)",
            cpu_pct, cores
        );
        println!("    Energy         : {}", energy);
    }

    // ── 3. DESPLIEGUE ───────────────────────────────────────────
    let bin_kb   = get_binary_size_kb();
    let sloc     = count_sloc();

    println!("\n{}", "─".repeat(64));
    println!("  3 · EFICIENCIA DE DESARROLLO Y DESPLIEGUE");
    println!("{}", "─".repeat(64));
    println!("    SLOC             : {} líneas lógicas", sloc);
    println!("    Binary Size      : {:.1} KB  (--release, LTO, strip)", bin_kb);
    println!("    Startup Time     : {:.1} ms  (desde lanzamiento)", startup_ms);

    println!("\n{}", "═".repeat(64));
    println!("  ✅  Benchmark completado.");
    println!("{}\n", "═".repeat(64));
}

// ─────────────────────────────────────────────
// UTILIDADES
// ─────────────────────────────────────────────

fn rustc_version() -> String {
    // Embebida en compilación via env!() trick — fallback al string del build
    option_env!("RUSTC_VERSION").unwrap_or("1.x (desconocido)").to_string()
}

fn format_number(n: u64) -> String {
    let s = n.to_string();
    let mut result = String::new();
    for (i, c) in s.chars().rev().enumerate() {
        if i > 0 && i % 3 == 0 { result.push(','); }
        result.push(c);
    }
    result.chars().rev().collect()
}

fn format_fp(n: f64) -> String {
    format_number(n as u64)
}