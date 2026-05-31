import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Input } from "./Input";

describe("Input", () => {
  it("renders an input element", () => {
    render(<Input />);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("renders with the given value", () => {
    render(<Input value="hello" onChange={() => {}} />);
    const input = screen.getByRole("textbox") as HTMLInputElement;
    expect(input.value).toBe("hello");
  });

  it("fires onChange when the value changes", () => {
    const onChange = vi.fn();
    render(<Input onChange={onChange} />);
    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "new value" } });
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it("renders a placeholder", () => {
    render(<Input placeholder="Enter name" />);
    const input = screen.getByRole("textbox");
    expect(input).toHaveAttribute("placeholder", "Enter name");
  });

  it("applies the mono class when mono is true", () => {
    render(<Input mono />);
    const input = screen.getByRole("textbox");
    expect(input.className).toContain("mono");
  });

  it("does not apply the mono class by default", () => {
    render(<Input />);
    const input = screen.getByRole("textbox");
    expect(input.className).not.toContain("mono");
  });

  it("applies a custom className", () => {
    render(<Input className="my-custom" />);
    const input = screen.getByRole("textbox");
    expect(input.className).toContain("my-custom");
  });

  it("forwards additional input attributes", () => {
    render(<Input disabled maxLength={10} />);
    const input = screen.getByRole("textbox");
    expect(input).toBeDisabled();
    expect(input).toHaveAttribute("maxLength", "10");
  });
});
