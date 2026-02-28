import type { LucideIcon } from "lucide-react";

interface StatCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  color: "blue" | "green" | "red" | "yellow" | "slate";
  subtitle?: string;
}

const colorMap = {
  blue: { bg: "bg-blue-50", icon: "bg-blue-500", text: "text-blue-500" },
  green: { bg: "bg-green-50", icon: "bg-green-500", text: "text-green-500" },
  red: { bg: "bg-red-50", icon: "bg-red-500", text: "text-red-500" },
  yellow: { bg: "bg-yellow-50", icon: "bg-yellow-500", text: "text-yellow-500" },
  slate: { bg: "bg-slate-50", icon: "bg-slate-500", text: "text-slate-500" },
};

export function StatCard({ title, value, icon: Icon, color, subtitle }: StatCardProps) {
  const c = colorMap[color];
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-slate-500">{title}</p>
          <p className="mt-1 text-3xl font-bold text-slate-900">{value}</p>
          {subtitle && <p className="mt-1 text-xs text-slate-400">{subtitle}</p>}
        </div>
        <div className={`rounded-xl ${c.bg} p-3`}>
          <Icon className={`h-6 w-6 ${c.text}`} />
        </div>
      </div>
    </div>
  );
}
