#!/usr/bin/env python3
"""Generates grafana/slurm-dashboard.json from the metrics exposed by
internal/collectors/. Run after adding/renaming any slurm_* metric:

    python3 grafana/build_dashboard.py
"""
import json
import os

DS = {"type": "prometheus", "uid": "${DS_PROMETHEUS}"}

panels = []
_id = [0]


def next_id():
    _id[0] += 1
    return _id[0]


def target(expr, legend=None, ref="A", instant=False):
    t = {"datasource": DS, "expr": expr, "refId": ref}
    if legend:
        t["legendFormat"] = legend
    if instant:
        t["instant"] = True
        t["format"] = "table"
    return t


def row(title, y):
    panels.append({
        "id": next_id(),
        "type": "row",
        "title": title,
        "collapsed": False,
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": y},
        "panels": [],
    })


def stat_panel(title, expr, x, y, w=4, h=4, unit="short", thresholds=None):
    panels.append({
        "id": next_id(),
        "type": "stat",
        "title": title,
        "datasource": DS,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "thresholds": thresholds or {
                    "mode": "absolute",
                    "steps": [{"color": "green", "value": None}],
                },
            },
            "overrides": [],
        },
        "options": {
            "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": False},
            "orientation": "auto",
            "textMode": "auto",
            "colorMode": "value",
            "graphMode": "none",
        },
        "targets": [target(expr)],
    })


def piechart_panel(title, targets, x, y, w=12, h=8):
    panels.append({
        "id": next_id(),
        "type": "piechart",
        "title": title,
        "datasource": DS,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "fieldConfig": {"defaults": {"unit": "short"}, "overrides": []},
        "options": {
            "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": False},
            "legend": {"displayMode": "table", "placement": "right", "values": ["value"]},
            "pieType": "donut",
        },
        "targets": targets,
    })


def bargauge_panel(title, targets, x, y, w=12, h=8, unit="short"):
    panels.append({
        "id": next_id(),
        "type": "bargauge",
        "title": title,
        "datasource": DS,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "min": 0,
                "thresholds": {
                    "mode": "absolute",
                    "steps": [
                        {"color": "green", "value": None},
                        {"color": "yellow", "value": 70},
                        {"color": "red", "value": 90},
                    ],
                },
            },
            "overrides": [],
        },
        "options": {
            "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": False},
            "orientation": "horizontal",
            "displayMode": "gradient",
        },
        "targets": targets,
    })


def timeseries_panel(title, targets, x, y, w=12, h=8, unit="short"):
    panels.append({
        "id": next_id(),
        "type": "timeseries",
        "title": title,
        "datasource": DS,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "custom": {
                    "drawStyle": "line",
                    "lineInterpolation": "linear",
                    "lineWidth": 1,
                    "fillOpacity": 10,
                    "showPoints": "never",
                    "spanNulls": True,
                },
            },
            "overrides": [],
        },
        "options": {
            "legend": {"displayMode": "table", "placement": "bottom", "calcs": ["lastNotNull", "max"]},
            "tooltip": {"mode": "multi"},
        },
        "targets": targets,
    })


def table_panel(title, targets, x, y, rename, units=None, w=24, h=8):
    """targets: list of (expr, legend) pairs, one per refId A, B, C...
    rename: dict mapping 'Value #A' etc -> display name
    units: optional dict mapping display name -> unit string
    """
    refs = "ABCDEFGHIJ"
    built = [target(expr, legend, ref=refs[i], instant=True) for i, (expr, legend) in enumerate(targets)]

    overrides = []
    for name, unit in (units or {}).items():
        overrides.append({
            "matcher": {"id": "byName", "options": name},
            "properties": [{"id": "unit", "value": unit}],
        })

    panels.append({
        "id": next_id(),
        "type": "table",
        "title": title,
        "datasource": DS,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "fieldConfig": {"defaults": {"unit": "short"}, "overrides": overrides},
        "options": {
            "showHeader": True,
            "cellHeight": "sm",
        },
        "targets": built,
        "transformations": [
            {"id": "labelsToFields", "options": {}},
            {"id": "filterFieldsByName", "options": {"exclude": {"names": ["Time"]}}},
            {"id": "merge", "options": {}},
            {"id": "organize", "options": {"renameByName": rename}},
        ],
    })


# ---------------------------------------------------------------------------
# Cluster overview
# ---------------------------------------------------------------------------
row("Cluster Overview", y=0)
stat_panel("CPUs Allocated", "slurm_cpus_alloc", x=0, y=1)
stat_panel("CPUs Idle", "slurm_cpus_idle", x=4, y=1)
stat_panel("CPUs Other", "slurm_cpus_other", x=8, y=1)
stat_panel("CPUs Total", "slurm_cpus_total", x=12, y=1)
stat_panel("GPUs Allocated", "slurm_gpus_alloc", x=16, y=1)
stat_panel(
    "GPU Utilization", "slurm_gpus_utilization * 100", x=20, y=1, unit="percent",
    thresholds={
        "mode": "absolute",
        "steps": [
            {"color": "blue", "value": None},
            {"color": "green", "value": 1},
        ],
    },
)

piechart_panel(
    "Node States",
    [
        target("slurm_nodes_alloc", "alloc", "A"),
        target("slurm_nodes_idle", "idle", "B"),
        target("slurm_nodes_mix", "mix", "C"),
        target("slurm_nodes_down", "down", "D"),
        target("slurm_nodes_drain", "drain", "E"),
        target("slurm_nodes_err", "err", "F"),
        target("slurm_nodes_fail", "fail", "G"),
        target("slurm_nodes_maint", "maint", "H"),
        target("slurm_nodes_resv", "resv", "I"),
        target("slurm_nodes_comp", "comp", "J"),
    ],
    x=0, y=5,
)
bargauge_panel(
    "Per-Node CPU Allocation %",
    [target("slurm_node_cpu_alloc / slurm_node_cpu_total * 100", "{{node}}")],
    x=12, y=5, unit="percent",
)

# ---------------------------------------------------------------------------
# Job queue
# ---------------------------------------------------------------------------
row("Job Queue", y=13)
stat_panel("Pending", "slurm_queue_pending", x=0, y=14)
stat_panel("Running", "slurm_queue_running", x=4, y=14)
stat_panel("Suspended", "slurm_queue_suspended", x=8, y=14)
stat_panel("Completing", "slurm_queue_completing", x=12, y=14)
stat_panel(
    "Failed", "slurm_queue_failed", x=16, y=14,
    thresholds={
        "mode": "absolute",
        "steps": [
            {"color": "green", "value": None},
            {"color": "red", "value": 1},
        ],
    },
)
stat_panel(
    "Timeout", "slurm_queue_timeout", x=20, y=14,
    thresholds={
        "mode": "absolute",
        "steps": [
            {"color": "green", "value": None},
            {"color": "orange", "value": 1},
        ],
    },
)
timeseries_panel(
    "Job States Over Time",
    [
        target("slurm_queue_pending", "pending", "A"),
        target("slurm_queue_pending_dependency", "pending_dependency", "B"),
        target("slurm_queue_running", "running", "C"),
        target("slurm_queue_suspended", "suspended", "D"),
        target("slurm_queue_cancelled", "cancelled", "E"),
        target("slurm_queue_completing", "completing", "F"),
        target("slurm_queue_completed", "completed", "G"),
        target("slurm_queue_configuring", "configuring", "H"),
        target("slurm_queue_failed", "failed", "I"),
        target("slurm_queue_timeout", "timeout", "J"),
        target("slurm_queue_preempted", "preempted", "K"),
        target("slurm_queue_node_fail", "node_fail", "L"),
    ],
    x=0, y=18, w=24,
)

# ---------------------------------------------------------------------------
# Scheduler
# ---------------------------------------------------------------------------
row("Scheduler", y=26)
timeseries_panel(
    "Scheduler Cycle Times",
    [
        target("slurm_scheduler_last_cycle", "last cycle", "A"),
        target("slurm_scheduler_mean_cycle", "mean cycle", "B"),
        target("slurm_scheduler_backfill_last_cycle", "backfill last cycle", "C"),
        target("slurm_scheduler_backfill_mean_cycle", "backfill mean cycle", "D"),
    ],
    x=0, y=27, unit="µs",
)
timeseries_panel(
    "Scheduler Activity",
    [
        target("slurm_scheduler_cycle_per_minute", "cycles/min", "A"),
        target("slurm_scheduler_queue_size", "queue size", "B"),
        target("slurm_scheduler_dbd_queue_size", "dbd queue size", "C"),
        target("slurm_scheduler_threads", "threads", "D"),
    ],
    x=12, y=27,
)
stat_panel("Backfilled Jobs (since start)", "slurm_scheduler_backfilled_jobs_since_start_total", x=0, y=35, w=6)
stat_panel("Backfilled Jobs (since cycle)", "slurm_scheduler_backfilled_jobs_since_cycle_total", x=6, y=35, w=6)
stat_panel("Backfilled Heterogeneous", "slurm_scheduler_backfilled_heterogeneous_total", x=12, y=35, w=6)
stat_panel("Backfill Depth (mean)", "slurm_scheduler_backfill_depth_mean", x=18, y=35, w=6)

# ---------------------------------------------------------------------------
# Partitions
# ---------------------------------------------------------------------------
row("Partitions", y=39)
table_panel(
    "Partition CPU Allocation",
    [
        ("slurm_partition_cpus_total", "{{partition}}"),
        ("slurm_partition_cpus_allocated", "{{partition}}"),
        ("slurm_partition_cpus_idle", "{{partition}}"),
        ("slurm_partition_cpus_other", "{{partition}}"),
        ("slurm_partition_jobs_pending", "{{partition}}"),
    ],
    x=0, y=40,
    rename={
        "Value #A": "CPUs Total",
        "Value #B": "CPUs Allocated",
        "Value #C": "CPUs Idle",
        "Value #D": "CPUs Other",
        "Value #E": "Jobs Pending",
        "partition": "Partition",
    },
)

# ---------------------------------------------------------------------------
# Nodes
# ---------------------------------------------------------------------------
row("Nodes", y=48)
table_panel(
    "Per-Node CPU and Memory",
    [
        ("slurm_node_cpu_total", "{{node}}"),
        ("slurm_node_cpu_alloc", "{{node}}"),
        ("slurm_node_cpu_idle", "{{node}}"),
        ("slurm_node_mem_total", "{{node}}"),
        ("slurm_node_mem_alloc", "{{node}}"),
    ],
    x=0, y=49, h=10,
    rename={
        "Value #A": "CPUs Total",
        "Value #B": "CPUs Allocated",
        "Value #C": "CPUs Idle",
        "Value #D": "Memory Total",
        "Value #E": "Memory Allocated",
        "node": "Node",
        "status": "Status",
    },
    units={"Memory Total": "decmbytes", "Memory Allocated": "decmbytes"},
)

# ---------------------------------------------------------------------------
# Accounts and fair-share
# ---------------------------------------------------------------------------
row("Accounts and Fair-Share", y=59)
table_panel(
    "Per-Account Job Stats",
    [
        ("slurm_account_jobs_running", "{{account}}"),
        ("slurm_account_jobs_pending", "{{account}}"),
        ("slurm_account_jobs_suspended", "{{account}}"),
        ("slurm_account_cpus_running", "{{account}}"),
        ("slurm_account_fairshare", "{{account}}"),
    ],
    x=0, y=60,
    rename={
        "Value #A": "Running Jobs",
        "Value #B": "Pending Jobs",
        "Value #C": "Suspended Jobs",
        "Value #D": "Running CPUs",
        "Value #E": "Fair-Share",
        "account": "Account",
    },
)

# ---------------------------------------------------------------------------
# Users
# ---------------------------------------------------------------------------
row("Users", y=68)
table_panel(
    "Per-User Job Stats",
    [
        ("slurm_user_jobs_running", "{{user}}"),
        ("slurm_user_jobs_pending", "{{user}}"),
        ("slurm_user_jobs_suspended", "{{user}}"),
        ("slurm_user_cpus_running", "{{user}}"),
    ],
    x=0, y=69,
    rename={
        "Value #A": "Running Jobs",
        "Value #B": "Pending Jobs",
        "Value #C": "Suspended Jobs",
        "Value #D": "Running CPUs",
        "user": "User",
    },
)

# ---------------------------------------------------------------------------
dashboard = {
    "__inputs": [
        {
            "name": "DS_PROMETHEUS",
            "label": "Prometheus",
            "description": "",
            "type": "datasource",
            "pluginId": "prometheus",
            "pluginName": "Prometheus",
        }
    ],
    "__requires": [
        {"type": "grafana", "id": "grafana", "name": "Grafana", "version": "12.0.0"},
        {"type": "datasource", "id": "prometheus", "name": "Prometheus", "version": "1.0.0"},
        {"type": "panel", "id": "stat", "name": "Stat", "version": ""},
        {"type": "panel", "id": "piechart", "name": "Pie chart", "version": ""},
        {"type": "panel", "id": "bargauge", "name": "Bar gauge", "version": ""},
        {"type": "panel", "id": "timeseries", "name": "Time series", "version": ""},
        {"type": "panel", "id": "table", "name": "Table", "version": ""},
        {"type": "panel", "id": "row", "name": "Row", "version": ""},
    ],
    "title": "Slurm Cluster Overview",
    "uid": "slurm-cluster-overview",
    "schemaVersion": 42,
    "version": 1,
    "editable": True,
    "graphTooltip": 1,
    "refresh": "30s",
    "time": {"from": "now-6h", "to": "now"},
    "timezone": "",
    "tags": ["slurm", "hpc"],
    "templating": {"list": []},
    "annotations": {"list": []},
    "panels": panels,
}

out_path = os.path.join(os.path.dirname(__file__), "slurm-dashboard.json")
with open(out_path, "w") as f:
    json.dump(dashboard, f, indent=2)
    f.write("\n")

print(f"wrote {out_path} ({len(panels)} panels)")
