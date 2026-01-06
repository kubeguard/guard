#!/usr/bin/env python3
"""
Compare k6 Load Test Results for Guard Cache Performance Testing

This script analyzes and compares k6 load test results across different
cache configurations to evaluate the effectiveness of cache improvements.

Usage:
    python3 compare-results.py [results_directory]

Output:
    - Console summary with comparison tables
    - comparison-report.json with detailed metrics
"""

import json
import sys
from pathlib import Path
from collections import defaultdict
from typing import Dict, Any, Optional


def load_summary(path: Path) -> Optional[Dict[str, Any]]:
    """Load k6 summary JSON file."""
    try:
        with open(path) as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"Warning: Could not load {path}: {e}")
        return None


def extract_k6_metrics(summary: Dict[str, Any]) -> Dict[str, float]:
    """Extract key metrics from k6 summary."""
    metrics = summary.get('metrics', {})
    return {
        'http_req_duration_avg': get_metric_value(metrics, 'http_req_duration', 'avg'),
        'http_req_duration_p50': get_metric_value(metrics, 'http_req_duration', 'p(50)'),
        'http_req_duration_p95': get_metric_value(metrics, 'http_req_duration', 'p(95)'),
        'http_req_duration_p99': get_metric_value(metrics, 'http_req_duration', 'p(99)'),
        'http_req_duration_min': get_metric_value(metrics, 'http_req_duration', 'min'),
        'http_req_duration_max': get_metric_value(metrics, 'http_req_duration', 'max'),
        'http_req_failed_rate': get_metric_value(metrics, 'http_req_failed', 'rate'),
        'http_reqs_count': get_metric_value(metrics, 'http_reqs', 'count'),
        'http_reqs_rate': get_metric_value(metrics, 'http_reqs', 'rate'),
    }


def get_metric_value(metrics: Dict, name: str, key: str) -> float:
    """Safely extract metric value."""
    if name in metrics and 'values' in metrics[name]:
        return metrics[name]['values'].get(key, 0)
    return 0


def parse_prometheus_metrics(path: Path) -> Dict[str, float]:
    """Parse Prometheus metrics file."""
    metrics = {}
    try:
        with open(path) as f:
            for line in f:
                if line.startswith('#') or not line.strip():
                    continue
                parts = line.strip().split()
                if len(parts) >= 2:
                    # Handle metrics with labels
                    metric_name = parts[0].split('{')[0]
                    try:
                        metrics[parts[0]] = float(parts[1])
                        # Also store without labels for easy access
                        if metric_name not in metrics:
                            metrics[metric_name] = float(parts[1])
                    except ValueError:
                        continue
    except FileNotFoundError:
        pass
    return metrics


def calculate_cache_metrics(before_metrics: Dict, after_metrics: Dict) -> Dict[str, float]:
    """Calculate cache hit rate from Prometheus metrics."""
    hits_before = before_metrics.get('guard_azure_authz_cache_hits_total', 0)
    misses_before = before_metrics.get('guard_azure_authz_cache_misses_total', 0)
    entries_before = before_metrics.get('guard_azure_authz_cache_entries', 0)

    hits_after = after_metrics.get('guard_azure_authz_cache_hits_total', 0)
    misses_after = after_metrics.get('guard_azure_authz_cache_misses_total', 0)
    entries_after = after_metrics.get('guard_azure_authz_cache_entries', 0)

    total_hits = hits_after - hits_before
    total_misses = misses_after - misses_before
    total = total_hits + total_misses

    return {
        'cache_hits': total_hits,
        'cache_misses': total_misses,
        'cache_hit_rate': total_hits / total if total > 0 else 0,
        'cache_entries_start': entries_before,
        'cache_entries_end': entries_after,
    }


def load_config_results(results_dir: Path, config_name: str) -> Dict[str, Any]:
    """Load all results for a configuration."""
    config_dir = results_dir / config_name
    if not config_dir.exists():
        return {}

    scenarios = ['cache-warmup', 'sustained-load', 'burst-load', 'cache-eviction']
    results = {}

    for scenario in scenarios:
        summary_file = config_dir / f'{scenario}_summary.json'
        metrics_before = config_dir / f'{scenario}_metrics_before.txt'
        metrics_after = config_dir / f'{scenario}_metrics_after.txt'

        if not summary_file.exists():
            continue

        summary = load_summary(summary_file)
        if not summary:
            continue

        # Try to extract from custom summary format first
        if 'metrics' in summary and isinstance(summary['metrics'], dict):
            # This is our custom summary format
            scenario_metrics = summary['metrics']
        else:
            # Standard k6 summary format
            scenario_metrics = extract_k6_metrics(summary)

        # Add cache metrics from Prometheus
        if metrics_before.exists() and metrics_after.exists():
            before = parse_prometheus_metrics(metrics_before)
            after = parse_prometheus_metrics(metrics_after)
            cache_metrics = calculate_cache_metrics(before, after)
            scenario_metrics['prometheus_cache'] = cache_metrics

        results[scenario] = scenario_metrics

    return results


def compare_configs(results_dir: Path) -> Dict[str, Dict[str, Any]]:
    """Compare all configurations."""
    configs = ['master', 'improved_default', 'improved_large', 'improved_long_ttl']
    comparison = {}

    for config in configs:
        results = load_config_results(results_dir, config)
        if results:
            comparison[config] = results

    return comparison


def calculate_improvements(comparison: Dict, baseline: str = 'master') -> Dict[str, Dict[str, float]]:
    """Calculate improvements relative to baseline."""
    if baseline not in comparison:
        return {}

    baseline_data = comparison[baseline]
    improvements = {}

    for config, data in comparison.items():
        if config == baseline:
            continue

        config_improvements = {}

        for scenario in data:
            if scenario not in baseline_data:
                continue

            baseline_scenario = baseline_data[scenario]
            current_scenario = data[scenario]

            # Extract p95 latency
            baseline_p95 = get_nested_value(baseline_scenario, 'latency', 'p95') or \
                           baseline_scenario.get('http_req_duration_p95', 0)
            current_p95 = get_nested_value(current_scenario, 'latency', 'p95') or \
                          current_scenario.get('http_req_duration_p95', 0)

            # Extract cache hit rate
            baseline_cache = get_nested_value(baseline_scenario, 'cache', 'hit_rate') or \
                             get_nested_value(baseline_scenario, 'prometheus_cache', 'cache_hit_rate') or 0
            current_cache = get_nested_value(current_scenario, 'cache', 'hit_rate') or \
                            get_nested_value(current_scenario, 'prometheus_cache', 'cache_hit_rate') or 0

            # Calculate improvements
            p95_improvement = 0
            if baseline_p95 > 0:
                p95_improvement = ((baseline_p95 - current_p95) / baseline_p95) * 100

            cache_improvement = 0
            if baseline_cache > 0:
                cache_improvement = ((current_cache - baseline_cache) / baseline_cache) * 100

            config_improvements[scenario] = {
                'p95_latency_improvement_pct': p95_improvement,
                'cache_hit_rate_improvement_pct': cache_improvement,
                'baseline_p95': baseline_p95,
                'current_p95': current_p95,
                'baseline_cache_hit_rate': baseline_cache,
                'current_cache_hit_rate': current_cache,
            }

        improvements[config] = config_improvements

    return improvements


def get_nested_value(data: Dict, *keys):
    """Safely get nested dictionary value."""
    for key in keys:
        if isinstance(data, dict) and key in data:
            data = data[key]
        else:
            return None
    return data


def print_comparison_table(comparison: Dict, scenario: str):
    """Print comparison table for a specific scenario."""
    print(f"\n{'='*80}")
    print(f"Scenario: {scenario}")
    print(f"{'='*80}")

    headers = ['Config', 'P50(ms)', 'P95(ms)', 'P99(ms)', 'Avg(ms)', 'RPS', 'Cache Hit%']
    header_fmt = "{:<20} {:>10} {:>10} {:>10} {:>10} {:>10} {:>12}"
    print(header_fmt.format(*headers))
    print('-' * 85)

    for config, data in sorted(comparison.items()):
        if scenario not in data:
            continue

        metrics = data[scenario]

        # Try different metric paths
        p50 = get_nested_value(metrics, 'latency', 'p50') or metrics.get('http_req_duration_p50', 0)
        p95 = get_nested_value(metrics, 'latency', 'p95') or metrics.get('http_req_duration_p95', 0)
        p99 = get_nested_value(metrics, 'latency', 'p99') or metrics.get('http_req_duration_p99', 0)
        avg = get_nested_value(metrics, 'latency', 'avg') or metrics.get('http_req_duration_avg', 0)
        rps = get_nested_value(metrics, 'actual_rps') or metrics.get('http_reqs_rate', 0)
        cache_hit = get_nested_value(metrics, 'cache', 'hit_rate') or \
                    get_nested_value(metrics, 'prometheus_cache', 'cache_hit_rate') or 0

        row_fmt = "{:<20} {:>10.2f} {:>10.2f} {:>10.2f} {:>10.2f} {:>10.2f} {:>11.1f}%"
        print(row_fmt.format(config, p50, p95, p99, avg, rps, cache_hit * 100))


def print_improvements(improvements: Dict, baseline: str = 'master'):
    """Print improvement summary."""
    print(f"\n{'='*80}")
    print(f"Improvements vs {baseline}")
    print(f"{'='*80}")

    for config, scenarios in sorted(improvements.items()):
        print(f"\n{config}:")
        for scenario, metrics in scenarios.items():
            p95_imp = metrics.get('p95_latency_improvement_pct', 0)
            cache_imp = metrics.get('cache_hit_rate_improvement_pct', 0)
            print(f"  {scenario}: p95 latency {p95_imp:+.1f}%, cache hit rate {cache_imp:+.1f}%")


def generate_report(comparison: Dict, improvements: Dict, output_path: Path):
    """Generate JSON report."""
    report = {
        'generated_at': str(Path().resolve()),
        'comparison': comparison,
        'improvements': improvements,
        'summary': {
            'configs_tested': list(comparison.keys()),
            'scenarios_tested': list(set(
                scenario
                for config_data in comparison.values()
                for scenario in config_data.keys()
            )),
        },
    }

    with open(output_path, 'w') as f:
        json.dump(report, f, indent=2, default=str)

    print(f"\nDetailed report saved to: {output_path}")


def main():
    results_dir = Path(sys.argv[1]) if len(sys.argv) > 1 else Path('./results')

    if not results_dir.exists():
        print(f"Results directory not found: {results_dir}")
        sys.exit(1)

    print("="*80)
    print("Guard Cache Performance Comparison Report")
    print("="*80)
    print(f"Results directory: {results_dir}")

    # Load and compare results
    comparison = compare_configs(results_dir)

    if not comparison:
        print("No results found to compare")
        sys.exit(1)

    print(f"Configurations found: {', '.join(comparison.keys())}")

    # Get all scenarios
    all_scenarios = set()
    for config_data in comparison.values():
        all_scenarios.update(config_data.keys())

    # Print comparison tables for each scenario
    for scenario in sorted(all_scenarios):
        print_comparison_table(comparison, scenario)

    # Calculate and print improvements
    if 'master' in comparison:
        improvements = calculate_improvements(comparison, 'master')
        print_improvements(improvements, 'master')
    else:
        improvements = {}
        print("\nNote: No 'master' baseline found, skipping improvement calculations")

    # Generate JSON report
    report_path = results_dir / 'comparison-report.json'
    generate_report(comparison, improvements, report_path)

    print("\n" + "="*80)
    print("Comparison complete!")
    print("="*80)


if __name__ == '__main__':
    main()
