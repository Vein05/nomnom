import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Toggle } from "./Toggle";

describe("Toggle", () => {
  it("renders the label", () => {
    render(<Toggle label="Dark Mode" />);
    expect(screen.getByText("Dark Mode")).toBeInTheDocument();
  });

  it("renders the description when provided", () => {
    render(<Toggle label="Dark Mode" description="Enable dark theme" />);
    expect(screen.getByText("Enable dark theme")).toBeInTheDocument();
  });

  it("does not render a description when not provided", () => {
    const { container } = render(<Toggle label="Dark Mode" />);
    // Only the label text should be inside the label's text container
    const textContainer = container.querySelector(".min-w-0");
    expect(textContainer?.children.length).toBe(1);
  });

  it("is unchecked by default", () => {
    render(<Toggle label="Dark Mode" />);
    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).not.toBeChecked();
  });

  it("renders checked when the checked prop is true", () => {
    render(<Toggle label="Dark Mode" checked onChange={() => {}} />);
    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).toBeChecked();
  });

  it("fires onChange when toggled", () => {
    const onChange = vi.fn();
    render(<Toggle label="Dark Mode" onChange={onChange} />);
    const checkbox = screen.getByRole("checkbox");
    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it("applies a custom className to the label", () => {
    const { container } = render(
      <Toggle label="Dark Mode" className="my-custom-class" />,
    );
    const label = container.querySelector("label");
    expect(label?.className).toContain("my-custom-class");
  });
});
