import { chromium } from '@playwright/test';

// Launches a browser page that records every outgoing request URL, so
// scenarios can assert "no network calls" (the ADR's hard requirement for
// Mermaid/KaTeX hydration).
export async function withPage(fn) {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  const requests = [];
  page.on('request', (req) => requests.push(req.url()));
  try {
    return await fn(page, requests);
  } finally {
    await browser.close();
  }
}
