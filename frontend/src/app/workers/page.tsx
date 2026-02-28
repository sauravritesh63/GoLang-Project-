"use client";

import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { Server, Wifi, WifiOff } from "lucide-react";
import { fetchWorkers } from "@/lib/api";
import { StatusBadge } from "@/components/StatusBadge";
import { TopBar } from "@/components/TopBar";

function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  if (diff < 60_000) return `${Math.floor(diff / 1000)}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  return `${Math.floor(diff / 3_600_000)}h ago`;
}

export default function WorkersPage() {
  const { data: workers, isLoading, error } = useQuery({
    queryKey: ["workers"],
    queryFn: fetchWorkers,
    refetchInterval: 10_000,
  });

  const activeCount = workers?.filter((w) => w.status === "active").length ?? 0;
  const totalCount = workers?.length ?? 0;

  return (
    <div className="flex flex-col h-full">
      <TopBar title="Workers" />
      <div className="flex-1 p-6">
        {/* Summary row */}
        <div className="mb-4 flex items-center gap-4">
          <div className="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 shadow-sm">
            <Wifi className="h-4 w-4 text-green-500" />
            <span className="text-sm font-medium text-slate-700">
              {activeCount} Active
            </span>
          </div>
          <div className="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 shadow-sm">
            <WifiOff className="h-4 w-4 text-slate-400" />
            <span className="text-sm font-medium text-slate-700">
              {totalCount - activeCount} Inactive
            </span>
          </div>
          <span className="text-sm text-slate-400">{totalCount} total workers</span>
        </div>

        {/* Workers grid */}
        {isLoading ? (
          <div className="flex items-center justify-center py-16 text-slate-400 text-sm">
            Loading workers…
          </div>
        ) : error ? (
          <div className="flex items-center justify-center py-16 text-red-500 text-sm">
            Failed to load workers
          </div>
        ) : workers && workers.length > 0 ? (
          <motion.div
            className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
            initial="hidden"
            animate="visible"
            variants={{ visible: { transition: { staggerChildren: 0.06 } } }}
          >
            {workers.map((worker) => (
              <motion.div
                key={worker.id}
                variants={{ hidden: { opacity: 0, y: 12 }, visible: { opacity: 1, y: 0 } }}
                className={`rounded-xl border bg-white p-5 shadow-sm transition-shadow hover:shadow-md ${
                  worker.status === "active"
                    ? "border-green-200"
                    : "border-slate-200"
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <div
                      className={`rounded-lg p-2 ${
                        worker.status === "active" ? "bg-green-50" : "bg-slate-100"
                      }`}
                    >
                      <Server
                        className={`h-5 w-5 ${
                          worker.status === "active"
                            ? "text-green-600"
                            : "text-slate-400"
                        }`}
                      />
                    </div>
                    <div>
                      <p className="font-semibold text-slate-800 text-sm">
                        {worker.hostname}
                      </p>
                      <p className="text-xs text-slate-400 font-mono">
                        {worker.id.slice(0, 12)}…
                      </p>
                    </div>
                  </div>
                  <StatusBadge status={worker.status} />
                </div>

                <div className="mt-4 flex items-center gap-1.5 text-xs text-slate-500">
                  <span
                    className={`h-1.5 w-1.5 rounded-full ${
                      worker.status === "active" ? "bg-green-400 animate-pulse" : "bg-slate-300"
                    }`}
                  />
                  Last heartbeat: {formatRelativeTime(worker.last_heartbeat)}
                </div>
              </motion.div>
            ))}
          </motion.div>
        ) : (
          <div className="rounded-xl border border-slate-200 bg-white p-16 text-center shadow-sm">
            <Server className="mx-auto h-10 w-10 text-slate-300 mb-3" />
            <p className="text-sm text-slate-400">No workers registered.</p>
          </div>
        )}
      </div>
    </div>
  );
}
