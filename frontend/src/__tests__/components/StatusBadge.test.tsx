import { render, screen } from "@testing-library/react";
import { StatusBadge } from "@/components/StatusBadge";
import type { Status, WorkerStatus } from "@/lib/api";

describe("StatusBadge", () => {
  it.each<[Status | WorkerStatus, string]>([
    ["pending", "pending"],
    ["running", "running"],
    ["success", "success"],
    ["failed", "failed"],
    ["active", "active"],
    ["inactive", "inactive"],
  ])("renders %s status label", (status, label) => {
    render(<StatusBadge status={status} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it("applies correct classes for pending status", () => {
    const { container } = render(<StatusBadge status="pending" />);
    const span = container.firstChild as HTMLElement;
    expect(span).toHaveClass("bg-yellow-100", "text-yellow-800");
  });

  it("applies correct classes for running status", () => {
    const { container } = render(<StatusBadge status="running" />);
    const span = container.firstChild as HTMLElement;
    expect(span).toHaveClass("bg-blue-100", "text-blue-800");
  });

  it("applies correct classes for success status", () => {
    const { container } = render(<StatusBadge status="success" />);
    const span = container.firstChild as HTMLElement;
    expect(span).toHaveClass("bg-green-100", "text-green-800");
  });

  it("applies correct classes for failed status", () => {
    const { container } = render(<StatusBadge status="failed" />);
    const span = container.firstChild as HTMLElement;
    expect(span).toHaveClass("bg-red-100", "text-red-800");
  });

  it("applies animate-pulse class for running indicator dot", () => {
    const { container } = render(<StatusBadge status="running" />);
    const dot = container.querySelector(".animate-pulse");
    expect(dot).toBeInTheDocument();
  });

  it("does not apply animate-pulse for non-running statuses", () => {
    const { container } = render(<StatusBadge status="success" />);
    expect(container.querySelector(".animate-pulse")).not.toBeInTheDocument();
  });
});
