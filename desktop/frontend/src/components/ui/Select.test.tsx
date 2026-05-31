import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Select } from "./Select";

describe("Select", () => {
  const options = (
    <>
      <option value="a">Option A</option>
      <option value="b">Option B</option>
      <option value="c">Option C</option>
    </>
  );

  it("renders the selected option label as the trigger text", () => {
    render(
      <Select value="b" onChange={() => {}}>
        {options}
      </Select>,
    );
    // The trigger button should show "Option B"
    const trigger = screen.getByRole("button");
    expect(trigger.textContent).toContain("Option B");
  });

  it("renders the first option when no value is provided", () => {
    render(<Select onChange={() => {}}>{options}</Select>);
    const trigger = screen.getByRole("button");
    expect(trigger.textContent).toContain("Option A");
  });

  it("opens the dropdown when the trigger button is clicked", () => {
    render(
      <Select value="a" onChange={() => {}}>
        {options}
      </Select>,
    );
    // Before clicking, only the trigger button exists
    expect(screen.getAllByRole("button")).toHaveLength(1);

    fireEvent.click(screen.getByRole("button"));

    // After clicking, 4 buttons: trigger + 3 option buttons
    expect(screen.getAllByRole("button")).toHaveLength(4);
  });

  it("fires onChange with the new value when an option is clicked", () => {
    const onChange = vi.fn();
    render(
      <Select value="a" onChange={onChange}>
        {options}
      </Select>,
    );

    // Open dropdown
    fireEvent.click(screen.getByRole("button"));

    // Click the option button for Option B
    const buttons = screen.getAllByRole("button");
    const optionB = buttons.find((b) => b.textContent === "Option B")!;
    fireEvent.click(optionB);

    expect(onChange).toHaveBeenCalledTimes(1);
    const event = onChange.mock.calls[0][0] as React.ChangeEvent<HTMLSelectElement>;
    expect(event.target.value).toBe("b");
  });

  it("closes the dropdown after selecting an option", () => {
    render(
      <Select value="a" onChange={() => {}}>
        {options}
      </Select>,
    );

    fireEvent.click(screen.getByRole("button"));
    // Dropdown open: 4 buttons
    expect(screen.getAllByRole("button")).toHaveLength(4);

    // Click Option C
    const buttons = screen.getAllByRole("button");
    const optionC = buttons.find((b) => b.textContent === "Option C")!;
    fireEvent.click(optionC);

    // Dropdown closed: back to 1 button
    expect(screen.getAllByRole("button")).toHaveLength(1);
  });

  it("applies the mono class to the trigger when mono is true", () => {
    render(
      <Select mono onChange={() => {}}>
        {options}
      </Select>,
    );
    const trigger = screen.getByRole("button");
    expect(trigger.className).toContain("mono");
  });

  it("does not apply the mono class by default", () => {
    render(<Select onChange={() => {}}>{options}</Select>);
    const trigger = screen.getByRole("button");
    expect(trigger.className).not.toContain("mono");
  });
});
