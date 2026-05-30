// Dashboard entry point
import { initCharts, fetchMetrics } from './lib/charts';

const INTERVAL_MS = 5000;

async function main() {
  const data = await fetchMetrics('/api/v2/metrics');
  initCharts('#dashboard', data);
  setInterval(() => refresh(), INTERVAL_MS);
}

document.addEventListener('DOMContentLoaded', main);