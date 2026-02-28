"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { Plus, Play, GitBranch, RefreshCw } from "lucide-react";
import { fetchWorkflows, createWorkflow, triggerWorkflow } from "@/lib/api";
import type { Workflow } from "@/lib/api";
import { TopBar } from "@/components/TopBar";

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  visible: { opacity: 1, y: 0 },
};

function CreateWorkflowModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const [form, setForm] = useState({
    name: "",
    description: "",
    schedule_cron: "",
    is_active: true,
  });

  const mutation = useMutation({
    mutationFn: createWorkflow,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["workflows"] });
      onClose();
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl"
      >
        <h2 className="mb-4 text-lg font-semibold text-slate-900">Create Workflow</h2>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Name</label>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="My Workflow"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Description</label>
            <textarea
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              rows={2}
              placeholder="Optional description"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700">Cron Schedule</label>
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              value={form.schedule_cron}
              onChange={(e) => setForm({ ...form, schedule_cron: e.target.value })}
              placeholder="0 * * * *"
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="is_active"
              checked={form.is_active}
              onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
              className="rounded border-slate-300"
            />
            <label htmlFor="is_active" className="text-sm text-slate-700">Active</label>
          </div>
        </div>
        {mutation.error && (
          <p className="mt-3 text-sm text-red-600">{(mutation.error as Error).message}</p>
        )}
        <div className="mt-5 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50"
          >
            Cancel
          </button>
          <button
            onClick={() => mutation.mutate(form)}
            disabled={mutation.isPending || !form.name}
            className="flex items-center gap-1.5 rounded-lg bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600 disabled:opacity-60"
          >
            {mutation.isPending ? <RefreshCw className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            Create
          </button>
        </div>
      </motion.div>
    </div>
  );
}

function WorkflowRow({ workflow }: { workflow: Workflow }) {
  const qc = useQueryClient();
  const [triggered, setTriggered] = useState(false);

  const trigger = useMutation({
    mutationFn: () => triggerWorkflow(workflow.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["workflow-runs"] });
      setTriggered(true);
      setTimeout(() => setTriggered(false), 2000);
    },
  });

  return (
    <motion.tr
      variants={itemVariants}
      className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
    >
      <td className="px-4 py-3">
        <div className="flex items-center gap-2">
          <GitBranch className="h-4 w-4 text-blue-400 flex-shrink-0" />
          <span className="font-medium text-slate-800">{workflow.name}</span>
        </div>
        {workflow.description && (
          <p className="ml-6 mt-0.5 text-xs text-slate-500">{workflow.description}</p>
        )}
      </td>
      <td className="px-4 py-3 font-mono text-xs text-slate-600">{workflow.schedule_cron || "—"}</td>
      <td className="px-4 py-3">
        <span
          className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold ${
            workflow.is_active
              ? "bg-green-100 text-green-700"
              : "bg-slate-100 text-slate-500"
          }`}
        >
          {workflow.is_active ? "Active" : "Inactive"}
        </span>
      </td>
      <td className="px-4 py-3 text-xs text-slate-500">
        {new Date(workflow.created_at).toLocaleDateString()}
      </td>
      <td className="px-4 py-3 text-right">
        <button
          onClick={() => trigger.mutate()}
          disabled={trigger.isPending}
          className={`flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
            triggered
              ? "bg-green-100 text-green-700"
              : "bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-60"
          }`}
        >
          {trigger.isPending ? (
            <RefreshCw className="h-3 w-3 animate-spin" />
          ) : (
            <Play className="h-3 w-3" />
          )}
          {triggered ? "Triggered!" : "Trigger"}
        </button>
      </td>
    </motion.tr>
  );
}

export default function WorkflowsPage() {
  const [showModal, setShowModal] = useState(false);

  const { data: workflows, isLoading, error } = useQuery({
    queryKey: ["workflows"],
    queryFn: () => fetchWorkflows(0, 100),
    refetchInterval: 15_000,
  });

  return (
    <div className="flex flex-col h-full">
      <TopBar title="Workflows" />
      <div className="flex-1 p-6">
        <div className="mb-4 flex items-center justify-between">
          <p className="text-sm text-slate-500">
            {workflows?.length ?? 0} workflow{(workflows?.length ?? 0) !== 1 ? "s" : ""} registered
          </p>
          <button
            onClick={() => setShowModal(true)}
            className="flex items-center gap-2 rounded-lg bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600 transition-colors"
          >
            <Plus className="h-4 w-4" />
            New Workflow
          </button>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-16 text-slate-400 text-sm">
              Loading workflows…
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-16 text-red-500 text-sm">
              Failed to load workflows
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Schedule</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Created</th>
                  <th className="px-4 py-3 text-right">Action</th>
                </tr>
              </thead>
              <motion.tbody initial="hidden" animate="visible" variants={{ visible: { transition: { staggerChildren: 0.05 } } }}>
                {workflows && workflows.length > 0 ? (
                  workflows.map((wf) => <WorkflowRow key={wf.id} workflow={wf} />)
                ) : (
                  <tr>
                    <td colSpan={5} className="py-16 text-center text-slate-400 text-sm">
                      No workflows yet. Create one to get started.
                    </td>
                  </tr>
                )}
              </motion.tbody>
            </table>
          )}
        </div>
      </div>

      {showModal && <CreateWorkflowModal onClose={() => setShowModal(false)} />}
    </div>
  );
}
