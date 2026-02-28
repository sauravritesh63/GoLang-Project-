import { render, screen } from "@testing-library/react";
import { GitBranch, Server } from "lucide-react";
import { StatCard } from "@/components/StatCard";

describe("StatCard", () => {
  it("renders title and value", () => {
    render(
      <StatCard title="Total Workflows" value={42} icon={GitBranch} color="blue" />
    );
    expect(screen.getByText("Total Workflows")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
  });

  it("renders optional subtitle when provided", () => {
    render(
      <StatCard
        title="Workers"
        value={3}
        icon={Server}
        color="green"
        subtitle="Active nodes"
      />
    );
    expect(screen.getByText("Active nodes")).toBeInTheDocument();
  });

  it("does not render subtitle element when omitted", () => {
    render(
      <StatCard title="Workers" value={3} icon={Server} color="green" />
    );
    expect(screen.queryByText("Active nodes")).not.toBeInTheDocument();
  });

  it("renders a string value", () => {
    render(
      <StatCard title="Status" value="online" icon={Server} color="slate" />
    );
    expect(screen.getByText("online")).toBeInTheDocument();
  });

  it.each([
    ["blue", "bg-blue-50"],
    ["green", "bg-green-50"],
    ["red", "bg-red-50"],
    ["yellow", "bg-yellow-50"],
    ["slate", "bg-slate-50"],
  ] as const)("applies %s color class", (color, expectedClass) => {
    const { container } = render(
      <StatCard title="T" value={0} icon={Server} color={color} />
    );
    expect(container.querySelector(`.${expectedClass}`)).toBeInTheDocument();
  });
});
