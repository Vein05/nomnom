import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Button } from "./Button";

describe("Button", () => {
  it("renders children", () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });

  it("renders as a button element", () => {
    render(<Button>Submit</Button>);
    const btn = screen.getByRole("button");
    expect(btn).toBeInTheDocument();
  });

  it("fires onClick when clicked", () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Click</Button>);
    fireEvent.click(screen.getByRole("button"));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("does not fire onClick when disabled", () => {
    const onClick = vi.fn();
    render(
      <Button disabled onClick={onClick}>
        Click
      </Button>,
    );
    const btn = screen.getByRole("button");
    expect(btn).toBeDisabled();
    fireEvent.click(btn);
    expect(onClick).not.toHaveBeenCalled();
  });

  it("renders with the ghost variant by default", () => {
    render(<Button>Ghost</Button>);
    const btn = screen.getByRole("button");
    // ghost variant class includes "border-border"
    expect(btn.className).toContain("border-border");
  });

  it("renders with the solid variant class", () => {
    render(<Button variant="solid">Solid</Button>);
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("bg-accent");
    expect(btn.className).toContain("text-accent-foreground");
  });

  it("renders with the outline variant class", () => {
    render(<Button variant="outline">Outline</Button>);
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("border-accent");
    expect(btn.className).toContain("text-accent");
  });

  it("renders with the danger variant class", () => {
    render(<Button variant="danger">Danger</Button>);
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("text-danger");
    expect(btn.className).toContain("border-danger");
  });

  it("applies a custom className", () => {
    render(<Button className="extra-class">Styled</Button>);
    const btn = screen.getByRole("button");
    expect(btn.className).toContain("extra-class");
  });
});
