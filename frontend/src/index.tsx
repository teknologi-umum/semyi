import { render } from "solid-js/web";
import { Router } from "solid-app-router";
import App from "@/App";
import "@/index.css";

import "@fontsource/ibm-plex-sans/500.css";
import "@fontsource/ibm-plex-sans/700.css";
import "@fontsource/libre-franklin/400.css";

render(
  () => (
    <Router>
      <App />
    </Router>
  ),
  document.getElementById("root") as HTMLElement
);
