import os
import json
from typing import Dict, List

import matplotlib.pyplot as plt
import seaborn as sns

# Directory where benchmark JSON files are stored
BENCHMARK_DIR = os.path.join(os.path.dirname(__file__), "benchmarks")

# Ensure the directory exists
if not os.path.isdir(BENCHMARK_DIR):
    raise SystemExit(f"Benchmarks directory not found: {BENCHMARK_DIR}")


def load_benchmarks(directory: str) -> List[Dict]:
    """Load all JSON benchmark files from the directory."""
    benchmarks = []
    for filename in os.listdir(directory):
        if filename.startswith("trufflehog_pipeline_benchmark_") and filename.endswith(".json"):
            path = os.path.join(directory, filename)
            try:
                with open(path, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    benchmarks.append({"filename": filename, "data": data})
            except (json.JSONDecodeError, FileNotFoundError) as exc:
                print(f"Skipping {filename}: {exc}")
    return benchmarks


def plot_pipeline_bottlenecks(benchmark: Dict, out_dir: str) -> str:
    """Plot channel utilization for each pipeline stage."""
    stages = []
    utilizations = []
    severities = []
    for entry in benchmark["data"].get("pipeline_bottlenecks", []):
        stages.append(entry.get("stage", "unknown"))
        utilizations.append(entry.get("channel_utilization_percent", 0))
        severities.append(entry.get("bottleneck_severity", "LOW"))

    if not stages:
        print(f"No bottleneck data for {benchmark['filename']}")
        return ""

    plt.figure(figsize=(8, 4))
    palette = {"LOW": "#2ecc71", "MEDIUM": "#f1c40f", "HIGH": "#e74c3c"}
    colors = [palette.get(sev, "#95a5a6") for sev in severities]
    sns.barplot(x=stages, y=utilizations, palette=colors)
    plt.ylabel("Channel Utilization (%)")
    plt.title("Pipeline Channel Utilization")
    plt.ylim(0, 100)
    plt.tight_layout()

    outfile = os.path.join(out_dir, benchmark["filename"].replace(".json", "_bottlenecks.png"))
    plt.savefig(outfile, dpi=150)
    plt.close()
    return outfile


def plot_worker_utilization(benchmark: Dict, out_dir: str) -> str:
    """Plot worker utilization percentages."""
    util_map = benchmark["data"].get("concurrency_metrics", {}).get("worker_utilization_percent", {})
    if not util_map:
        print(f"No worker utilization data for {benchmark['filename']}")
        return ""

    workers, utils = zip(*util_map.items())
    plt.figure(figsize=(8, 4))
    sns.barplot(x=list(workers), y=list(utils), palette="Blues_d")
    plt.ylabel("Utilization (%)")
    plt.title("Worker Utilization")
    plt.ylim(0, 100)
    plt.tight_layout()

    outfile = os.path.join(out_dir, benchmark["filename"].replace(".json", "_workers.png"))
    plt.savefig(outfile, dpi=150)
    plt.close()
    return outfile


def main():
    benchmarks = load_benchmarks(BENCHMARK_DIR)
    if not benchmarks:
        print("No benchmark JSON files found; exiting.")
        return

    output_dir = os.path.join(BENCHMARK_DIR, "charts")
    os.makedirs(output_dir, exist_ok=True)

    generated = []
    for benchmark in benchmarks:
        png1 = plot_pipeline_bottlenecks(benchmark, output_dir)
        if png1:
            generated.append(png1)
        png2 = plot_worker_utilization(benchmark, output_dir)
        if png2:
            generated.append(png2)

    if generated:
        print("Generated PNG files:")
        for f in generated:
            print(" -", f)
    else:
        print("No PNG files generated; check benchmark data.")


if __name__ == "__main__":
    main()