import { rmSync } from "node:fs";
import { E2E_DIR } from "./global.setup";

export default function globalTeardown() {
  try {
    rmSync(E2E_DIR, { recursive: true, force: true });
    console.log(`[e2e:teardown] cleaned up ${E2E_DIR}`);
  } catch {
    console.warn(`[e2e:teardown] failed to clean up ${E2E_DIR}`);
  }
}
