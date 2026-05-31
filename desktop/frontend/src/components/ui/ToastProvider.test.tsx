import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { ToastProvider, useToast } from "./ToastProvider";

// Helper component that uses the toast hook
function ToastTrigger({ message = "Test toast", tone }: { message?: string; tone?: "success" | "error" | "info" }) {
  const { notify } = useToast();
  return <button onClick={() => notify(message, tone)}>Show Toast</button>;
}

describe("ToastProvider", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders children", () => {
    render(
      <ToastProvider>
        <div>Child Content</div>
      </ToastProvider>,
    );
    expect(screen.getByText("Child Content")).toBeInTheDocument();
  });

  it("shows a toast notification when notify is called", () => {
    render(
      <ToastProvider>
        <ToastTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show Toast"));

    expect(screen.getByText("Test toast")).toBeInTheDocument();
  });

  it("shows multiple toasts", () => {
    render(
      <ToastProvider>
        <ToastTrigger message="Toast Alpha" />
        <ToastTrigger message="Toast Beta" />
      </ToastProvider>,
    );

    const buttons = screen.getAllByText("Show Toast");
    fireEvent.click(buttons[0]);
    fireEvent.click(buttons[1]);

    expect(screen.getByText("Toast Alpha")).toBeInTheDocument();
    expect(screen.getByText("Toast Beta")).toBeInTheDocument();
  });

  it("removes a toast when the close button is clicked", () => {
    render(
      <ToastProvider>
        <ToastTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show Toast"));

    const toast = screen.getByText("Test toast");
    expect(toast).toBeInTheDocument();

    // The close button is rendered by the Toast component
    // Find the close button (it contains an X icon, has no aria-label)
    const closeButton = screen
      .getByText("Test toast")
      .closest("[aria-live]")
      ?.querySelector("button")!;
    expect(closeButton).not.toBeNull();

    fireEvent.click(closeButton);

    // After clicking close, the toast starts its leave animation (200ms)
    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(screen.queryByText("Test toast")).not.toBeInTheDocument();
  });

  it("auto-dismisses a toast after 3 seconds", () => {
    render(
      <ToastProvider>
        <ToastTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show Toast"));
    expect(screen.getByText("Test toast")).toBeInTheDocument();

    // Advance past the 3s auto-dismiss timer
    act(() => {
      vi.advanceTimersByTime(3000);
    });

    // After 3s the leaving animation starts (200ms)
    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(screen.queryByText("Test toast")).not.toBeInTheDocument();
  });
});
