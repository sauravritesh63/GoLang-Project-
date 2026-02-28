"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { fetchWorkflowRuns } from "@/lib/api";
import type { Status } from "@/lib/api";
import { StatusBadge } from "@/components/StatusBadge";
import { TopBar } from "@/components/TopBar";

const ALL_STATUSES: (Status | "all")[] = ["all", "pending", "running", "success", "failed"];

export default function WorkflowRunsPage() {
  const [statusFilter, setStatusFilter] = useState<Status | undefined>(undefined);

  const { data: runs, isLoading, error } = useQuery({
    queryKey: ["workflow-runs", statusFilter],
    queryFn: () => fetchWorkflowRuns(statusFilter),
    refetchInterval: 10_000,
  });

  return (
    <div className="flex flex-col h-full">
      <TopBar title="Workflow Runs" />
      <div className="flex-1 p-6">
        {/* Filter tabs */}
        <div className="mb-4 flex gap-1 rounded-xl border border-slate-200 bg-white p-1 w-fit shadow-sm">
          {ALL_STATUSES.map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s === "all" ? undefined : (s as Status))}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize transition-colors ${
                (s === "all" && !statusFilter) || s === statusFilter
                  ? "bg-blue-500 text-white shadow-sm"
                  : "text-slate-500 hover:text-slate-800"
              }`}
            >
              {s}
            </button>
          ))}
        </div>

        <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-16 text-slate-400 text-sm">
              Loading workflow runs…
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-16 text-red-500 text-sm">
              Failed to load workflow runs
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  <th className="px-4 py-3">Run ID</th>
                  <th className="px-4 py-3">Workflow ID</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Started</th>
                  <th className="px-4 py-3">Finished</th>
                  <th className="px-4 py-3">Duration</th>
                </tr>
              </thead>
              <motion.tbody initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.04 } } }}>
                {runs && runs.length > 0 ? (
                  runs.map((run) => {
                    const started = new Date(run.started_at);
                    const finished = run.finished_at ? new Date(run.finished_at) : null;
                    const durationMs = finished ? finished.getTime() - started.getTime() : null;
                    const duration = durationMs != null
                      ? durationMs < 60_000
                        ? `${(durationMs / 1000).toFixed(1)}s`
                        : `${(durationMs / 60_000).toFixed(1)}m`
                      : "—";

                    return (
                      <motion.tr
                        key={run.id}
                        variants={{ hidden: { opacity: 0, y: 8 }, visible: { opacity: 1, y: 0 } }}
                        className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                      >
                        <td className="px-4 py-3 font-mono text-xs text-slate-500">{run.id.slice(0, 12)}…</td>
                        <td className="px-4 py-3 font-mono text-xs text-slate-500">{run.workflow_id.slice(0, 12)}…</td>
                        <td className="px-4 py-3"><StatusBadge status={run.status} /></td>
                        <td className="px-4 py-3 text-xs text-slate-600">{started.toLocaleString()}</td>
                        <td className="px-4 py-3 text-xs text-slate-600">{finished ? finished.toLocaleString() : "—"}</td>
                        <td className="px-4 py-3 text-xs text-slate-600">{duration}</td>
                      </motion.tr>
                    );
                  })
                ) : (
                  <tr>
                    <td colSpan={6} className="py-16 text-center text-slate-400 text-sm">
                      No workflow runs found.
                    </td>
                  </tr>
                )}
              </motion.tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
