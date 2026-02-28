"use client";

import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import {
  GitBranch,
  Server,
  CheckCircle,
  Clock,
  TrendingUp,
} from "lucide-react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import { fetchWorkflows, fetchWorkflowRuns, fetchTaskRuns, fetchWorkers } from "@/lib/api";
import { StatCard } from "@/components/StatCard";
import { TopBar } from "@/components/TopBar";
import { useWebSocket } from "@/hooks/useWebSocket";

const STATUS_COLORS: Record<string, string> = {
  success: "#22c55e",
  failed: "#ef4444",
  running: "#3b82f6",
  pending: "#eab308",
};

export default function DashboardPage() {
  useWebSocket();

  const { data: workflows } = useQuery({
    queryKey: ["workflows"],
    queryFn: () => fetchWorkflows(0, 100),
    refetchInterval: 15_000,
  });

  const { data: workflowRuns } = useQuery({
    queryKey: ["workflow-runs"],
    queryFn: () => fetchWorkflowRuns(),
    refetchInterval: 10_000,
  });

  const { data: taskRuns } = useQuery({
    queryKey: ["task-runs"],
    queryFn: () => fetchTaskRuns(),
    refetchInterval: 10_000,
  });

  const { data: workers } = useQuery({
    queryKey: ["workers"],
    queryFn: fetchWorkers,
    refetchInterval: 10_000,
  });

  const totalWorkflows = workflows?.length ?? 0;
  const activeWorkers = workers?.filter((w) => w.status === "active").length ?? 0;
  const successRuns = workflowRuns?.filter((r) => r.status === "success").length ?? 0;
  const totalRuns = workflowRuns?.length ?? 1;
  const successRate = Math.round((successRuns / totalRuns) * 100);
  const pendingTasks = taskRuns?.filter((t) => t.status === "pending").length ?? 0;

  const statusCounts = ["success", "failed", "running", "pending"].map((s) => ({
    name: s,
    value: workflowRuns?.filter((r) => r.status === s).length ?? 0,
  })).filter((d) => d.value > 0);

  const taskBarData = ["pending", "running", "success", "failed"].map((s) => ({
    status: s,
    count: taskRuns?.filter((t) => t.status === s).length ?? 0,
  }));

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: { opacity: 1, transition: { staggerChildren: 0.08 } },
  };
  const itemVariants = {
    hidden: { opacity: 0, y: 16 },
    visible: { opacity: 1, y: 0 },
  };

  return (
    <div className="flex flex-col h-full">
      <TopBar title="Dashboard" />
      <div className="flex-1 p-6 space-y-6">
        <motion.div
          className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
        >
          <motion.div variants={itemVariants}>
            <StatCard title="Total Workflows" value={totalWorkflows} icon={GitBranch} color="blue" subtitle="Registered pipelines" />
          </motion.div>
          <motion.div variants={itemVariants}>
            <StatCard title="Active Workers" value={activeWorkers} icon={Server} color="green" subtitle={`of ${workers?.length ?? 0} total`} />
          </motion.div>
          <motion.div variants={itemVariants}>
            <StatCard title="Success Rate" value={`${successRate}%`} icon={TrendingUp} color="green" subtitle="Workflow runs" />
          </motion.div>
          <motion.div variants={itemVariants}>
            <StatCard title="Pending Tasks" value={pendingTasks} icon={Clock} color="yellow" subtitle="Awaiting execution" />
          </motion.div>
        </motion.div>

        <motion.div
          className="grid grid-cols-1 gap-6 lg:grid-cols-2"
          variants={containerVariants}
          initial="hidden"
          animate="visible"
        >
          <motion.div variants={itemVariants} className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-sm font-semibold text-slate-700 flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-blue-500" />
              Task Run Status Breakdown
            </h2>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={taskBarData} barCategoryGap="30%">
                <XAxis dataKey="status" tick={{ fontSize: 12, fill: "#64748b" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: "#64748b" }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ borderRadius: 8, border: "1px solid #e2e8f0", fontSize: 12 }} />
                <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                  {taskBarData.map((entry) => (
                    <Cell key={entry.status} fill={STATUS_COLORS[entry.status] ?? "#94a3b8"} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </motion.div>

          <motion.div variants={itemVariants} className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-sm font-semibold text-slate-700 flex items-center gap-2">
              <GitBranch className="h-4 w-4 text-blue-500" />
              Workflow Run Distribution
            </h2>
            {statusCounts.length > 0 ? (
              <div className="flex items-center gap-6">
                <ResponsiveContainer width="60%" height={200}>
                  <PieChart>
                    <Pie data={statusCounts} cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={3} dataKey="value">
                      {statusCounts.map((entry) => (
                        <Cell key={entry.name} fill={STATUS_COLORS[entry.name] ?? "#94a3b8"} />
                      ))}
                    </Pie>
                    <Tooltip contentStyle={{ borderRadius: 8, border: "1px solid #e2e8f0", fontSize: 12 }} />
                  </PieChart>
                </ResponsiveContainer>
                <ul className="space-y-2 text-sm">
                  {statusCounts.map((s) => (
                    <li key={s.name} className="flex items-center gap-2">
                      <span className="h-3 w-3 rounded-full flex-shrink-0" style={{ backgroundColor: STATUS_COLORS[s.name] ?? "#94a3b8" }} />
                      <span className="capitalize text-slate-600">{s.name}</span>
                      <span className="font-semibold text-slate-800">{s.value}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ) : (
              <p className="text-sm text-slate-400 mt-8 text-center">No run data yet</p>
            )}
          </motion.div>
        </motion.div>

        <motion.div variants={itemVariants} initial="hidden" animate="visible" className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-sm font-semibold text-slate-700">Recent Workflow Runs</h2>
          {workflowRuns && workflowRuns.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-100 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                    <th className="pb-2 pr-4">Run ID</th>
                    <th className="pb-2 pr-4">Workflow</th>
                    <th className="pb-2 pr-4">Status</th>
                    <th className="pb-2">Started</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {workflowRuns.slice(0, 5).map((run) => (
                    <tr key={run.id} className="hover:bg-slate-50 transition-colors">
                      <td className="py-2.5 pr-4 font-mono text-xs text-slate-500">{run.id.slice(0, 8)}…</td>
                      <td className="py-2.5 pr-4 text-slate-700">{run.workflow_id.slice(0, 8)}…</td>
                      <td className="py-2.5 pr-4">
                        <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold capitalize ${
                          run.status === "success" ? "bg-green-100 text-green-700"
                          : run.status === "failed" ? "bg-red-100 text-red-700"
                          : run.status === "running" ? "bg-blue-100 text-blue-700"
                          : "bg-yellow-100 text-yellow-700"
                        }`}>
                          {run.status}
                        </span>
                      </td>
                      <td className="py-2.5 text-slate-500">{new Date(run.started_at).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-sm text-slate-400">No workflow runs yet.</p>
          )}
        </motion.div>
      </div>
    </div>
  );
}
