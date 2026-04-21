import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import numpy as np

# ─────────────────────────────────────────────
# DATOS
# ─────────────────────────────────────────────

labels     = ["Python puro", "mpmath", "Rust"]
colors     = ["#888780", "#378ADD", "#1D9E75"]
hatch_list = ["", "", ""]

throughput_10s = [2_616,     647_691,    997_648]
throughput_60s = [633,       582_110,    641_496]
digits_10s     = [12_921,    4_096_000,  4_096_000]
digits_60s     = [29_071,    16_384_000, 32_768_000]
rss_60s        = [14.05,     286.3,      308.89]
startup_ms     = [418,       501,        10]
artifact_kb    = [23.1*1024, 4.68*1024,  695]
latency_us     = [7100,      100,        100]

# ─────────────────────────────────────────────
# ESTILO GLOBAL
# ─────────────────────────────────────────────

plt.rcParams.update({
    "figure.facecolor":  "#FAFAF8",
    "axes.facecolor":    "#FAFAF8",
    "axes.spines.top":   False,
    "axes.spines.right": False,
    "axes.spines.left":  False,
    "axes.spines.bottom":False,
    "axes.grid":         True,
    "axes.grid.axis":    "x",
    "grid.color":        "#E0DED8",
    "grid.linewidth":    0.6,
    "font.family":       "sans-serif",
    "font.size":         11,
    "text.color":        "#2C2C2A",
    "axes.labelcolor":   "#5F5E5A",
    "xtick.color":       "#5F5E5A",
    "ytick.color":       "#5F5E5A",
    "xtick.labelsize":   10,
    "ytick.labelsize":   10,
})

x = np.arange(len(labels))
bar_w = 0.55


def fmt_large(n):
    if n >= 1_000_000:
        return f"{n/1_000_000:.1f}M"
    if n >= 1_000:
        return f"{n/1_000:.0f}K"
    return str(int(n))


def bar_labels(ax, rects, fmt_fn=fmt_large, pad=4):
    for r in rects:
        v = r.get_width()
        ax.text(v + pad, r.get_y() + r.get_height() / 2,
                fmt_fn(v), va="center", ha="left",
                fontsize=9.5, color="#5F5E5A")


def hbar(ax, values, title, xlabel, log=False):
    rects = ax.barh(x, values, bar_w,
                    color=colors, edgecolor="none")

    ax.set_yticks(x)
    ax.set_yticklabels(labels, fontsize=11)
    ax.set_title(title, fontsize=13, fontweight="bold",
                 color="#2C2C2A", pad=10, loc="left")
    ax.set_xlabel(xlabel, fontsize=10)
    if log:
        ax.set_xscale("log")
        ax.xaxis.set_major_formatter(
            ticker.FuncFormatter(lambda v, _: fmt_large(v)))
    else:
        ax.xaxis.set_major_formatter(
            ticker.FuncFormatter(lambda v, _: fmt_large(v)))
    ax.invert_yaxis()
    ax.tick_params(axis="y", length=0)
    return rects


# ─────────────────────────────────────────────
# FIGURA: 6 subplots
# ─────────────────────────────────────────────

fig, axes = plt.subplots(3, 2, figsize=(13, 11))
fig.suptitle("PI Benchmark — Python puro vs mpmath vs Rust",
             fontsize=15, fontweight="bold", color="#2C2C2A", y=1.01)
fig.patch.set_facecolor("#FAFAF8")

# ── 1. Throughput 10s ────────────────────────
ax = axes[0, 0]
rects = hbar(ax, throughput_10s,
             "Throughput · 10 s", "Dígitos/segundo (escala log)", log=True)
for r, v in zip(rects, throughput_10s):
    ax.text(v * 1.15, r.get_y() + r.get_height() / 2,
            fmt_large(v) + " D/s",
            va="center", ha="left", fontsize=9, color="#5F5E5A")

# ── 2. Throughput 60s ────────────────────────
ax = axes[0, 1]
rects = hbar(ax, throughput_60s,
             "Throughput · 60 s", "Dígitos/segundo (escala log)", log=True)
for r, v in zip(rects, throughput_60s):
    ax.text(v * 1.15, r.get_y() + r.get_height() / 2,
            fmt_large(v) + " D/s",
            va="center", ha="left", fontsize=9, color="#5F5E5A")

# ── 3. Total dígitos 10s ─────────────────────
ax = axes[1, 0]
rects = hbar(ax, digits_10s,
             "Total dígitos · 10 s", "Dígitos (escala log)", log=True)
for r, v in zip(rects, digits_10s):
    ax.text(v * 1.15, r.get_y() + r.get_height() / 2,
            fmt_large(v),
            va="center", ha="left", fontsize=9, color="#5F5E5A")

# ── 4. Total dígitos 60s ─────────────────────
ax = axes[1, 1]
rects = hbar(ax, digits_60s,
             "Total dígitos · 60 s", "Dígitos (escala log)", log=True)
for r, v in zip(rects, digits_60s):
    ax.text(v * 1.15, r.get_y() + r.get_height() / 2,
            fmt_large(v),
            va="center", ha="left", fontsize=9, color="#5F5E5A")

# ── 5. Peak RSS (60s) ────────────────────────
ax = axes[2, 0]
rects = hbar(ax, rss_60s, "Peak RSS · 60 s", "MB")
for r, v in zip(rects, rss_60s):
    ax.text(v + 2, r.get_y() + r.get_height() / 2,
            f"{v:.0f} MB",
            va="center", ha="left", fontsize=9, color="#5F5E5A")

# ── 6. Startup time + Artifact size ──────────
ax = axes[2, 1]
ax.set_title("Startup & Artifact size", fontsize=13,
             fontweight="bold", color="#2C2C2A", pad=10, loc="left")
ax.axis("off")

col_labels = ["", "Startup", "Artifact"]
rows = [
    ["Python puro", "418 ms", "23.1 MB"],
    ["mpmath",      "501 ms", "4.7 MB"],
    ["Rust",        "10 ms",  "695 KB"],
]
cell_colors = [
    ["#F1EFE8", "#F1EFE8", "#F1EFE8"],
    ["#E6F1FB", "#E6F1FB", "#E6F1FB"],
    ["#E1F5EE", "#E1F5EE", "#E1F5EE"],
]
tbl = ax.table(
    cellText=rows,
    colLabels=col_labels,
    cellLoc="center",
    loc="center",
    cellColours=cell_colors,
)
tbl.auto_set_font_size(False)
tbl.set_fontsize(11)
tbl.scale(1, 2.2)
for (r, c), cell in tbl.get_celld().items():
    cell.set_edgecolor("#D3D1C7")
    cell.set_linewidth(0.5)
    if r == 0:
        cell.set_facecolor("#D3D1C7")
        cell.set_text_props(fontweight="bold", color="#444441")

# ─────────────────────────────────────────────
# LEYENDA GLOBAL
# ─────────────────────────────────────────────

from matplotlib.patches import Patch
legend_elements = [
    Patch(facecolor=colors[i], edgecolor="none", label=labels[i])
    for i in range(3)
]
fig.legend(handles=legend_elements,
           loc="lower center", ncol=3,
           frameon=False, fontsize=11,
           bbox_to_anchor=(0.5, -0.02))

fig.tight_layout(pad=2.5)
plt.savefig("benchmark_charts.png", dpi=150,
            bbox_inches="tight", facecolor="#FAFAF8")
plt.show()
print("Guardado: benchmark_charts.png")