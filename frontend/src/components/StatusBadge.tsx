import type { Status, WorkerStatus } from "@/lib/api";

const statusStyles: Record<Status | WorkerStatus, string> = {
  pending: "bg-yellow-100 text-yellow-800 border border-yellow-300",
  running: "bg-blue-100 text-blue-800 border border-blue-300",
  success: "bg-green-100 text-green-800 border border-green-300",
  failed: "bg-red-100 text-red-800 border border-red-300",
  active: "bg-green-100 text-green-800 border border-green-300",
  inactive: "bg-slate-100 text-slate-600 border border-slate-300",
};

interface StatusBadgeProps {
  status: Status | WorkerStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${statusStyles[status]}`}
    >
      <span
        className={`h-1.5 w-1.5 rounded-full ${
          status === "running"
            ? "animate-pulse bg-blue-500"
            : status === "success" || status === "active"
            ? "bg-green-500"
            : status === "failed"
            ? "bg-red-500"
            : "bg-yellow-500"
        }`}
      />
      {status}
    </span>
  );
}
