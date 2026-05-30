import type {
  AIStatus,
  AnalyticsSummary,
  DesktopConfig,
  JobStatus,
  OpenRouterKeyStatus,
  OpenRouterTestResult,
  RenameEntry,
  RunJobOptions,
  Session,
} from "./types";

type Backend = {
  SelectFolder: (defaultDirectory: string) => Promise<string>;
  SelectConfigFile: (defaultPath: string) => Promise<string>;
  CreateConfigFile: (defaultPath: string) => Promise<string>;
  GetConfigPath: () => Promise<string>;
  GetAIStatus: () => Promise<AIStatus>;
  CheckOpenRouterAPIKey: () => Promise<OpenRouterKeyStatus>;
  TestOpenRouterAPIKey: (apiKey: string, model: string) => Promise<OpenRouterTestResult>;
  SetConfigPath: (path: string) => Promise<DesktopConfig>;
  ScanDirectory: (path: string) => Promise<string>;
  GenerateNames: (jobID: string) => Promise<void>;
  GetPlan: (jobID: string) => Promise<RenameEntry[]>;
  RunJob: (jobID: string, opts: RunJobOptions) => Promise<string>;
  CancelJob: (jobID: string) => Promise<boolean>;
  GetJobStatus: (jobID: string) => Promise<JobStatus>;
  GetHistory: () => Promise<Session[]>;
  GetAnalytics: () => Promise<AnalyticsSummary>;
  GetConfig: () => Promise<DesktopConfig>;
  SaveConfig: (config: DesktopConfig) => Promise<boolean>;
  OpenFile: (path: string) => Promise<void>;
};

function backend(): Backend {
  const app = (window as any)?.go?.main?.App as Backend | undefined;
  if (!app) {
    throw new Error("Wails backend bridge is unavailable. Start with 'wails dev'.");
  }
  return app;
}

export const wails = {
  selectFolder(defaultDirectory = "") {
    return backend().SelectFolder(defaultDirectory);
  },
  selectConfigFile(defaultPath = "") {
    return backend().SelectConfigFile(defaultPath);
  },
  createConfigFile(defaultPath = "") {
    return backend().CreateConfigFile(defaultPath);
  },
  getConfigPath() {
    return backend().GetConfigPath();
  },
  getAIStatus() {
    return backend().GetAIStatus();
  },
  checkOpenRouterAPIKey() {
    return backend().CheckOpenRouterAPIKey();
  },
  testOpenRouterAPIKey(apiKey = "", model = "") {
    return backend().TestOpenRouterAPIKey(apiKey, model);
  },
  setConfigPath(path: string) {
    return backend().SetConfigPath(path);
  },
  scanDirectory(path: string) {
    return backend().ScanDirectory(path);
  },
  generateNames(jobID: string) {
    return backend().GenerateNames(jobID);
  },
  getPlan(jobID: string) {
    return backend().GetPlan(jobID);
  },
  runJob(jobID: string, opts: RunJobOptions) {
    return backend().RunJob(jobID, opts);
  },
  cancelJob(jobID: string) {
    return backend().CancelJob(jobID);
  },
  getJobStatus(jobID: string) {
    return backend().GetJobStatus(jobID);
  },
  getHistory() {
    return backend().GetHistory();
  },
  getAnalytics() {
    return backend().GetAnalytics();
  },
  getConfig() {
    return backend().GetConfig();
  },
  saveConfig(config: DesktopConfig) {
    return backend().SaveConfig(config);
  },
  openFile(path: string) {
    return backend().OpenFile(path);
  },
};
