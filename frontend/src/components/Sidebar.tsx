"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion } from "framer-motion";
import {
  LayoutDashboard,
  GitBranch,
  History,
  ListTodo,
  Server,
  Zap,
} from "lucide-react";

const navItems = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/workflows", label: "Workflows", icon: GitBranch },
  { href: "/workflow-runs", label: "Workflow Runs", icon: History },
  { href: "/task-runs", label: "Task Runs", icon: ListTodo },
  { href: "/workers", label: "Workers", icon: Server },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex h-screen w-60 flex-col bg-slate-900 text-white">
      {/* Logo */}
      <div className="flex items-center gap-2 px-6 py-5 border-b border-slate-700">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500">
          <Zap className="h-4 w-4 text-white" />
        </div>
        <span className="text-base font-bold tracking-tight">TaskScheduler</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-1">
        {navItems.map(({ href, label, icon: Icon }) => {
          const isActive = pathname === href;
          return (
            <Link key={href} href={href} className="relative block">
              {isActive && (
                <motion.div
                  layoutId="sidebar-active"
                  className="absolute inset-0 rounded-lg bg-blue-500/20"
                  transition={{ type: "spring", bounce: 0.2, duration: 0.4 }}
                />
              )}
              <span
                className={`relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? "text-blue-400"
                    : "text-slate-400 hover:text-white hover:bg-slate-800"
                }`}
              >
                <Icon className="h-4 w-4 flex-shrink-0" />
                {label}
              </span>
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-slate-700 px-6 py-4">
        <p className="text-xs text-slate-500">Task Scheduler v1.0</p>
      </div>
    </aside>
  );
}
