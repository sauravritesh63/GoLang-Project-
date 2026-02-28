"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchHealth } from "@/lib/api";
import { Activity, Wifi, WifiOff } from "lucide-react";

interface TopBarProps {
  title: string;
}

export function TopBar({ title }: TopBarProps) {
  const { data, isError } = useQuery({
    queryKey: ["health"],
    queryFn: fetchHealth,
    refetchInterval: 15_000,
    retry: false,
  });

  const isHealthy = !isError && data?.status === "ok";

  return (
    <header className="flex h-16 items-center justify-between border-b border-slate-200 bg-white px-6 shadow-sm">
      <div className="flex items-center gap-2">
        <Activity className="h-5 w-5 text-slate-400" />
        <h1 className="text-lg font-semibold text-slate-800">{title}</h1>
      </div>
      <div className="flex items-center gap-2">
        {isHealthy ? (
          <Wifi className="h-4 w-4 text-green-500" />
        ) : (
          <WifiOff className="h-4 w-4 text-red-500" />
        )}
        <span
          className={`text-xs font-medium ${
            isHealthy ? "text-green-600" : "text-red-600"
          }`}
        >
          {isHealthy ? "System Healthy" : "API Unreachable"}
        </span>
        <span
          className={`ml-1 h-2 w-2 rounded-full ${
            isHealthy ? "bg-green-500 animate-pulse" : "bg-red-500"
          }`}
        />
      </div>
    </header>
  );
}
