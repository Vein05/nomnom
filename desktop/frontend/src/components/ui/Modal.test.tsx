import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Modal } from "./Modal";

describe("Modal", () => {
  it("does not render when open is false", () => {
    const { container } = render(
      <Modal open={false} title="My Modal" onClose={() => {}}>
        Content
      </Modal>,
    );
    expect(container.innerHTML).toBe("");
  });

  it("renders the title and children when open is true", () => {
    render(
      <Modal open={true} title="My Modal" onClose={() => {}}>
        <p>Modal content</p>
      </Modal>,
    );
    expect(screen.getByText("My Modal")).toBeInTheDocument();
    expect(screen.getByText("Modal content")).toBeInTheDocument();
  });

  it("fires onClose when the close button is clicked", () => {
    const onClose = vi.fn();
    render(
      <Modal open={true} title="My Modal" onClose={onClose}>
        Content
      </Modal>,
    );
    const closeBtn = screen.getByLabelText("Close modal");
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("renders a footer when provided", () => {
    render(
      <Modal
        open={true}
        title="My Modal"
        onClose={() => {}}
        footer={<button>Save</button>}
      >
        Content
      </Modal>,
    );
    expect(screen.getByText("Save")).toBeInTheDocument();
  });

  it("does not render a footer when not provided", () => {
    const { container } = render(
      <Modal open={true} title="My Modal" onClose={() => {}}>
        Content
      </Modal>,
    );
    const footerElements = container.querySelectorAll("footer");
    expect(footerElements.length).toBe(0);
  });
});
