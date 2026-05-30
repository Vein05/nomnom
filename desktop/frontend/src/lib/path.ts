export function shortDir(dir: string): string {
  if (!dir) return dir;
  const parts = dir.split("/");
  if (parts[0] === "" && parts[1] === "Users" && parts[2]) {
    return "~/" + parts.slice(3).join("/");
  }
  return dir;
}
