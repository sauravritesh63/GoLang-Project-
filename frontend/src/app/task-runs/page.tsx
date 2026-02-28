"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import { X, FileText } from "lucide-react";
import { fetchTaskRuns } from "@/lib/api";
import type { Status, TaskRun } from "@/lib/api";
import { StatusBadge } from "@/components/StatusBadge";
import { TopBar } from "@/components/TopBar";

const ALL_STATUSES: (Status | "all")[] = ["all", "pending", "running", "success", "failed"];

function LogModal({ taskRun, onClose }: { taskRun: TaskRun; onClose: () => void }) {
  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 16 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95 }}
          className="w-full max-w-2xl rounded-xl bg-white shadow-xl overflow-hidden"
        >
          <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
            <div>
              <h2 className="font-semibold text-slate-900 flex items-center gap-2">
                <FileText className="h-4 w-4 text-blue-500" />
                Task Logs
              </h2>
              <p className="mt-0.5 text-xs text-slate-500">
                Task {taskRun.task_id.slice(0, 12)}… · Attempt #{taskRun.attempt}
              </p>
            </div>
            <button
              onClick={onClose}
              className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
          <div className="max-h-96 overflow-y-auto bg-slate-950 p-4">
            <pre className="whitespace-pre-wrap text-xs leading-relaxed text-green-400 font-mono">
              {taskRun.logs || "No logs available."}
            </pre>
          </div>
          <div className="flex items-center justify-between border-t border-slate-200 px-5 py-3 bg-slate-50">
            <StatusBadge status={taskRun.status} />
            <button
              onClick={onClose}
              className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-100"
            >
              Close
            </button>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}

export default function TaskRunsPage() {
  const [statusFilter, setStatusFilter] = useState<Status | undefined>(undefined);
  const [selectedTask, setSelectedTask] = useState<TaskRun | null>(null);

  const { data: taskRuns, isLoading, error } = useQuery({
    queryKey: ["task-runs", statusFilter],
    queryFn: () => fetchTaskRuns(statusFilter),
    refetchInterval: 10_000,
  });

  return (
    <div className="flex flex-col h-full">
      <TopBar title="Task Runs" />
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
              Loading task runs…
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-16 text-red-500 text-sm">
              Failed to load task runs
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  <th className="px-4 py-3">Task ID</th>
                  <th className="px-4 py-3">Run ID</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Attempt</th>
                  <th className="px-4 py-3">Started</th>
                  <th className="px-4 py-3">Finished</th>
                  <th className="px-4 py-3 text-right">Logs</th>
                </tr>
              </thead>
              <motion.tbody
                initial="hidden"
                animate="visible"
                variants={{ visible: { transition: { staggerChildren: 0.04 } } }}
              >
                {taskRuns && taskRuns.length > 0 ? (
                  taskRuns.map((task) => (
                    <motion.tr
                      key={task.id}
                      variants={{ hidden: { opacity: 0, y: 8 }, visible: { opacity: 1, y: 0 } }}
                      className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                    >
                      <td className="px-4 py-3 font-mono text-xs text-slate-500">{task.task_id.slice(0, 12)}…</td>
                      <td className="px-4 py-3 font-mono text-xs text-slate-500">{task.workflow_run_id.slice(0, 12)}…</td>
                      <td className="px-4 py-3"><StatusBadge status={task.status} /></td>
                      <td className="px-4 py-3 text-xs text-slate-600">#{task.attempt}</td>
                      <td className="px-4 py-3 text-xs text-slate-600">{new Date(task.started_at).toLocaleString()}</td>
                      <td className="px-4 py-3 text-xs text-slate-600">
                        {task.finished_at ? new Date(task.finished_at).toLocaleString() : "—"}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button
                          onClick={() => setSelectedTask(task)}
                          className="flex items-center gap-1 ml-auto rounded-lg border border-slate-200 px-2.5 py-1 text-xs font-medium text-slate-600 hover:bg-slate-100 transition-colors"
                        >
                          <FileText className="h-3 w-3" />
                          View
                        </button>
                      </td>
                    </motion.tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={7} className="py-16 text-center text-slate-400 text-sm">
                      No task runs found.
                    </td>
                  </tr>
                )}
              </motion.tbody>
            </table>
          )}
        </div>
      </div>

      {selectedTask && <LogModal taskRun={selectedTask} onClose={() => setSelectedTask(null)} />}
    </div>
  );
}
