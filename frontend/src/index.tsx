import * as Sentry from "@sentry/solid";
import { solidRouterBrowserTracingIntegration, withSentryRouterRouting } from "@sentry/solid/solidrouter";
import { render } from "solid-js/web";
import "@/index.css";

import "@fontsource/ibm-plex-sans/500.css";
import "@fontsource/ibm-plex-sans/700.css";
import "@fontsource/libre-franklin/400.css";
import DetailPage from "@/pages/Detail";
import OverviewPage from "@/pages/Overview";
import { Route, Router } from "@solidjs/router";

Sentry.init({
  dsn: import.meta.env.VITE_FRONTEND_SENTRY_DSN,
  integrations: [solidRouterBrowserTracingIntegration(), Sentry.replayIntegration()],
  sampleRate: Number.parseFloat(import.meta.env.VITE_FRONTEND_SENTRY_SAMPLE_RATE ?? 1.0),
  // Set tracesSampleRate to 1.0 to capture 100% of transactions for tracing.
  // We recommend adjusting this value in production
  // Learn more at
  // https://docs.sentry.io/platforms/javascript/configuration/options/#traces-sample-rate
  tracesSampleRate: Number.parseFloat(import.meta.env.VITE_FRONTEND_SENTRY_TRACES_SAMPLE_RATE ?? 1.0),
  // Set `tracePropagationTargets` to control for which URLs trace propagation should be enabled
  tracePropagationTargets: ["localhost", window.location.origin],
  // Capture Replay for 10% of all sessions, plus 100% of sessions with an error
  // Learn more at
  // https://docs.sentry.io/platforms/javascript/session-replay/configuration/#general-integration-configuration
  replaysSessionSampleRate: Number.parseFloat(import.meta.env.VITE_FRONTEND_SENTRY_REPLAYS_SESSION_SAMPLE_RATE ?? 0.05),
  replaysOnErrorSampleRate: Number.parseFloat(import.meta.env.VITE_FRONTEND_SENTRY_REPLAYS_ON_ERROR_SAMPLE_RATE ?? 0.5),
});

// Wrap Solid Router to collect meaningful performance data on route changes
const SentryRouter = withSentryRouterRouting(Router);

render(
  () => (
    <SentryRouter>
      <Route path="/" component={OverviewPage} />
      <Route path="/by" component={DetailPage} />
    </SentryRouter>
  ),
  document.getElementById("root") as HTMLElement,
);
